package helmctlr

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/tiller/environment"
)

// OwnerRefEngine wraps a tiller Render engine, adding ownerrefs to rendered assets
type OwnerRefEngine struct {
	environment.Engine
	refs []metav1.OwnerReference
}

// assert interface
var _ environment.Engine = &OwnerRefEngine{}

// Render proxies to the wrapped Render engine and then adds ownerRefs to each rendered file
func (o *OwnerRefEngine) Render(chart *chart.Chart, values chartutil.Values) (map[string]string, error) {
	rendered, err := o.Engine.Render(chart, values)
	if err != nil {
		return nil, err
	}

	ownedRenderedFiles := map[string]string{}
	for fileName, renderedFile := range rendered {
		withOwner, err := o.addOwnerRefs(renderedFile)
		if err != nil {
			return nil, err
		}
		ownedRenderedFiles[fileName] = withOwner
	}
	return ownedRenderedFiles, nil
}

// addOwnerRefs adds the configured ownerRefs to a single rendered file
func (o *OwnerRefEngine) addOwnerRefs(fileContents string) (string, error) {
	parsed := chartutil.FromYaml(fileContents)
	if errors, ok := parsed["Error"]; ok {
		return "", fmt.Errorf("error parsing rendered template to add ownerrefs: %v", errors)
	}

	unst, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&parsed)
	if err != nil {
		return "", err
	}
	unstructured := &unstructured.Unstructured{Object: unst}
	unstructured.SetOwnerReferences(o.refs)
	return chartutil.ToYaml(unstructured.Object), nil
}

// NewOwnerRefEngine creates a new OwnerRef engine with a set of metav1.OwnerReferences to be added to assets
func NewOwnerRefEngine(baseEngine environment.Engine, refs []metav1.OwnerReference) environment.Engine {
	return &OwnerRefEngine{
		Engine: baseEngine,
		refs:   refs,
	}
}
