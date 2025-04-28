# Getting Started

Currently, getting started is a bit complicated due to the images only being hosted on the Github Repository and the helm chart not existing as a directly pullable chart.

## Requirements
* Kubernetes Cluster
* Github Account
* git installed on a machine with the cluster accessible

## Setup the github secret in Kubernetes

### Personal Access Token
For this, you need to first setup a Github Personal Access Token with read:packagages permission. This can be done on the 
[Personal Access Tokens](https://github.com/settings/tokens) page.

### Kubernetes Secret
After that, you need to setup the token as a secret in Kubernetes, you can do that with:
```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token> \
  --docker-email=<email-address>
```
Note the name you set for this secret, and that it is available in the same namespace where you want to install the controller to.

## Installing the charts
### Git Clone
First, clone the repository to your machine. The precise way to do that depends on your setup, but usually:
`git clone https://github.com/UnfamousThomas/thesis-initial`

### Installing the CRDs chart
Installing the custom resources chart is quite straightforward, simply navigate to:
`operator/charts/` and then run 
```bash
helm install thesis-crds ./thesis-crds
```
This will setup the crds on the cluster.

### Installing the manager chart
Installing the manager is a little bit more complicated, as you need to make sure the values are correct first. First, navigate to:
`operator/charts/thesis-operator`. 

#### Values.yaml
This folder has a `values.yaml` file, with all the files, it should look like this:

```yaml
controllerManager:
  manager:
    args:
    - --metrics-bind-address=:8443
    - --leader-elect
    - --health-probe-bind-address=:8081
    containerSecurityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
    env:
      imagePullSecretName: ghcr-secret
    image:
      repository: ghcr.io/unfamousthomas/controller
      tag: latest
    resources:
      limits:
        cpu: 500m
        memory: 128Mi
      requests:
        cpu: 10m
        memory: 64Mi
  replicas: 1
  serviceAccount:
    annotations: {}
kubernetesClusterDomain: cluster.local
metricsService:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: 8443
  type: ClusterIP
webhookService:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 9443
  type: ClusterIP
service:
  enabled: true
```

The important parts of this are the following fields:
`controllerManager.manager.env.imagePullSecretName` and `service.enabled`.

The `controllerManager.manager.env.imagePullSecretName` field should be set to be what you created the github access token secret as. This is used to
pull the sidecar image.

The `service.enabled` determines whether or not we deploy the Service to the cluster. You can read about that in [Service](service.md).

#### Installing the chart
Once you are good with the values.yaml, just run:
```bash
helm install thesis-manager .
```
in the folder.