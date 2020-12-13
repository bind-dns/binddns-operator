package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	dnsclientset "github.com/bind-dns/binddns-operator/pkg/generated/clientset/versioned"
)

var (
	globalKubeClient *KubeClient
)

type KubeClient struct {
	restConfig *rest.Config

	clientSet       kubernetes.Interface
	dnsClientSet dnsclientset.Interface
}

func InitKubernetesClient() error {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// k8s standard clientset
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// crd clientset
	dnsClientSet, err := dnsclientset.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	globalKubeClient = &KubeClient{
		restConfig:      restConfig,
		clientSet:       clientSet,
		dnsClientSet: dnsClientSet,
	}
	return nil
}

func GetKubeClient() *KubeClient {
	return globalKubeClient
}

func (client *KubeClient) GetRestConfig() *rest.Config {
	return client.restConfig
}

func (client *KubeClient) GetClientSet() kubernetes.Interface {
	return client.clientSet
}

func (client *KubeClient) GetDnsClientSet() dnsclientset.Interface {
	return client.dnsClientSet
}
