SHELL :=/bin/bash

all: build
.PHONY: all

GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

# Tools for deploy
KUBECONFIG ?= ./.kubeconfig
KUBECTL ?= kubectl
PWD=$(shell pwd)
OUTPUTDIR ?= $(PWD)/_output
KUSTOMIZE ?= $(OUTPUTDIR)/kustomize
KUSTOMIZE_VERSION ?= v3.5.4
KUSTOMIZE_ARCHIVE_NAME ?= kustomize_$(KUSTOMIZE_VERSION)_$(GOHOSTOS)_$(GOHOSTARCH).tar.gz
kustomize_dir:=$(dir $(KUSTOMIZE))


IMAGE ?= addons
IMAGE_REGISTRY ?= quay.io/open-cluster-management
IMAGE_TAG ?= latest
EXAMPLE_IMAGE_NAME ?= $(IMAGE_REGISTRY)/$(IMAGE):$(IMAGE_TAG)


fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

build: fmt vet ## Build manager binary.
	go build -a -o $(OUTPUTDIR)/busybox-addon examples/busyboxaddon/manager/main.go
	go build -a -o $(OUTPUTDIR)/leaseprober-addon examples/leaseproberaddon/manager/main.go
	go build -a -o $(OUTPUTDIR)/leaseprober-agent examples/leaseproberaddon/agent/main.go
	go build -a -o $(OUTPUTDIR)/workprober-addon examples/workproberaddon/manager/main.go
	go build -a -o $(OUTPUTDIR)/large-addon examples/largeaddon/manager/main.go

images:
	docker build -t $(EXAMPLE_IMAGE_NAME) -f Dockerfile .


deploy-ocm:
	deploy/ocm/install.sh

deploy-busybox-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/busybox-addon | $(KUBECTL) apply -f -

undeploy-busybox-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/busybox-addon | $(KUBECTL) delete --ignore-not-found -f -

deploy-leaseprober-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/leaseprober-addon | $(KUBECTL) apply -f -

undeploy-leaseprober-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/leaseprober | $(KUBECTL) delete --ignore-not-found -f -

deploy-workprober-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/workprober-addon | $(KUBECTL) apply -f -

undeploy-workprober-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/workprober-addon | $(KUBECTL) delete --ignore-not-found -f -

deploy-helm-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/helm-addon | $(KUBECTL) apply -f -

undeploy-helm-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/helm-addon | $(KUBECTL) delete --ignore-not-found -f -

deploy-large-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/large-addon | $(KUBECTL) apply -f -

undeploy-large-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/addons/large-addon | $(KUBECTL) delete --ignore-not-found -f -



# Ensure kustomize
ensure-kustomize:
ifeq "" "$(wildcard $(KUSTOMIZE))"
	$(info Installing kustomize into '$(KUSTOMIZE)')
	mkdir -p '$(kustomize_dir)'
	curl -s -f -L https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F$(KUSTOMIZE_VERSION)/$(KUSTOMIZE_ARCHIVE_NAME) -o '$(kustomize_dir)$(KUSTOMIZE_ARCHIVE_NAME)'
	tar -C '$(kustomize_dir)' -zvxf '$(kustomize_dir)$(KUSTOMIZE_ARCHIVE_NAME)'
	chmod +x '$(KUSTOMIZE)';
else
	$(info Using existing kustomize from "$(KUSTOMIZE)")
endif
