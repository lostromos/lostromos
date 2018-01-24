package helmctlr

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/json-iterator/go/require"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type mockEngine struct {
	out map[string]string
}

func (e *mockEngine) Render(chrt *chart.Chart, v chartutil.Values) (map[string]string, error) {
	return e.out, nil
}

func TestOwnerRefEngine(t *testing.T) {
	ownerRefs := []metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Test",
			Name:       "test",
			UID:        "123",
		},
	}

	baseOut := `apiVersion: stable.nicolerenee.io/v1
kind: Character
metadata:
  name: nemo
spec:
  Name: Nemo
`

	expectedOut := `apiVersion: stable.nicolerenee.io/v1
kind: Character
metadata:
  name: nemo
  ownerReferences:
  - apiVersion: v1
    kind: Test
    name: test
    uid: "123"
spec:
  Name: Nemo
`
	expected := map[string]string{"template.yaml": expectedOut, "template2.yaml": expectedOut}

	baseEngineOutput := map[string]string{
		"template.yaml":  baseOut,
		"template2.yaml": baseOut,
	}

	engine := NewOwnerRefEngine(&mockEngine{out: baseEngineOutput}, ownerRefs)
	out, err := engine.Render(&chart.Chart{}, map[string]interface{}{})
	require.NoError(t, err)
	require.EqualValues(t, expected, out)
}
