package controller

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"github.com/bind-dns/binddns-operator/pkg/controller/queue"
	"github.com/bind-dns/binddns-operator/pkg/controller/router"
	dnsinformers "github.com/bind-dns/binddns-operator/pkg/generated/informers/externalversions"
	dnslister "github.com/bind-dns/binddns-operator/pkg/generated/listers/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/utils"

	"github.com/bind-dns/binddns-operator/pkg/kube"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsController struct {
	// k8sClientSet is a standard kubernetes clientset
	k8sClientSet kubernetes.Interface

	// DNS informers.
	DnsInformerFactory dnsinformers.SharedInformerFactory

	// DnsDomain
	domainLister dnslister.DnsDomainLister
	domainSynced cache.InformerSynced

	// DnsRule
	ruleLister dnslister.DnsRuleLister
	ruleSynced cache.InformerSynced

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// dnsSynced define the dns already init at the beginning
	dnsSynced bool

	// httpServer http api object.
	httpServer *router.HttpServer

	workQueues []*queue.DnsQueue
}

func NewDnsController(workThreads int32) (controller *DnsController, err error) {
	kubeClient := kube.GetKubeClient()

	zlog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.GetClientSet().CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: AppName})

	// init dns informer factory.
	dnsInformerFactory := dnsinformers.NewSharedInformerFactory(kubeClient.GetDnsClientSet(), 30*time.Second)
	dnsDomainInformer := dnsInformerFactory.Binddns().V1().DnsDomains()
	dnsRuleInformer := dnsInformerFactory.Binddns().V1().DnsRules()

	controller = &DnsController{
		k8sClientSet:       kubeClient.GetClientSet(),
		DnsInformerFactory: dnsInformerFactory,
		domainLister:       dnsDomainInformer.Lister(),
		domainSynced:       dnsDomainInformer.Informer().HasSynced,
		ruleLister:         dnsRuleInformer.Lister(),
		ruleSynced:         dnsRuleInformer.Informer().HasSynced,
		recorder:           recorder,
	}
	// init work queues.
	controller.generateWorkQueue(workThreads)

	// add crd event handler.
	dnsDomainInformer.Informer().AddEventHandlerWithResyncPeriod(&dnsDomainEventHandler{dnsController: controller}, 10*time.Minute)
	dnsRuleInformer.Informer().AddEventHandlerWithResyncPeriod(&dnsRuleEventHandler{dnsController: controller}, 10*time.Minute)
	return controller, nil
}

func (controller *DnsController) RegisterHttpRouter(port string) {
	controller.httpServer = router.NewHttpServer(port)
}

func (controller *DnsController) generateWorkQueue(workThreads int32) {
	for i := 0; i < int(workThreads); i++ {
		controller.workQueues = append(controller.workQueues, queue.NewDnsQueue("worker-"+strconv.Itoa(i+1)))
	}
}

// enqueue according to the hash of domain, assign message to event work queue.
// If informer is syncing, dnsSynced control not send message.
func (controller *DnsController) enqueue(event *queue.DnsEvent) bool {
	if controller.dnsSynced {
		mod := utils.CalcAscII(event.Domain) % len(controller.workQueues)
		controller.workQueues[mod].Enqueue(event)
		return true
	}
	return false
}

func (controller *DnsController) Run(stopCh <-chan struct{}) error {
	zlog.Infof("Waiting for informer caches to sync.")
	if ok := cache.WaitForCacheSync(stopCh, controller.domainSynced, controller.ruleSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// After informer synced, set dnsSynced true.
	controller.dnsSynced = true

	queueLength := len(controller.workQueues)
	wg := &sync.WaitGroup{}
	wg.Add(queueLength)

	// Run work queues.
	for i := 0; i < queueLength; i++ {
		go func(i int) {
			controller.workQueues[i].Run()
			wg.Done()
		}(i)
	}

	// Start http server.
	if controller.httpServer != nil {
		wg.Add(1)
		go func() {
			controller.httpServer.Start()
			wg.Done()
		}()
	}

	// Received stop channel, and stop the work queues.
	go func() {
		<-stopCh
		zlog.Infof("Stopping all the work queues.")
		for i := 0; i < queueLength; i++ {
			controller.workQueues[i].Stop()
		}
		controller.httpServer.Stop()
	}()
	wg.Wait()

	zlog.Infof("Controller is stopped successfully.")
	return nil
}
