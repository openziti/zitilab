module github.com/openziti/zitilab

go 1.16

replace github.com/openziti/fablab => ../fablab

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/antlr/antlr4 v0.0.0-20210114010855-d34d2e1c271a // indirect
	github.com/golang/protobuf v1.5.2
	github.com/lucas-clemente/quic-go v0.18.1 // indirect
	github.com/michaelquigley/pfxlog v0.6.9
	github.com/openziti/channel v0.18.23 // indirect
	github.com/openziti/fablab v0.3.7-0.20210521201302-477be5476d1d
	github.com/openziti/fabric v0.17.96
	github.com/openziti/foundation v0.17.22
	github.com/openziti/sdk-golang v0.16.44
	github.com/openziti/transport v0.1.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	gopkg.in/yaml.v2 v2.4.0
)
