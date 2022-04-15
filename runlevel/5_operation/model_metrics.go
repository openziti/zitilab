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

package zitilib_runlevel_5_operation

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/openziti/channel"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/ziti/ziti/cmd/ziti/cmd/api"
	"github.com/sirupsen/logrus"
	"time"
)

func ModelMetrics(closer <-chan struct{}) model.OperatingStage {
	return MetricsWithIdMapper(closer, func(id string) string {
		return "#" + id
	})
}

func ModelMetricsWithIdMapper(closer <-chan struct{}, f func(string) string) model.OperatingStage {
	return &modelMetrics{
		closer:             closer,
		idToSelectorMapper: f,
	}
}

type modelMetrics struct {
	ch                 channel.Channel
	m                  *model.Model
	closer             <-chan struct{}
	idToSelectorMapper func(string) string
}

func (self *modelMetrics) Operate(run model.Run) error {
	bindHandler := channel.BindHandlerF(func(binding channel.Binding) error {
		binding.AddTypedReceiveHandler(self)
		return nil
	})

	var err error
	self.ch, err = api.NewWsMgmtChannel(channel.BindHandlerF(bindHandler))
	if err != nil {
		return err
	}

	request := &mgmt_pb.StreamMetricsRequest{
		Matchers: []*mgmt_pb.StreamMetricsRequest_MetricMatcher{},
	}
	body, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling metrics request (%w)", err)
	}

	requestMsg := channel.NewMessage(int32(mgmt_pb.ContentType_StreamMetricsRequestType), body)
	err = requestMsg.WithTimeout(5 * time.Second).SendAndWaitForWire(self.ch)
	if err != nil {
		logrus.Fatalf("error queuing metrics request (%v)", err)
	}

	self.m = run.GetModel()
	go self.runMetrics()

	return nil
}

func (self *modelMetrics) ContentType() int32 {
	return int32(mgmt_pb.ContentType_StreamMetricsEventType)
}

func (self *modelMetrics) HandleReceive(msg *channel.Message, _ channel.Channel) {
	response := &mgmt_pb.StreamMetricsEvent{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Error("error handling metrics receive (%w)", err)
	}

	hostSelector := self.idToSelectorMapper(response.SourceId)
	host, err := self.m.SelectHost(hostSelector)
	if err == nil {
		modelEvent := self.toModelMetricsEvent(response)
		self.m.AcceptHostMetrics(host, modelEvent)
		logrus.Infof("<$= [%s]", response.SourceId)
	} else {
		logrus.Errorf("unable to find host (%v)", err)
	}
}

func (self *modelMetrics) runMetrics() {
	logrus.Infof("starting")
	defer logrus.Infof("exiting")

	<-self.closer
	_ = self.ch.Close()
}

func (self *modelMetrics) toModelMetricsEvent(fabricEvent *mgmt_pb.StreamMetricsEvent) *model.MetricsEvent {
	modelEvent := &model.MetricsEvent{
		Timestamp: time.Unix(fabricEvent.Timestamp.Seconds, int64(fabricEvent.Timestamp.Nanos)),
		Metrics:   model.MetricSet{},
		Tags:      fabricEvent.Tags,
	}

	for name, val := range fabricEvent.IntMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	for name, val := range fabricEvent.FloatMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	return modelEvent
}
