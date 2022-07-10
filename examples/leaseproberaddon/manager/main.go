package main

import (
	"context"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

const (
	addonName             = "leaseprober-addon"
	addonInstallNamespace = "open-cluster-management-agent-addon"
)

type leaseProberAddonAgent struct{}

func (a *leaseProberAddonAgent) Manifests(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) ([]runtime.Object, error) {
	objects := []runtime.Object{
		&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "leaseprober-addon-agent",
				Namespace: addon.Spec.InstallNamespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"addon-agent": "leaseprober-addon",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"addon-agent": "leaseprober-addon",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "leaseprober-addon-agent",
								Image:           "quay.io/open-cluster-management/addons:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command: []string{
									"/leaseprober-agent",
								},
							},
						},
						ServiceAccountName: "leaseprober-addon-agent-sa",
					},
				},
			},
		},
		&rbacv1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "leaseprober-addon-agent",
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"coordination.k8s.io"},
					Resources: []string{"leases"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps", "events"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "delete", "deletecollection", "patch"},
				},
			},
		},
		&rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "leaseprober-addon-agent",
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: "leaseprober-addon-agent",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "leaseprober-addon-agent-sa",
					Namespace: addon.Spec.InstallNamespace,
				},
			},
		},
		&corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: addon.Spec.InstallNamespace,
				Name:      "leaseprober-addon-agent-sa",
			},
		},
	}

	return objects, nil
}

func (a *leaseProberAddonAgent) GetAgentAddonOptions() agent.AgentAddonOptions {
	return agent.AgentAddonOptions{
		AddonName:       addonName,
		InstallStrategy: agent.InstallAllStrategy(addonInstallNamespace),
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
