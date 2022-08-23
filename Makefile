# Set the shell to bash always
SHELL := /bin/bash

# Look for a .env file, and if present, set make variables from it.
ifneq (,$(wildcard ./.env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

KIND_CLUSTER_NAME ?= local-dev
KUBECONFIG ?= $(HOME)/.kube/config

VERSION := $(shell git describe --always --tags | sed 's/-/./2' | sed 's/-/./2')
ifndef VERSION
VERSION := 0.0.0
endif

# Tools
KIND=$(shell which kind)
LINT=$(shell which golangci-lint)
KUBECTL=$(shell which kubectl)
HELM=$(shell which helm)
SED=$(shell which sed)

.DEFAULT_GOAL := help

.PHONY: dev
dev: generate ## run the controller in debug mode
	$(KUBECTL) apply -f package/crds/ -R
	go run cmd/main.go -d

.PHONY: generate
generate: tidy ## generate all CRDs
	go generate ./...

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: test
test: ## go test
	go test -v ./...

.PHONY: lint
lint: ## go lint
	$(LINT) run

.PHONY: kind.up
kind.up: ## starts a KinD cluster for local development
	@$(KIND) get kubeconfig --name $(KIND_CLUSTER_NAME) >/dev/null 2>&1 || $(KIND) create cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: kind.down
kind.down: ## shuts down the KinD cluster
	@$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: install.crossplane
install.crossplane: ## Install Crossplane into the local KinD cluster
	$(KUBECTL) create namespace crossplane-system || true
	$(HELM) repo add crossplane-stable https://charts.crossplane.io/stable
	$(HELM) repo update
	$(HELM) install crossplane --namespace crossplane-system crossplane-stable/crossplane


.PHONY: install.provider
install.provider: ## Install this provider
	@$(SED) 's/VERSION/$(VERSION)/g' ./examples/provider.yaml | $(KUBECTL) apply -f -

.PHONY: install.eventrouter
install.eventrouter: ## Install the event router
	$(HELM) repo add krateo https://charts.krateo.io
	$(HELM) repo update krateo
	$(HELM) install --set EVENT_ROUTER_DEBUG=true eventrouter krateo/eventrouter

.PHONY: demo
demo: ## Run the demo examples
	@$(KUBECTL) create secret generic github-secret --from-literal=token=$(PROVIDER_GITHUB_DEMO_TOKEN) || true
	@$(KUBECTL) apply -f examples/demo-config.yaml
	@$(KUBECTL) apply -f examples/demo-repo.yaml


.PHONY: help
help: ## print this help
	@grep -E '^[a-zA-Z\._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
