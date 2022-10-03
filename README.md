# screeps-server-operator

When developing a Screeps bot a time comes when experiments need to be run and observed. The process of standing up
and running multiple Screeps Private Servers can be tedious and error prone. The purpose of this project is to reduce
the pain around standing up, configuring, running, and cleaning-up Private Servers. This is accomplished by running a
this controller in a Kubernetes (K8s) cluster. The controller will handle the entire Private Server lifecycle.
Initially, manual creations/updates/deletions of Custom Resources (CRs) will drive Private Server management.
Eventually, those CRs will be linked to a Pull Request in GitHub, so that servers are created/destroyed by the
opening/closing of Pull Requests using GitHub Actions.

This project uses [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

#### Roadmap

- [ ] Stand up Redis server
- [ ] Stand up Mongo server
- [ ] Create ConfigMaps and Secrets
- [ ] Stand up Private Server
- [ ] Create Service and Ingress for Private Server
- [ ] Milestone: Able to shell into Private Server pod and manually configure user and bot
- [ ] Automatically configure Private Server global settings when Server is started
- [ ] Automatically configure user account when Private Server is available (may be build step in bot repo)
- [ ] Automatically upload bot after user account is created (may be build step in bot repo)
- [ ] Milestone: Creating a ScreepsServer CR results in a fully running simulation with provided Code
- [ ] Stand up single Prometheus and Grafana for Private Servers
- [ ] Configure Private Server Collector to scrape metrics from 
- [ ] Metrics for multiple bots in each server being collected
- [ ] Drive creation/deletion of ScreepServer CRs with GitHub Actions
 
## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

> Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/screeps-server-controller:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/screeps-server-controller:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

> You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

> Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Notes

```
kubebuilder init --domain pedanticorderliness.com --repo github.com/ryanrolds/screeps-server-controller
kubebuilder create api --group screeps --version v1 --kind ScreepsServer
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
