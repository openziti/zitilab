/*
	Copyright 2019 NetFoundry, Inc.

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

package zitilab

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
)

type Bootstrap struct{}

func (bootstrap *Bootstrap) Bootstrap(m *model.Model) error {
	zitiRoot = os.Getenv("ZITI_ROOT")
	if zitiRoot == "" {
		if zitiPath, err := exec.LookPath("ziti"); err == nil {
			zitiRoot = path.Dir(path.Dir(zitiPath))
		} else {
			return fmt.Errorf("ZITI_PATH not set and ziti executable not found in path. please set 'ZITI_ROOT'")
		}
	}

	if fi, err := os.Stat(zitiRoot); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("invalid 'ZITI_ROOT' (!directory)")
		}
	} else {
		return fmt.Errorf("non-existent 'ZITI_ROOT'")
	}

	logrus.Debugf("ZITI_ROOT = [%s]", zitiRoot)

	zitiDistRoot = os.Getenv("ZITI_DIST_ROOT")
	if zitiDistRoot == "" {
		zitiDistRoot = zitiRoot
	} else {
		if fi, err := os.Stat(zitiDistRoot); err == nil {
			if !fi.IsDir() {
				return fmt.Errorf("invalid 'ZITI_DIST_ROOT' (!directory)")
			}
		} else {
			return fmt.Errorf("non-existent 'ZITI_DIST_BIN'")
		}
	}

	logrus.Debugf("ZITI_DIST_ROOT = [%s]", zitiDistRoot)

	return nil
}
