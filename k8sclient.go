package main

import (
	"context"
	"errors"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// getRestConfig is a helper to load the config from a kubeconfig file
func getRestConfig() (*rest.Config, error) {
	if sOpts.KubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", sOpts.KubeConfigPath)
	}
	return rest.InClusterConfig()
}

// NewK8SClient loads a new k8s client for integration with the configured cluster
func NewK8SClient() (*kubernetes.Clientset, error) {
	config, err := getRestConfig()

	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("Failed to load kubernetes config")
	}

	return kubernetes.NewForConfig(config)
}

// GetNamespaces retrieves a list of namespaces from the cluster as a slice of strings
func GetNamespaces(client *kubernetes.Clientset) ([]string, error) {
	k8sNamespaces, err := client.CoreV1().Namespaces().List(context.Background(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, 0)
	for _, k8sNamespace := range k8sNamespaces.Items {
		namespaces = append(namespaces, k8sNamespace.Name)
	}
	return namespaces, nil
}
