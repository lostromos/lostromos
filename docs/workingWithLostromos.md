# Working with Lostrómos

## Tutorial

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

* `crd` (Required) Custom Resource information used to determine what Lostrómos will watch for add/update/delete.
  * `name` (Required) The plural name of the Custom Resource Definition (CRD) you want monitored (ex: users)
  * `group` (Required) The group of the CRD you want monitored (ex: stable.wpengine.io)
  * `version` (Required) The version of the CRD you want monitored
  * `namespace` The namespace of the CRD you want monitored
  * `filter` Filter to specify if Lostromos will act on a resource create/update/delete. For more detailed information
             about what events happen on filtered updates check [here](./events.md)
* `helm` Information pertaining to helm deployments. Defaults to use the go template controller if no information is
    given.
  * `chart` Path to helm chart
  * `namespace` Namespace for resources deployed by helm
  * `releasePrefix` Prefix for release names in helm
  * `tiller` Address for helm tiller.
* `k8s` Kubernetes configuration file required to run Lostrómos on a different cluster. Defaults to use local cluster if
    no config is specified.
  * `config` Path to configuration file
* `templates` Path to template directory. If using helm, this is skipped. Defaults to "".

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