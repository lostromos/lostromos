![Lostrómos logo](images/logo.png)

# <a name="usinglostromos"></a>Using Lostrómos

## Recommended Reading
* [Custom Resource Definitions](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/)
* [Custom Resources](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
* [Kubernetes Operators](https://coreos.com/blog/introducing-operators.html)
* [Helm](https://docs.helm.sh/)
* [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)

## <a name="quickstart"></a>Quick Start

### Dependencies

| Dependency | Version |
| ---------- | ------- |
| `Golang` | 1.9.0+ |
| `Minikube` | 0.22.3+ |
| `Docker` | 17.09.0+ |
| `Python` | 3.0+ |

Install the above dependencies. Run the following script (changing out your_os_version for darwin/linux/windows depending on your system) to get a basic
setup. This script will install Go and Python dependencies, build Lostrómos, build a docker image with Lostrómos, then
run it in Minikube and perform integration testing.

```bash
make install-go-deps
make vendor
make install-python-deps
make build-cross
./out/lostromos-{your_os_version}-amd64 version
minikube start
eval $(minikube docker-env) # This links docker with minikube so that the image you build in the next step will be available.
make docker-build-test
kubectl create -f test/data/crd.yml
make LOSTROMOS_IP_AND_PORT=`minikube service lostromos --url | cut -c 8-` integration-tests
eval $(minikube docker-env -u) # Unlinks minikube and docker.```

## <a name="tutorial"></a>Tutorial

To start working with Lostrómos, you should begin by playing around with some basic CR management. Build Lostrómos via
`make build`, then do some/all of the following steps.

1. Setup kubectl against a cluster (minikube works just fine)
2. `kubectl apply -f test/data/crd.yml`
3. `kubectl apply -f test/data/cr_things.yml`
4. `./lostromos start --config test/data/config.yaml --nop`
  - See that it prints out that `thing1` and `thing2` were added
5. In another shell `kubectl apply -f test/data/cr_nemo.yml`
  - See that it prints out that `nemo` was added
6. `kubectl edit character nemo` and change a field
  - See that it prints out that `nemo` was changed
7. `kubectl delete -f test/data/cr_nemo.yml`
  - See that it prints out that `nemo` was deleted
8. `kubectl delete -f test/data/cr_things.yml`
  - See that it prints out that `thing1` and `thing2` were deleted
9. You can stop the process and `kubectl delete -f test/data/crd.yml` to cleanup the rest of the test data.

## Customization

### Configuration file/flags

* `crd` (Required) Custom Resource information used to determine what Lostrómos will watch for add/update/delete
  * `name` (Required) The plural name of the Custom Resource Definition (CRD) you want monitored (ex: users)
  * `group` (Required) The group of the CRD you want monitored (ex: stable.wpengine.io)
  * `version` (Required) The version of the CRD you want monitored
  * `namespace` The namespace of the CRD you want monitored
  * `filter` Filter to specify if Lostromos will act on a resource create/update/delete. For more detailed information about what events happen on filtered updates check [here](./events.md)
* `helm` Information pertaining to helm deployments. Defaults to use the go template controller if no information is
    given
  * `chart` Path to helm chart
  * `namespace` Namespace for resources deployed by helm
  * `releasePrefix` Prefix for release names in helm
  * `tiller` Address for helm tiller
* `k8s` Kubernetes configuration file required to run Lostrómos on a different cluster. Defaults to use local cluster if
    no config is specified
  * `config` Path to configuration file
* `templates` Path to template directory. If using helm, this is skipped. Defaults to ""

See `./lostromos start --help` for more info.

[Sample config file](../test/data/config.yaml)

### Templates

#### Helm Templates

From the CR, `metadata.name`, `metadata.namespace`, and `spec` fields are marshalled. These values are accessible in
Helm as `Values.resource.name`, `Values.resource.namespace`, and `Values.resource.spec` respectively.

See documentation on [Using Helm](./helm.md) for more info

[Sample helm template](../test/data/helm/chart/templates/deployment.yaml)

#### Go Templates

CR fields are accessible to the template by using .GetField

[Sample go template](../test/data/templates/deployment.yaml.tmpl)


## <a name="deployment"></a>Deployment

We don't deploy a docker image with Lostrómos, but instead only release the binary to github so that it can be built in
to the image you need for your particular application. This is due to the link between kubectl and Lostromos, and it's
possible the kubectl we intend to use isn't valid for everyones use case. We do build a docker image as part of testing
that can be thought of as an [example](../test/docker/Dockerfile) however.

### GOOS Environment Variable

In order for `lostromos` to work in a scratch/alpine container (and therefore be as small as possible), we need to build
the binary with the normal `make build` command, but having set GOOS=linux.

### Kubectl Version

We test with [kubectl] version 1.7.x for our installs due to an issue with 1.8 that causes error messages
(false failures).
