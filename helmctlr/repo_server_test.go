package helmctlr_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo/repotest"
)

type TestRepoServer struct {
	server *repotest.Server
	tmp    string
}

// SetupMockServer sets up test helm repo server
func SetupMockServer(t *testing.T) *TestRepoServer {
	// Create a temp local server
	tmp, err := ioutil.TempDir("", "helm-downloadto-")
	if err != nil {
		t.Fatal(err)
	}

	hh := helmpath.Home(tmp)
	dest := filepath.Join(hh.String(), "dest")
	configDirectories := []string{
		hh.String(),
		hh.Repository(),
		hh.Cache(),
		dest,
	}

	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			if err := os.MkdirAll(p, 0755); err != nil {
				t.Fatalf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			t.Fatalf("%s must be a directory", p)
		}
	}

	// Set up a fake repo
	repoSrv := repotest.NewServer(tmp)
	if _, err := repoSrv.CopyCharts("../test/data/helm/*.tgz*"); err != nil {
		t.Error(err)
		return &TestRepoServer{nil, tmp}
	}
	if err := repoSrv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	// Set helm home to be helm home of the temp local server
	os.Setenv("HELM_HOME", tmp)

	return &TestRepoServer{repoSrv, tmp}
}

// Cleanup stops the helm repo server and cleans up related artifacts
func (t *TestRepoServer) Cleanup() {
	t.server.Stop()
	os.RemoveAll(t.tmp)
}
