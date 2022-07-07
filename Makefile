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
	go build -a -o $(OUTPUTDIR)/busyboxaddon examples/busyboxaddon/manager/main.go

images:
	docker build -t $(EXAMPLE_IMAGE_NAME) -f Dockerfile .


deploy-ocm:
	deploy/ocm/install.sh

deploy-busybox-addon: ensure-kustomize
	cp deploy/busyboxaddon/kustomization.yaml deploy/busyboxaddon/kustomization.yaml.tmp
	cd deploy/busyboxaddon && $(KUSTOMIZE) edit set image example-addon-image=$(EXAMPLE_IMAGE_NAME) && $(KUSTOMIZE) edit add configmap image-config --from-literal=EXAMPLE_IMAGE_NAME=$(EXAMPLE_IMAGE_NAME)
	$(KUSTOMIZE) build deploy/busyboxaddon | $(KUBECTL) apply -f -
	mv deploy/busyboxaddon/kustomization.yaml.tmp deploy/busyboxaddon/kustomization.yaml

undeploy-busybox-addon: ensure-kustomize
	$(KUSTOMIZE) build deploy/busyboxaddon | $(KUBECTL) delete --ignore-not-found -f -


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
