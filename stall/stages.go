package main

import (
	"fmt"
	fablib_5_operation "github.com/openziti/fablab/kernel/lib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/zitilab/models"
	zitilib_5_operation "github.com/openziti/zitilab/runlevel/5_operation"
	"strings"
	"time"
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
	//m.AddOperatingStage(zitilib_5_operation.Mesh(runPhase.GetCloser()))
	idMapper := func(id string) string {
		if id == "ctrl" {
			return "#ctrl"
		}
		id = strings.ReplaceAll(id, ".", ":")
		return "component.edgeId:" + id
	}

	m.AddOperatingStage(zitilib_5_operation.ModelMetricsWithIdMapper(runPhase.GetCloser(), idMapper))
	m.AddOperatingStage(zitilib_5_operation.CircuitMetrics(time.Second, idMapper, runPhase.GetCloser()))
	m.AddOperatingStage(clientMetrics)

	for _, host := range m.SelectHosts("*") {
		m.AddOperatingStage(fablib_5_operation.StreamSarMetrics(host, 5, 3, runPhase, cleanupPhase))
	}

	if err := self.listeners(m); err != nil {
		return fmt.Errorf("error creating listeners (%w)", err)
	}

	m.AddOperatingStage(fablib_5_operation.Timer(5*time.Second, nil))

	if err := self.dialers(m, runPhase); err != nil {
		return fmt.Errorf("error creating dialers (%w)", err)
	}

	m.AddOperatingStage(runPhase)
	m.AddOperatingStage(fablib_5_operation.Persist())

	return nil
}

func (_ *stageFactory) listeners(m *model.Model) error {
	components := m.SelectComponents(models.ServiceTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", "#loop.listener")
	}

	// only start 1 listener
	c := components[0]

	remoteConfigFile := fmt.Sprintf("/home/%v/fablab/cfg/%v.json", m.MustVariable("credentials.ssh.username"), c.PublicIdentity)
	stage := zitilib_5_operation.Loop3Listener(c.GetHost(), nil, "tcp:0.0.0.0:8171", "--config-file", remoteConfigFile)
	m.AddOperatingStage(stage)

	return nil
}

func (_ *stageFactory) dialers(m *model.Model, phase fablib_5_operation.Phase) error {
	var components []*model.Component
	components = m.SelectComponents(models.ClientTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ClientTag)
	}

	// only start 1 dialer
	c := components[0]

	remoteConfigFile := fmt.Sprintf("/home/%v/fablab/cfg/%v.json", m.MustVariable("credentials.ssh.username"), c.PublicIdentity)
	stage := zitilib_5_operation.Loop3Dialer(c.GetHost(), c.ConfigName, "tcp:test.service:8171", phase.AddJoiner(), "--config-file", remoteConfigFile, "-d")
	m.AddOperatingStage(stage)

	return nil
}

type stageFactory struct{}
