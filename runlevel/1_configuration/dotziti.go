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
	_ "embed"
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/local_identities.yml
var localIdentitiesTemplate string

func DotZiti() model.ConfigurationStage {
	return &dotZiti{}
}

func (d *dotZiti) Configure(run model.Run) error {
	m := run.GetModel()
	if err := generateCert("dotziti", "127.0.0.1"); err != nil {
		return fmt.Errorf("error generating cert for [dotziti] (%s)", err)
	}
	if err := generateLocalIdentities(m); err != nil {
		return fmt.Errorf("error generating local identities for [dotziti] (%s)", err)
	}
	if err := mergeLocalIdentities(); err != nil {
		return fmt.Errorf("error merging local identities for [dotziti] (%s)", err)
	}
	return nil
}

type dotZiti struct {
}

func generateLocalIdentities(m *model.Model) error {
	t, err := template.New("config").Funcs(lib.TemplateFuncMap(m)).Parse(localIdentitiesTemplate)
	if err != nil {
		return errors.Wrap(err, "error parsing embedded template [local_identities.yml]")
	}

	outputPath := filepath.Join(model.PkiBuild(), "local_identities.yml")
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories [%s] (%s)", outputPath, err)
	}

	outputF, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating config [%s] (%s)", outputPath, err)
	}
	defer func() { _ = outputF.Close() }()

	err = t.Execute(outputF, struct {
		RunPath string
	}{
		RunPath: model.ActiveInstancePath(),
	})
	if err != nil {
		return fmt.Errorf("error rendering template [%s] (%s)", outputPath, err)
	}

	logrus.Infof("config => [local_identities.yml]")

	return nil
}

func mergeLocalIdentities() error {
	home := os.Getenv("ZITI_HOME")

	var err error
	idPath := ""
	if home != "" {
		idPath = filepath.Join(home, "identities.yml")
	} else {
		home, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory (%s)", err)
		}
		idPath = filepath.Join(home, ".ziti/identities.yml")
	}

	var identities map[interface{}]interface{}
	_, err = os.Stat(idPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(idPath), os.ModePerm); err != nil {
				return fmt.Errorf("error making parent directory [%s] (%s)", filepath.Dir(idPath), err)
			}
			identities = make(map[interface{}]interface{})
		}
	} else {
		data, err := os.ReadFile(idPath)
		if err != nil {
			return fmt.Errorf("unable to read existing identities [%s] (%s)", idPath, err)
		}

		err = yaml.Unmarshal(data, &identities)
		if err != nil {
			return fmt.Errorf("error unmarshaling existing identities [%s] (%s)", idPath, err)
		}
	}

	var localIdentities map[interface{}]interface{}
	localIdPath := filepath.Join(model.PkiBuild(), "local_identities.yml")
	data, err := os.ReadFile(localIdPath)
	if err != nil {
		return fmt.Errorf("error reading local identities [%s] (%s)", localIdPath, err)
	}

	err = yaml.Unmarshal(data, &localIdentities)
	if err != nil {
		return fmt.Errorf("error unmarshalling local identities [%s] (%s)", localIdPath, err)
	}

	fablabIdentity, found := localIdentities["default"]
	if !found {
		return fmt.Errorf("no 'default' identity in local identities [%s] (%s)", localIdPath, err)
	}

	identities[model.ActiveInstanceId()] = fablabIdentity
	data, err = yaml.Marshal(identities)
	if err != nil {
		return errors.Wrap(err, "error marshalling identities to yaml")
	}
	if err := os.WriteFile(idPath, data, os.ModePerm); err != nil {
		return fmt.Errorf("error writing user identities [%s] (%s)", idPath, err)
	}

	return nil
}
