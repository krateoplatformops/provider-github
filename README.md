# Provider Git

## Overview

This is a Kubernetes Operator (Crossplane provider) that clones one remote git repository over another one.

The provider that is built from the source code in this repository adds the following new functionality:

- a Custom Resource Definition (CRD) that model git repositories remote clones

## Getting Started

With Crossplane installed in your cluster:

```sh
$ helm repo add crossplane-stable https://charts.crossplane.io/stable
$ helm repo update
$ helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

### Install this provider

Before installing the below manifest:

- replace `VERSION` with latest or your preferred provider version
- since this repo is private, create a docker secret named `cr-token` with the docker credentials

  ```sh
  $ create secret docker-registry cr-token \
	  --namespace crossplane-system --docker-server=ghcr.io \
	  --docker-password=$(YOUR_GHCR_TOKEN) --docker-username=$(ORG_NAME) || true
  ```


```sh
$ cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1alpha1
kind: ControllerConfig
metadata:
  name: debug-config
spec:
  args:
    - --debug
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: crossplane-provider-git
spec:
  package: 'ghcr.io/krateoplatformops/crossplane-provider-git:VERSION'
  packagePullPolicy: Always
  packagePullSecrets:
  - name: cr-token
  controllerConfigRef:
    name: debug-config
EOF
```

### Configure the `Repo` CRD instance


```sh
$ cat <<EOF | kubectl apply -f -
apiVersion: git.krateo.io/v1alpha1
kind: ProviderConfig
metadata:
  name: provider-git-config
spec:
  verbose: false
---
apiVersion: git.krateo.io/v1alpha1
kind: Repo
metadata:
  name: git-provider-example
spec:
  forProvider:
    fromRepo:
      url: # ENTER SOURCE REPOSITORY URL HERE
      path: skeleton
      apiCredentials:
        source: Secret
        secretRef:
          namespace: default
          name: from-repo-token
          key: token
    toRepo:
      url: # ENTER DESTINATION REPOSITORY URL HERE
      private: true
      apiCredentials:
        source: Secret
        secretRef:
          namespace: default
          name: to-repo-token
          key: token
  providerConfigRef:
    name: provider-git-config
EOF
```

You can found example manifest files here:

- provider [provider.yaml](./examples/provider.yaml)
- provider config [config.yaml](./examples/config.yaml)
- crd instance [example.yaml](./examples/example.yaml)

---

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please open an [issue](https://github.com/krateoplatformops/provider-git/issues).