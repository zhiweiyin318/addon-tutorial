package main

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	workapiv1 "open-cluster-management.io/api/work/v1"
)

const (
	addonName             = "workprober-addon"
	addonInstallNamespace = "open-cluster-management-agent-addon"
)

type leaseProberAddonAgent struct{}

func (a *leaseProberAddonAgent) Manifests(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) ([]runtime.Object, error) {
	objects := []runtime.Object{
		&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workprober-addon-agent",
				Namespace: addon.Spec.InstallNamespace,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "workprober-addon-agent",
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

func (a *leaseProberAddonAgent) GetAgentAddonOptions() agent.AgentAddonOptions {
	return agent.AgentAddonOptions{
		AddonName:       addonName,
		InstallStrategy: agent.InstallAllStrategy(addonInstallNamespace),
		HealthProber: &agent.HealthProber{
			Type: agent.HealthProberTypeWork,
			WorkProber: &agent.WorkHealthProber{
				ProbeFields: []agent.ProbeField{
					{
						ResourceIdentifier: workapiv1.ResourceIdentifier{
							Group:     "",
							Resource:  "pods",
							Name:      "workprober",
							Namespace: addonInstallNamespace,
						},
						ProbeRules: []workapiv1.FeedbackRule{
							{
								Type: workapiv1.WellKnownStatusType,
							},
						},
					},
				},
				HealthCheck: func(identifier workapiv1.ResourceIdentifier, result workapiv1.StatusFeedbackResult) error {
					switch identifier.Resource {
					case "pods":
						for _, v := range result.Values {
							if v.Name == "PodPhase" && *v.Value.String == "Running" {
								return nil
							}
						}
					}
					return fmt.Errorf("the PodPhase of pod workprober is not Succeeded")
				},
			},
		},
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

	err = addonMgr.AddAgent(&leaseProberAddonAgent{})
	if err != nil {
		klog.Errorf("unable to add addon agent: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	go addonMgr.Start(ctx)

	<-ctx.Done()
}
