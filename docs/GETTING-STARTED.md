# Getting Started

## 1. Requirement

### ✓ Kubernetes Version
  
Version ≥ 1.19

COSMO is targeting to run on any Kubernetes distribution.

### 🌐 Networking & DNS

COSMO uses Traefik Ingress Controller and Host-header routing for each Workspaces.

Configure a wildcard-host domain record like `*.YOUR_DOMAIN.com` in your name server targeting to your Load Balancer for Traefik Proxy.

For example in Amazon EKS using AWS Load Balancer Controller and Route53 for a name server

### 💾 Volumes

To persist workspace data, each workspace templates should include a [Persistent Volume Claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).

You need to configure [Dynamic Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/) in Kubernetes, which does not require a Persistent Volume manually created by administrator.

- **On-prem or self-installed cluster**

  See [Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/) or install storage extentions suporting Dynamic Provisioning such as [Longhorn](https://longhorn.io/docs/1.2.0/volumes-and-nodes/create-volumes/) or [OpenEBS](https://github.com/openebs/openebs).

- **Cloud providers's cluster**

  See the official docs. Almost all cloud providers supports their own storage driver.

  For example:
  - [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/storage.html)

  - [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/ssd-pd)

  - [Azure Kubernetes Service](https://docs.microsoft.com/azure/aks/concepts-storage#storage-classes)

## 2. Install COSMO on your Kubernetes Cluster

Confirm followings before installation.
- [kubectl](https://kubernetes.io/ja/docs/tasks/tools/install-kubectl/) and [helm](https://helm.sh/docs/intro/install/) is installed in PATH.
- kubectl current context to the cluster where you install COSMO.

### 2-1. Install [`cert-manager`](https://cert-manager.io/)

COSMO requires a lot of TLS certificates.

The easiest way to prepare certificates is to use [`cert-manager`](https://cert-manager.io/) (>=v1.0.0)

See the offical install docs https://cert-manager.io/docs/installation/

```sh
kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
```

### 2-2. Install COSMO

Add helm repo

```sh
helm repo add cosmo https://cosmo-workspace.github.io/charts
```

Install cosmo

```sh
WORKSPACE_DOMAIN=YOUR_DOMAIN.com

# When you configured TLS cert '*.YOURDOMAIN.com' for traefik
helm upgrade --install -n cosmo-system --create-namespace cosmo-controller-manager cosmo/cosmo-controller-manager \
  --set domain=$WORKSPACE_DOMAIN

# When you did't configured TLS cert '*.YOURDOMAIN.com' for traefik
helm upgrade --install -n cosmo-system --create-namespace cosmo-controller-manager cosmo/cosmo-controller-manager \
  --set domain=$WORKSPACE_DOMAIN \
  --set protocol=http
```

Output:

```sh
### Example Output
Release "cosmo" does not exist. Installing it now.
NAME: cosmo
LAST DEPLOYED: Tue Jul 18 02:37:15 2023
NAMESPACE: cosmo-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
COSMO is installed!

* Your Environment Infomation
  +-----------------------+------------------------------------------------------------
  | DOMAIN                | *.YOUR_DOMAIN.com
  | DASHBOARD_URL         | https://dashboard.YOUR_DOMAIN.com/#/signin
  | WORKSPACE_URLBase     | https://{{NETRULE}}-{{WORKSPACE}}-{{USER}}.YOUR_DOMAIN.com
  +-----------------------+------------------------------------------------------------
```

Now access COSMO Dashboard URL you configured from your browser.

> 💡 See [charts repository](https://github.com/cosmo-workspace/cosmo/blob/main/charts/README.md) for more information on helm installation options.

### 2-3. Install cosmoctl

Download binary from [latest release](https://github.com/cosmo-workspace/cosmo/releases/latest) and extract it into PATH.

## 3. Create first User

Use cosmoctl to create first User.

```sh
cosmoctl user create admin --privileged
```

Output:

```sh
### Example Output
Successfully created user admin
Default password: DEFAULT_PASSWORD
```

Access the dashboard with a browser and login as admin with the output password. 

Now you are ready to use COSMO 🚀

## 4. Create example Template

However there is still NO Template.
Create example Templates.

```sh
kubectl create -f https://raw.githubusercontent.com/cosmo-workspace/cosmo/main/example/workspaces/code-server.yaml
```

Now you are ready to use Template `code-server-example` 🙌

## 5. Create Workspace

From browser, create a example workspace with created Template.


<video src="https://user-images.githubusercontent.com/9918931/136198651-75626165-d421-4fa8-bbf0-932d7020ae84.mp4"></video>


## 6. Cleanup

Uninstall COSMO.

```sh
helm uninstall -n cosmo-system cosmo
```

Uninstall COSMO CRD.

This will remove all Workspaces and Templates.

```sh
kubectl delete -k https://github.com/cosmo-workspace/cosmo/config/crd/
```
