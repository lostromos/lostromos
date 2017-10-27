package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// CreatedReleases is a metric for the number of releases created by this operator
	CreatedReleases = prometheus.NewCounter(prometheus.CounterOpts{
		Help:      "The number of create processed by this operator",
		Name:      "created_count",
		Namespace: "releases",
	})

	// DeletedReleases is a metric for the number of releases deleted by this operator
	DeletedReleases = prometheus.NewCounter(prometheus.CounterOpts{
		Help:      "The number of delete requests processed by this operator",
		Name:      "deleted_count",
		Namespace: "releases",
	})

	// ManagedReleases is a metric of the number of current releases managed by the operator
	ManagedReleases = prometheus.NewGauge(prometheus.GaugeOpts{
		Help:      "The number of releases managed by this operator",
		Name:      "managed_total",
		Namespace: "releases",
	})

	// UpdatedReleases is a metric for the number of releases updated by this operator
	UpdatedReleases = prometheus.NewCounter(prometheus.CounterOpts{
		Help:      "The number of updates processed by this operator",
		Name:      "updated_count",
		Namespace: "releases",
	})
)

func init() {
	prometheus.MustRegister(CreatedReleases)
	prometheus.MustRegister(DeletedReleases)
	prometheus.MustRegister(ManagedReleases)
	prometheus.MustRegister(UpdatedReleases)
}
