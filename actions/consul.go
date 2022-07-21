package zitilib_actions

import (
	"fmt"

	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

type consulStart struct {
	hostSpec     string
	consulServer string
	configDir    string
	dataPath     string
}

func StartConsul(hostSpec, consulServer, configDir, dataPath string) model.Action {
	return &consulStart{
		hostSpec:     hostSpec,
		consulServer: consulServer,
		configDir:    configDir,
		dataPath:     dataPath,
	}
}

func (cs *consulStart) Execute(m *model.Model) error {
	return m.ForEachHost(cs.hostSpec, 24, func(c *model.Host) error {
		ssh := lib.NewSshConfigFactory(c)

		cmd := fmt.Sprintf("nohup consul agent -join %s -config-dir %s -data-dir %s 2>&1 &", cs.consulServer, cs.configDir, cs.dataPath)

		if output, err := lib.RemoteExec(ssh, cmd); err != nil {
			logrus.Errorf("error starting consul service [%s] (%v)", output, err)
			return err
		}
		return nil
	})
}
