# Provider GitHub

## Overview

This is a Kubernetes Operator (Crossplane provider) that creates a GitHub repository.

The provider that is built from the source code in this repository adds the following new functionality:

- a Custom Resource Definition (CRD) that model git repositories github

## Getting Started

With Crossplane installed in your cluster:

```sh
$ helm repo add crossplane-stable https://charts.crossplane.io/stable
$ helm repo update
$ helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

### Install this provider

Before installing the below manifest:

- [replace `VERSION` with latest or your preferred provider version](./examples/provider.yaml)

### Configure the `Repo` CRD instance

You can found example manifest files here:

- provider config [config.yaml](./examples/config.yaml)
- crd instance [example.yaml](./examples/example.yaml)
