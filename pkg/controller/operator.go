package controller

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	dnsinformers "github.com/bind-dns/binddns-operator/pkg/generated/informers/externalversions"
	dnslister "github.com/bind-dns/binddns-operator/pkg/generated/listers/binddns/v1"

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
}

func NewDnsController() (operator *DnsController, err error) {
	kubeClient := kube.GetKubeClient()

	zlog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.GetClientSet().CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: AppName})

	// Init dns informer factory.
	dnsInformerFactory := dnsinformers.NewSharedInformerFactory(kubeClient.GetDnsClientSet(), 30 *time.Second)

	dnsOperator := &DnsController{
		k8sClientSet:        kubeClient.GetClientSet(),
		DnsInformerFactory:  dnsInformerFactory,
		domainLister: dnsInformerFactory.Binddns().V1().DnsDomains().Lister(),
		domainSynced: dnsInformerFactory.Binddns().V1().DnsDomains().Informer().HasSynced,
		ruleLister: dnsInformerFactory.Binddns().V1().DnsRules().Lister(),
		ruleSynced: dnsInformerFactory.Binddns().V1().DnsRules().Informer().HasSynced,
		recorder:            recorder,
	}
	return dnsOperator, nil
}

func (operator *DnsController) Run() error {

	return nil
}
