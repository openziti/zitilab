package zitilib_runlevel_5_operation

import (
	"github.com/openziti/channel"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fabric/pb/mgmt_pb"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"time"
)

func CircuitMetrics(closer chan struct{}) model.OperatingStage {
	return &modelMetrics{
		closer: closer,
	}
}

type circuitMetrics struct {
	ch       channel.Channel
	m        *model.Model
	closer   <-chan struct{}
	circuits cmap.ConcurrentMap
}

func (metrics *circuitMetrics) Operate(run model.Run) error {
	return nil
}

func (metrics *circuitMetrics) runMetrics() {
	logrus.Infof("starting")
	defer logrus.Infof("exiting")

	<-metrics.closer
	_ = metrics.ch.Close()
}

func (metrics *circuitMetrics) toModelMetricsEvent(fabricEvent *mgmt_pb.StreamMetricsEvent) *model.MetricsEvent {
	modelEvent := &model.MetricsEvent{
		Timestamp: time.Unix(fabricEvent.Timestamp.Seconds, int64(fabricEvent.Timestamp.Nanos)),
		Metrics:   model.MetricSet{},
	}

	for name, val := range fabricEvent.IntMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	for name, val := range fabricEvent.FloatMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	return modelEvent
}
