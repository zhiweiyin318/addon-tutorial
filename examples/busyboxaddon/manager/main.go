package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

type busyboxAddonAgent struct{}

func (a *busyboxAddonAgent) Manifests(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) ([]runtime.Object, error) {
	objects := []runtime.Object{
		&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "busybox",
				Namespace: addon.Spec.InstallNamespace,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "busybox",
						Image: "busybox",
						Command: []string{
							"sleep",
							"3600",
						},
						ImagePullPolicy: v1.PullIfNotPresent,
					},
				},
			},
		},
	}

	return objects, nil
}

func (a *busyboxAddonAgent) GetAgentAddonOptions() agent.AgentAddonOptions {
	return agent.AgentAddonOptions{
		AddonName: "busybox-addon",
		// can configure InstallStrategy to create managedClusterAddon automatically.
		// currently we support InstallAllStrategy and InstallByLabelStrategy strategies.
		// InstallStrategy: agent.InstallAllStrategy("open-cluster-management-agent-addon"),
	}
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

	err = addonMgr.AddAgent(&busyboxAddonAgent{})
	if err != nil {
		klog.Errorf("unable to add addon agent: %v", err)
		os.Exit(1)
	}

	go addonMgr.Start(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
