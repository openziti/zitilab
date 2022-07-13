package zitilib_runlevel_5_operation

import (
	"fmt"

	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/zitilab/actions"
	"github.com/sirupsen/logrus"
)

type echoClient struct {
	componentSpec string
}

func EchoClient(componentSpec string) model.OperatingStage {
	return &echoClient{
		componentSpec: componentSpec,
	}
}

func (ec *echoClient) Operate(run model.Run) error {
	fmt.Println("Starting echos!")
	return run.GetModel().ForEachComponent(ec.componentSpec, 1, func(c *model.Component) error {
		ssh := lib.NewSshConfigFactory(c.GetHost())
		remoteConfigFile := "/home/ubuntu/fablab/cfg/" + c.PublicIdentity + ".json"

		if err := zitilib_actions.EdgeExec(c.GetModel(), "list", "terminators"); err != nil {
			return err
		}

		echoClientCmd := fmt.Sprintf("/home/%s/fablab/bin/ziti edge tutorial ziti-echo-client --identity %s test 2>&1",
			ssh.User(), remoteConfigFile)

		if output, err := lib.RemoteExec(ssh, echoClientCmd); err != nil {
			logrus.Errorf("error starting echo client [%s] (%v)", output, err)
			return err
		} else {
			logrus.Info(output)
		}
		return nil
	})
}
