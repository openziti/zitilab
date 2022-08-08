/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package zitilib_runlevel_1_configuration

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/ziti/ziti/cmd/ziti/cmd"
	"github.com/sirupsen/logrus"
)

func generateCa() error {
	var pkiOut bytes.Buffer
	var pkiErr bytes.Buffer
	ziticli := cmd.NewRootCommand(nil, &pkiOut, &pkiErr)

	ziticli.SetArgs([]string{"pki", "create", "ca", "--pki-root", model.PkiBuild(), "--ca-name", "root", "--ca-file", "root"})
	logrus.Infof("%v", ziticli.Args)
	if err := ziticli.Execute(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	pkiOut.Reset()
	pkiErr.Reset()
	ziticli.SetArgs([]string{"pki", "create", "intermediate", "--pki-root", model.PkiBuild(), "--ca-name", "root"})
	logrus.Infof("%v", ziticli.Args)
	if err := ziticli.Execute(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	return nil
}

func generateCert(name, ip string) error {
	logrus.Infof("generating certificate [%s:%s]", name, ip)
	logrus.Infof("generating certificate key")

	var pkiOut bytes.Buffer
	var pkiErr bytes.Buffer
	ziticli := cmd.NewRootCommand(nil, &pkiOut, &pkiErr)

	ziticli.SetArgs([]string{"pki", "create", "key", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--key-file", name})
	logrus.Infof("%v", ziticli.Args)
	if err := ziticli.Execute(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	logrus.Infof("generating certificate server")
	pkiOut.Reset()
	pkiErr.Reset()
	ziticli.SetArgs([]string{"pki", "create", "server", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--server-file", fmt.Sprintf("%s-server", name), "--ip", ip, "--key-file", name})
	logrus.Infof("%v", ziticli.Args)
	if err := ziticli.Execute(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating server certificate (%s)", err)
	}

	logrus.Infof("generating certificate client")
	pkiOut.Reset()
	pkiErr.Reset()
	ziticli.SetArgs([]string{"pki", "create", "client", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--client-file", fmt.Sprintf("%s-client", name), "--key-file", name, "--client-name", name})
	logrus.Infof("%v", ziticli.Args)
	if err := ziticli.Execute(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating client certificate (%s)", err)
	}

	return nil
}
