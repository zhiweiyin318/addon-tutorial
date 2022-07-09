package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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

	go leaseUpdater.Start(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
