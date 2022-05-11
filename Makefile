# Set the shell to bash always
SHELL := /bin/bash

# Look for a .env file, and if present, set make variables from it.
ifneq (,$(wildcard ./.env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

KIND_CLUSTER_NAME ?= local-dev
KUBECONFIG ?= $(HOME)/.kube/config

VERSION := $(shell git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')
ifndef VERSION
VERSION := 0.0.0
endif

BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
REPO_URL := $(shell git config --get remote.origin.url | sed "s/git@/https\:\/\//; s/\.com\:/\.com\//; s/\.git//")
LAST_COMMIT := $(shell git log -1 --pretty=%h)

PROJECT_NAME := provider-github
ORG_NAME := krateoplatformops
VENDOR := Kiratech

# Github Container Registry
DOCKER_REGISTRY := ghcr.io/$(ORG_NAME)

TARGET_OS := linux
TARGET_ARCH := amd64

# Tools
KIND=$(shell which kind)
LINT=$(shell which golangci-lint)
KUBECTL=$(shell which kubectl)
DOCKER=$(shell which docker)
SED=$(shell which sed)

.DEFAULT_GOAL := help

.PHONY: print.vars
print.vars: ## print all the build variables
	@echo VENDOR=$(VENDOR)
	@echo ORG_NAME=$(ORG_NAME)
	@echo PROJECT_NAME=$(PROJECT_NAME)
	@echo REPO_URL=$(REPO_URL)
	@echo LAST_COMMIT=$(LAST_COMMIT)
	@echo VERSION=$(VERSION)
	@echo BUILD_DATE=$(BUILD_DATE)
	@echo TARGET_OS=$(TARGET_OS)
	@echo TARGET_ARCH=$(TARGET_ARCH)
	@echo DOCKER_REGISTRY=$(DOCKER_REGISTRY)


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

.PHONY: image.build
image.build: ## build the controller Docker image
	@$(DOCKER) build -t "$(DOCKER_REGISTRY)/$(PROJECT_NAME):$(VERSION)" \
	--build-arg METRICS_PORT=9090 \
	--build-arg VERSION="$(VERSION)" \
	--build-arg BUILD_DATE="$(BUILD_DATE)" \
	--build-arg REPO_URL="$(REPO_URL)" \
	--build-arg LAST_COMMIT="$(LAST_COMMIT)" \
	--build-arg PROJECT_NAME="$(PROJECT_NAME)" \
	--build-arg VENDOR="$(VENDOR)" .
	@$(DOCKER) rmi -f $$(docker images -f "dangling=true" -q)


.PHONY: image.push
image.push: ## Push the Docker image to the Github Registry
	@$(DOCKER) push "$(DOCKER_REGISTRY)/$(PROJECT_NAME):$(VERSION)"

.PHONY: install.crossplane
install.crossplane: ## Install Crossplane into the local KinD cluster
	$(KUBECTL) create namespace crossplane-system || true
	helm repo add crossplane-stable https://charts.crossplane.io/stable
	helm repo update
	helm install crossplane --namespace crossplane-system crossplane-stable/crossplane


.PHONY: cr.secret
cr.secret: ## Create the secret for container registry credentials
	$(KUBECTL) create secret docker-registry cr-token \
	--namespace crossplane-system --docker-server=ghcr.io \
	--docker-password=$(PROVIDER_GIT) --docker-username=$(ORG_NAME) || true


.PHONY: install.provider
install.provider: cr.secret ## Install this provider
	@$(SED) 's/VERSION/$(VERSION)/g' ./examples/provider.yaml | $(KUBECTL) apply -f -

.PHONY: example.secrets
example.secret: ## Create the example secrets
	@$(KUBECTL) create secret generic github-secret --from-literal=token=$(PROVIDER_GIT) || true


.PHONY: help
help: ## print this help
	@grep -E '^[a-zA-Z\._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
