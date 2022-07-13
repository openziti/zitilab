package echo

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/zitilab/actions"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type echoServerStart struct {
	componentSpec string
}

func StartEchoServers(componentSpec string) model.Action {
	return &echoServerStart{
		componentSpec: componentSpec,
	}
}

func (esi *echoServerStart) Execute(m *model.Model) error {
	return m.ForEachComponent(esi.componentSpec, 1, func(c *model.Component) error {
		ssh := lib.NewSshConfigFactory(c.GetHost())
		remoteConfigFile := "/home/ubuntu/fablab/cfg/" + c.PublicIdentity + ".json"

		if out, err := zitilib_actions.EdgeExecWithOutput(c.GetModel(), "policy-advisor", "identities", "-q", "echo-client", "echo"); err != nil {
			return err
		} else {
			logrus.Error(out)
		}

		if out, err := zitilib_actions.EdgeExecWithOutput(c.GetModel(), "list", "service-policies", "name contains \"echo\""); err != nil {
			return err
		} else {
			logrus.Error(out)
		}
		if out, err := zitilib_actions.EdgeExecWithOutput(c.GetModel(), "policy-advisor", "identities", "-q", "echo-server", "echo"); err != nil {
			return err
		} else {
			logrus.Error(out)
		}

		//if out, err := zitilib_actions.FabricExecWithOutput(c.GetModel(), "list", "routers"); err != nil {
		//	return err
		//} else {
		//	logrus.Error(out)
		//}
		//logFile := fmt.Sprintf("/home/%s/logs/echo-server.log", ssh.User())
		echoServerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-echo --identity %s 2>&1 &",
			ssh.User(), remoteConfigFile)

		//echoServerCmd = "echo test &"

		//if _, err := zitilib_cli.Exec(c.GetModel(), "edge", "tutorial", "ziti-echo-server", "--identity", remoteConfigFile); err != nil {
		//	return err
		//}
		//if output, err := lib.RemoteExec(ssh, fmt.Sprintf("ls /home")); err != nil {
		//	logrus.Errorf("error ls [%s] (%v)", output, err)
		//	return err
		//}
		if output, err := RemoteExec(ssh, echoServerCmd); err != nil {
			logrus.Errorf("error starting echo server [%s] (%v)", output, err)
			return err
		}
		logrus.Error("After remote starting echo server")
		return nil
	})
}

type testAction struct{}

func test() model.Action {
	return &testAction{}
}

func (t *testAction) Execute(m *model.Model) error {
	return m.ForEachComponent("#echo-server", 1, func(c *model.Component) error {
		sshConfigFactory := lib.NewSshConfigFactory(c.GetHost())

		sudoCmd := ""
		if true {
			sudoCmd = " sudo "
		}
		serviceCmd := fmt.Sprintf("echo test &",
			sudoCmd, sshConfigFactory.User(), c.BinaryName, sshConfigFactory.User(), c.ConfigName)
		if value, err := RemoteExec(sshConfigFactory, serviceCmd); err == nil {
			if len(value) > 0 {
				logrus.Infof("output [%s]", strings.Trim(value, " \t\r\n"))
			}
		} else {
			return fmt.Errorf("error starting component [%s] on [%s] [%s] (%s)", c.BinaryName, c.GetHost().PublicIp, value, err)
		}
		return nil
		if err := lib.LaunchService(sshConfigFactory, c.BinaryName, c.ConfigName, c.RunWithSudo); err != nil {
			return fmt.Errorf("error starting component [%s] on [%s] (%s)", c.BinaryName, c.GetHost().PublicIp, err)
		}
		return nil
	})
}

func RemoteExec(sshConfig lib.SshConfigFactory, cmd string) (string, error) {
	return RemoteExecAll(sshConfig, cmd)
}

func RemoteExecAll(sshConfig lib.SshConfigFactory, cmds ...string) (string, error) {
	var b bytes.Buffer
	err := RemoteExecAllTo(sshConfig, &b, cmds...)
	return b.String(), err
}

func RemoteExecAllTo(sshConfig lib.SshConfigFactory, out io.Writer, cmds ...string) error {
	if len(cmds) == 0 {
		return nil
	}
	config := sshConfig.Config()

	logrus.Infof("executing [%s]: '%s'", sshConfig.Address(), cmds[0])

	client, err := ssh.Dial("tcp", sshConfig.Address(), config)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	for idx, cmd := range cmds {
		session, err := client.NewSession()
		if err != nil {
			return err
		}
		session.Stdout = out

		if idx > 0 {
			logrus.Infof("executing [%s]: '%s'", sshConfig.Address(), cmd)
		}
		err = session.Run(cmd)
		_ = session.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
