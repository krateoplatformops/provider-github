# Provider GitHub

## Overview

This is a Kubernetes Operator (Crossplane provider) that creates a GitHub repository.

## Getting Started

With Crossplane installed in your cluster:

```sh
$ helm repo add crossplane-stable https://charts.crossplane.io/stable
$ helm repo update
$ helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

### Install this provider

Install the below manifest:

```sh
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  namespace: krateo-system
  name: provider-github
spec:
  package: 'ghcr.io/krateoplatformops/provider-github:VERSION'
  packagePullPolicy: IfNotPresent
EOF
```

!!! note ""

   replace `VERSION` with [latest provider tag](https://github.com/krateoplatformops/provider-github/tags)

### Configure the provider

```sh
cat <<EOF | kubectl apply -f -
apiVersion: github.krateo.io/v1alpha1
kind: ProviderConfig
metadata:
  name: provider-github-demo-config
spec:
  # URL to Github API
  apiUrl: https://github.yourdomain.it/api/v3
  # Set verbose true to dump all client requests
  verbose: true
  # The secret with your Github PAT
  credentials:
    source: Secret
    secretRef:
      namespace: default
      name: github-secret
      key: token
EOF
```


### Configure the `Repo` CRD instance

```sh
cat <<EOF | kubectl apply -f -
apiVersion: github.krateo.io/v1alpha1
kind: Repo
metadata:
  name: provider-github-demo
spec:
  forProvider:
    # Github Owner
    org: krateoplatformops
    # Repository name
    name: demo-repo
  providerConfigRef:
    name: provider-github-demo-config
EOF
```