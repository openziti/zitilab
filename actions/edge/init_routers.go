package edge

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/lib/actions/host"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/zitilab/cli"
	"path/filepath"
	"strings"
)

func InitEdgeRouters(componentSpec string, concurrency int) model.Action {
	return &initEdgeRoutersAction{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (action *initEdgeRoutersAction) Execute(m *model.Model) error {
	return m.ForEachComponent(action.componentSpec, action.concurrency, func(c *model.Component) error {
		if _, err := cli.Exec(m, "edge", "delete", "edge-router", c.PublicIdentity); err != nil {
			return err
		}

		return action.createAndEnrollRouter(c)
	})
}

func (action *initEdgeRoutersAction) createAndEnrollRouter(c *model.Component) error {
	ssh := lib.NewSshConfigFactory(c.GetHost())

	jwtFileName := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".jwt")

	_, err := cli.Exec(c.GetModel(), "edge", "create", "edge-router", c.PublicIdentity, "-j",
		"--jwt-output-file", jwtFileName,
		"-a", strings.Join(c.Tags, ","))

	if err != nil {
		return err
	}

	remoteJwt := "/home/ubuntu/fablab/cfg/" + c.PublicIdentity + ".jwt"
	if err := lib.SendFile(ssh, jwtFileName, remoteJwt); err != nil {
		return err
	}

	tmpl := "set -o pipefail; /home/ubuntu/fablab/bin/%s enroll /home/ubuntu/fablab/cfg/%s -j %s 2>&1 | tee /home/ubuntu/logs/%s.router.enroll.log "
	return host.Exec(c.GetHost(),
		"mkdir -p /home/ubuntu/logs",
		fmt.Sprintf(tmpl, c.BinaryName, c.ConfigName, remoteJwt, c.ConfigName)).Execute(c.GetModel())
}

type initEdgeRoutersAction struct {
	componentSpec string
	concurrency   int
}
