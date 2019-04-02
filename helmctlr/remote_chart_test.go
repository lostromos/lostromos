package helmctlr_test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/lostromos/lostromos/helmctlr"
)

func Test_getRemoteChart(t *testing.T) {
	repoSrv := SetupMockServer(t)
	defer repoSrv.Cleanup()

	if repoSrv == nil {
		t.FailNow()
	}

	chartArchive := "chart-0.1.0.tgz"
	chartPath := filepath.Join(os.TempDir(), "test/helloworld", helmctlr.Hash(chartArchive), chartArchive)
	tests := []struct {
		name    string
		args    string
		f       func()
		want    string
		wantErr bool
	}{
		{"HappyCase", "test/helloworld:0.1.0", func() {}, chartPath, false},
		{"ChartNoVersion", "test/helloworld:", func() {}, chartPath, false},
		{"DirCreationFailed", "test/helloworld:0.1.0", func() { createSymlinkReplacingDir(chartPath) }, "", true},
		{"DownloadFailed", "test/helloworld:0.1.0", func() { readOnlyChartDir(chartPath) }, "", true},
		{"ChartNoName", ":1.2rev34", func() {}, "", true},
		{"InvalidChart", "test/test_chart:1.2rev34", func() {}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset(filepath.Dir(chartPath))
			tt.f()

			got, err := testController.GetRemoteChart(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRemoteChart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRemoteChart() = %v, want %v", got, tt.want)
			}

			reset(filepath.Dir(chartPath))
		})
	}
}

func readOnlyChartDir(chartPath string) {
	chartDir := filepath.Dir(chartPath)
	os.RemoveAll(chartDir)
	os.Mkdir(chartDir, 0400)
}

func createSymlinkReplacingDir(chartPath string) {
	chartDir := filepath.Dir(chartPath)
	os.RemoveAll(chartDir)
	parent, _ := path.Split(chartDir)
	os.Chmod(parent, 0400)
}

func reset(chartDir string) {
	os.RemoveAll(chartDir)
	parent, _ := path.Split(chartDir)
	os.Chmod(parent, 0700)
}

func Test_getChartRef(t *testing.T) {
	type args struct {
		r *unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Happy case", args{&unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{"name": "dory",
				"annotations": map[string]interface{}{"chart": "test/helloworld:1.2rev34"}},
			"spec": map[string]interface{}{"Name": "Dory"}}}},
			"test/helloworld:1.2rev34"},
		{"No chart present", args{&unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{"name": "dory"},
			"spec":     map[string]interface{}{"Name": "Dory"}}}},
			""},
		{"No spec present", args{&unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{"name": "dory"}}}},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helmctlr.GetChartRef(tt.args.r)
			if got != tt.want {
				t.Errorf("getChartRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitChartRef(t *testing.T) {
	type args struct {
		chart string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"Happy case", args{"test/helloworld:1.2rev34"}, "test/helloworld", "1.2rev34"},
		{"No version but colon", args{"test/helloworld:"}, "test/helloworld", ""},
		{"No version no colon", args{"test/helloworld:"}, "test/helloworld", ""},
		{"No chart name", args{":1.2rev34"}, "", "1.2rev34"},
		{"More than 1 colon", args{"test/helloworld:1.2rev34:456"}, "test/helloworld", ""},
		{"Empty string", args{""}, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := helmctlr.SplitChartRef(tt.args.chart)
			if got != tt.want {
				t.Errorf("splitChartRef() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitChartRef() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
