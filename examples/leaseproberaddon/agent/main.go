package main

import (
	"context"
	"os"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/lease"
)

const (
	addonName             = "leaseprober-addon"
	addonInstallNamespace = "open-cluster-management-agent-addon"
)

func main() {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		os.Exit(1)
	}
	spokeKubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		os.Exit(1)
	}

	// create a lease updater
	leaseUpdater := lease.NewLeaseUpdater(
		spokeKubeClient,
		addonName,
		addonInstallNamespace,
	)

	ctx := context.Background()
	go leaseUpdater.Start(context.Background())

	<-ctx.Done()
}
