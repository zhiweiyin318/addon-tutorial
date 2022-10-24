package main

import (
	"context"
	"embed"
	"os"

	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/utils"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

//go:embed manifests
var FS embed.FS

const (
	addonName = "large-addon"
)

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

	healthProber := utils.NewDeploymentProber(types.NamespacedName{Name: "large-addon-agent", Namespace: "open-cluster-management-agent-addon"})
	agentAddon, err := addonfactory.NewAgentAddonFactory(addonName, FS, "manifests").
		WithAgentHealthProber(healthProber).WithGetValuesFuncs(getDefaultValues).
		BuildTemplateAgentAddon()
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

func getDefaultValues(cluster *clusterv1.ManagedCluster,
	addon *addonapiv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	version := os.Getenv("ADDON_VERSION")

	manifestConfig := struct {
		AddonVersion string
	}{
		AddonVersion: version,
	}

	return addonfactory.StructToValues(manifestConfig), nil
}
