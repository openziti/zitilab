module github.com/openziti/zitilab

go 1.16

replace github.com/openziti/fablab => ../fablab

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/golang/protobuf v1.5.2
	github.com/michaelquigley/pfxlog v0.6.1
	github.com/openziti/fablab v0.3.7-0.20210521201302-477be5476d1d
	github.com/openziti/fabric v0.16.105
	github.com/openziti/foundation v0.15.75
	github.com/openziti/sdk-golang v0.15.95
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	gopkg.in/yaml.v2 v2.4.0
)
