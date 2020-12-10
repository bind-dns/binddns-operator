package operator

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	appslisters "k8s.io/client-go/listers/apps/v1"

	kubeinformers "k8s.io/client-go/informers"

	"github.com/bind-dns/binddns-operator/pkg/kube"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsOperator struct {
	// k8sClientSet is a standard kubernetes clientset
	k8sClientSet kubernetes.Interface

	KubeInformerFactory kubeinformers.SharedInformerFactory
	deploymentsLister   appslisters.DeploymentLister
	deploymentsSynced   cache.InformerSynced

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewDnsOperator() (operator *DnsOperator, err error) {
	kubeClient, err := kube.GetKubeClient()
	if err != nil {
		return nil, err
	}

	zlog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.GetClientSet().CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: AppName})

	// Init k8s standard informer.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient.GetClientSet(), 30*time.Second)
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()

	// Init k8s crd informer

	dnsOperator := &DnsOperator{
		k8sClientSet:        kubeClient.GetClientSet(),
		KubeInformerFactory: kubeInformerFactory,
		deploymentsLister:   deploymentInformer.Lister(),
		deploymentsSynced:   deploymentInformer.Informer().HasSynced,
		recorder:            recorder,
	}

	// Set up an event handler for when Deployment resources change.
	deploymentInformer.Informer().AddEventHandler(&DeploymentEventHandler{})
	return dnsOperator, nil
}

func (operator *DnsOperator) Run() error {
	return nil
}
