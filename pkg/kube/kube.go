package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	globalKubeClient *KubeClient
)

type KubeClient struct {
	restConfig *rest.Config
	clientSet  kubernetes.Interface
}

// GetKubeClient get the singleton object.
func GetKubeClient() (client *KubeClient, err error) {
	if globalKubeClient != nil {
		return globalKubeClient, nil
	}

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	globalKubeClient = &KubeClient{
		restConfig: restConfig,
		clientSet:  clientSet,
	}
	return globalKubeClient, nil
}

func (client *KubeClient) GetRestConfig() *rest.Config {
	return client.restConfig
}

func (client *KubeClient) GetClientSet() kubernetes.Interface {
	return client.clientSet
}
