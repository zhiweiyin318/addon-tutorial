package main

import (
	"context"
	"embed"

	"os"

	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

//go:embed manifests
//go:embed manifests/leaseprober
//go:embed manifests/leaseprober/templates

var FS embed.FS

const (
	addonName             = "helm-addon"
	addonInstallNamespace = "open-cluster-management-agent-addon"
)

type Values struct {
	Name        string `json:"name"`
	InstallName string `json:"installName"`
	Image       string `json:"image"`
}

func getValues(cluster *clusterv1.ManagedCluster,
	addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	userJsonValues := Values{
		Name:        "helm-addon-agent",
		InstallName: addon.Spec.InstallNamespace,
		Image:       "quay.io/open-cluster-management/addons:latest",
	}

	return addonfactory.JsonStructToValues(userJsonValues)
}

func main() {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		os.Exit(1)
	}
	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		klog.Errorf("unable to setup addon manager: %v", err)
		os.Exit(1)
	}

	agentAddon, err := addonfactory.NewAgentAddonFactory(addonName, FS, "manifests/leaseprober").
		WithGetValuesFuncs(getValues, addonfactory.GetValuesFromAddonAnnotation).
		WithInstallStrategy(agent.InstallAllStrategy(addonInstallNamespace)).
		BuildHelmAgentAddon()
	if err != nil {
		klog.Errorf("failed to build agent addon %v", err)
		os.Exit(1)
	}

	err = addonMgr.AddAgent(agentAddon)
	if err != nil {
		klog.Errorf("failed to add addon agent: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	go addonMgr.Start(ctx)

	<-ctx.Done()
}
