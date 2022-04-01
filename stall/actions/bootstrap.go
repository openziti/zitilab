/*
	Copyright 2020 NetFoundry, Inc.

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

package actions

import (
	"github.com/openziti/fablab/kernel/lib/actions"
	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/lib/actions/host"
	"github.com/openziti/fablab/kernel/lib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/zitilab/actions"
	"github.com/openziti/zitilab/actions/edge"
	"github.com/openziti/zitilab/models"
	"time"
)

func NewBootstrapAction() model.ActionBinder {
	action := &bootstrapAction{}
	return action.bind
}

func (a *bootstrapAction) bind(m *model.Model) model.Action {
	//sshUsername := m.MustStringVariable("credentials.ssh.username")

	workflow := actions.Workflow()

	workflow.AddAction(host.GroupExec("*", 25, "rm -f logs/*"))
	workflow.AddAction(component.Stop("#ctrl"))
	workflow.AddAction(edge.InitController("#ctrl"))
	workflow.AddAction(component.Start("#ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	//workflow.AddAction(host.GroupExec("*", 25,
	//	fmt.Sprintf("mkdir -p /home/%s/.ziti", sshUsername),
	//	fmt.Sprintf("rm -f /home/%s/.ziti/identities.yml", sshUsername),
	//	fmt.Sprintf("ln -s /home/%s/fablab/cfg/remote_identities.yml /home/%s/.ziti/identities.yml", sshUsername, sshUsername),
	//))

	// HasControllerComponent = "component.ctrl"
	workflow.AddAction(edge.Login("component#ctrl"))

	// EdgeRouterTag = ".edge-router"
	workflow.AddAction(component.StopInParallel("host.edge-router", 25))
	workflow.AddAction(edge.InitEdgeRouters("host.edge-router", 2))
	//workflow.AddAction(edge.InitIdentities(models.SdkAppTag, 2))

	workflow.AddAction(zitilib_actions.Edge("create", "service", "perf-test", "--encryption", "on"))
	workflow.AddAction(zitilib_actions.Edge("create", "service", "metrics", "--encryption", "on"))

	workflow.AddAction(zitilib_actions.Fabric("create", "service", "perf-proxy"))

	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "perf-bind", "Bind", "--service-roles", "@perf-test", "--identity-roles", "#service"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "perf-dial", "Dial", "--service-roles", "@perf-test", "--identity-roles", "#client"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "metrics-dial", "Dial", "--service-roles", "@metrics", "--identity-roles", "#client"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "metrics-bind", "Bind", "--service-roles", "@metrics", "--identity-roles", "#metrics-host"))

	workflow.AddAction(zitilib_actions.Edge("create", "edge-router-policy", "client-routers", "--edge-router-roles", "#initiator", "--identity-roles", "#client"))
	workflow.AddAction(zitilib_actions.Edge("create", "edge-router-policy", "server-routers", "--edge-router-roles", "#terminator", "--identity-roles", "#service"))

	workflow.AddAction(zitilib_actions.Edge("create", "edge-router-policy", "metrics-routers", "--edge-router-roles", "#metrics", "--identity-roles", "#all"))

	workflow.AddAction(zitilib_actions.Edge("create", "service-edge-router-policy", "perf-test", "--semantic", "AnyOf", "--service-roles", "@perf-test", "--edge-router-roles", "#initiator,#terminator"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-edge-router-policy", "metrics", "--service-roles", "@metrics", "--edge-router-roles", "#metrics"))

	workflow.AddAction(model.ActionFunc(func(m *model.Model) error {
		createTerminatorsWf := actions.Workflow()
		for _, terminator := range m.SelectComponents(".edge-router.terminator") {
			routerId, err := edge.GetEntityId(m, "edge-routers", terminator.PublicIdentity)
			if err != nil {
				panic(err)
			}
			for _, host := range m.SelectHosts("component.service") {
				createTerminatorsWf.AddAction(zitilib_actions.Edge("create", "terminator", "perf-test", terminator.PublicIdentity, "tcp:"+host.PrivateIp+":8171", "--binding", "transport"))
				createTerminatorsWf.AddAction(zitilib_actions.Fabric("create", "terminator", "perf-proxy", routerId, "tcp:"+host.PrivateIp+":8171", "--binding", "transport"))
			}
		}
		return createTerminatorsWf.Execute(m)
	}))

	workflow.AddAction(component.Stop(models.ControllerTag))

	return workflow
}

type bootstrapAction struct{}
