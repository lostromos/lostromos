package helmctlr

import (
	"strings"

	"fmt"
	"os"
	"path/filepath"

	"errors"

	"encoding/base64"

	"github.com/wpengine/lostromos/metrics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
)

// GetChartRef reads the chart entry from CR spec and returns chart entry or empty string.
func GetChartRef(r *unstructured.Unstructured) string {
	if chart, ok := r.GetAnnotations()["chart"]; ok {
		return chart
	}

	return ""
}

// GetRemoteChart Gets chart name and chart version from the passed chartRef, downloads the chart from repo,
// puts it into versioned folder, and returns that folder
// Uses chart downloader to download the charts
// 	- chart downloader uses the version passed, but if version is empty, pulls the latest version.
// Prerequisite: Repo should have been initialized under HELM_HOME
func (c *Controller) GetRemoteChart(chartRef string) (string, error) {
	chartDir, err := getRemoteChart(chartRef)
	if err != nil {
		metrics.RemoteRepoError.Inc()
		c.logger.Errorw("failed to fetch chart from remote repo", "error", err, "chart", chartRef)
	} else {
		metrics.RemoteRepoReleases.Inc()
	}

	return chartDir, err
}

func getRemoteChart(chartRef string) (string, error) {
	chartName, chartVersion := SplitChartRef(chartRef)
	if chartName == "" {
		return "", errors.New("no chart name provided")
	}

	dl := getChartDownloader()

	url, _, err := dl.ResolveChartVersion(chartName, chartVersion)
	if err != nil {
		return "", fmt.Errorf("cannot resolve chart version: %s", err)
	}
	_, chartFile := filepath.Split(url.Path)
	// Create versioned directory for chart, using hash of chart file to avoid special characters
	chartCacheDir := filepath.Join(os.TempDir(), chartName, Hash(chartFile))
	if err = os.MkdirAll(chartCacheDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create work directory `%s`", chartCacheDir)
	}
	// Get absolute path of the chart file
	chartPath, err := filepath.Abs(filepath.Join(chartCacheDir, chartFile))
	if err != nil {
		return "", err
	}

	// download if the chart file is not present
	if _, err := os.Stat(chartPath); err != nil {
		// download the chart file
		_, _, err = dl.DownloadTo(chartName, chartVersion, chartCacheDir)
		if err != nil {
			return "", fmt.Errorf("failed to download `%s`: %s", chartRef, err)
		}
	}

	return chartCacheDir, nil
}

func getChartDownloader() downloader.ChartDownloader {
	helmHome := helmpath.Home(os.Getenv("HELM_HOME"))
	dl := downloader.ChartDownloader{
		HelmHome: helmHome,
		Out:      os.Stdout,
		Getters: getter.All(environment.EnvSettings{
			Home: helmHome,
		}),
		Verify: downloader.VerifyIfPossible,
	}

	return dl
}

// Hash generates base64 encoded string
func Hash(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// SplitChartRef splits chart reference into chart name and chart version
func SplitChartRef(chart string) (string, string) {
	chartInfo := strings.Split(chart, ":")
	chartName, chartVersion := strings.TrimSpace(chartInfo[0]), ""
	if len(chartInfo) == 2 {
		chartVersion = strings.TrimSpace(chartInfo[1])
	}

	return chartName, chartVersion
}
