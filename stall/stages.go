package main

import (
	fablib_5_operation "github.com/openziti/fablab/kernel/lib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	zitilib_5_operation "github.com/openziti/zitilab/runlevel/5_operation"
	"strings"
)

func newStageFactory() model.Factory {
	return &stageFactory{}
}

func (self *stageFactory) Build(m *model.Model) error {
	// m.MetricsHandlers = append(m.MetricsHandlers, model.StdOutMetricsWriter{})

	runPhase := fablib_5_operation.NewPhase()
	cleanupPhase := fablib_5_operation.NewPhase()

	clientMetrics := zitilib_5_operation.NewClientMetrics("metrics", runPhase.GetCloser())
	m.AddActivationStage(clientMetrics)

	m.AddOperatingActions("syncModelEdgeState")
	m.AddOperatingStage(fablib_5_operation.InfluxMetricsReporter())
	m.AddOperatingStage(zitilib_5_operation.Mesh(runPhase.GetCloser()))
	m.AddOperatingStage(zitilib_5_operation.ModelMetricsWithIdMapper(runPhase.GetCloser(), func(id string) string {
		if id == "ctrl" {
			return "#ctrl"
		}
		id = strings.ReplaceAll(id, ".", ":")
		return "component.edgeId:" + id
	}))
	m.AddOperatingStage(clientMetrics)

	for _, host := range m.SelectHosts("*") {
		m.AddOperatingStage(fablib_5_operation.StreamSarMetrics(host, 5, 3, runPhase, cleanupPhase))
	}

	m.AddOperatingStage(runPhase)
	m.AddOperatingStage(fablib_5_operation.Persist())

	return nil
}

type stageFactory struct{}
