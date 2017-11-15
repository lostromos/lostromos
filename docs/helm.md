# Using Helm with Lostrómos

In order to make migrating to Lostrómos easier we support using helm to handle
the templating. This allows you to provide a helm chart to Lostrómos.

## Tiller

For Lostrómos to be able to run commands with Helm it needs access to a tiller.
If you are running inside kubernetes the value for this would look like
`tiller-deploy:44134` if the tiller is in the same namespace as your Lostrómos
deployment. If you are in a different namespace you would use
`tiller-deploy.<namespace>:44134`.

## Version

Helm requires the the tiller and the client be running the same version.
Currently Lostrómos uses version v2.7.0 of helm.

## Running outside the cluster

If you are running outside of the kubernetes cluster access to the tiller is a
little more complicated. It is not a good idea to expose the tiller outside of
the cluster since it is not secure by default. Your best option is to
use the `kubectl port-forward` command. This is a simple set of commands you can
run that will setup a port forward for you that will work with the default tiller deployment.

```bash
export TILLER_NS=kube-system
export TILLER_POD=`kubectl -n $TILLER_NS get pods \
       --selector=app=helm,name=tiller \
       -o jsonpath="{range .items[*]}{@.metadata.name}{end}" | head -n1`
kubectl port-forward -n $TILLER_NS $TILLER_POD 44134:44134
```

After running this command you would start Lostrómos and set the tiller to point
to `127.0.0.1:44134`

## Using the Custom Resource in your charts

Lostrómos provides access to the resource from within your charts under the
variable name `resource`. The resource has the following fields available on it:

- `name` - Which is the custom resource name
- `namespace` - represents the namespace the custom resource is deployed in
- `spec` - holds all of the data that is stored in your resources spec field.

Given this custom resource

```yaml
apiVersion: stable.nicolerenee.io/v1
kind: Character
metadata:
  name: nemo
spec:
  Name: Nemo
  From: Finding Nemo
  By: Disney
```

- `{{ .Values.resource.name }}` would return "nemo"
- `{{ .Values.resource.spec.name }}` would return "Nemo"
- `{{ .Values.resource.spec.from }}` would return "Finding Nemo"
- `{{ .Values.resource.spec.by }}` would return "Disney"
