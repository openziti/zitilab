/*
	Copyright 2020 NetFoundry Inc.

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

package main

import (
	"github.com/openziti/fablab/kernel/lib/actions"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/zitilab/actions/edge"
	"github.com/openziti/zitilab/models"
)

func NewSyncModelEdgeStateAction() model.ActionBinder {
	action := &syncModelEdgeStateAction{}
	return action.bind
}

func (a *syncModelEdgeStateAction) bind(*model.Model) model.Action {
	workflow := actions.Workflow()
	workflow.AddAction(edge.Login("#ctrl"))
	workflow.AddAction(edge.SyncModelEdgeState(models.EdgeRouterTag))
	return workflow
}

type syncModelEdgeStateAction struct{}
