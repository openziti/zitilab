package main

import (
	"embed"
	_ "embed"
	"github.com/openziti/fablab"
	"github.com/openziti/fablab/kernel/lib/actions"
	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/lib/actions/host"
	"github.com/openziti/fablab/kernel/lib/binding"
	"github.com/openziti/fablab/kernel/lib/runlevel/0_infrastructure/aws_ssh_key"
	"github.com/openziti/fablab/kernel/lib/runlevel/0_infrastructure/semaphore"
	"github.com/openziti/fablab/kernel/lib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/lib/runlevel/1_configuration/config"
	distribution "github.com/openziti/fablab/kernel/lib/runlevel/3_distribution"
	"github.com/openziti/fablab/kernel/lib/runlevel/3_distribution/rsync"
	aws_ssh_key2 "github.com/openziti/fablab/kernel/lib/runlevel/6_disposal/aws_ssh_key"
	"github.com/openziti/fablab/kernel/lib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/resources"
	"github.com/openziti/zitilab"
	zitilib_runlevel_1_configuration "github.com/openziti/zitilab/runlevel/1_configuration"
	stallActions "github.com/openziti/zitilab/stall/actions"
	"os"
	"time"
)

//go:embed configs
var configResource embed.FS

type scaleStrategy struct{}

func (self scaleStrategy) IsScaled(entity model.Entity) bool {
	return entity.GetType() == model.EntityTypeHost && entity.GetScope().HasTag("edge-router")
}

func (self scaleStrategy) GetEntityCount(entity model.Entity) uint32 {
	if entity.GetType() == model.EntityTypeHost && entity.GetScope().HasTag("edge-router") {
		return 4
	}
	return 1
}

var m = &model.Model{
	Id: "stall",
	Scope: model.Scope{
		Defaults: model.Variables{
			"environment": "ziti-stall-test",
			"credentials": model.Variables{
				"ssh": model.Variables{
					"username": "ubuntu",
				},
				"edge": model.Variables{
					"username": "admin",
					"password": "admin",
				},
			},
		},
	},
	StructureFactories: []model.Factory{
		model.NewScaleFactoryWithDefaultEntityFactory(scaleStrategy{}),
	},
	Factories: []model.Factory{
		newStageFactory(),
	},
	Resources: model.Resources{
		resources.Configs:   resources.SubFolder(configResource, "configs"),
		resources.Binaries:  os.DirFS("/home/plorenz/go/bin"),
		resources.Terraform: resources.DefaultTerraformResources(),
	},
	Regions: model.Regions{
		"us-east-1": {
			Region: "us-east-1",
			Site:   "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope:        model.Scope{Tags: model.Tags{"ctrl"}},
					InstanceType: "c5.large",
					Components: model.Components{
						"ctrl": {
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
			},
		},
		"us-west-2": {
			Region: "us-west-2",
			Site:   "us-west-2b",
			Hosts: model.Hosts{
				"router-west-{{ .ScaleIndex }}": {
					Scope:        model.Scope{Tags: model.Tags{"edge-router"}},
					InstanceType: "c5.large",
					Components: model.Components{
						"router": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "router.yml",
							ConfigName:     "router.yml",
							PublicIdentity: "router-west-{{ .Host.ScaleIndex }}",
						},
					},
				},
				"client": {
					Scope:        model.Scope{Tags: model.Tags{"client"}},
					InstanceType: "c5.large",
				},
			},
		},
		"ap-southeast-1": {
			Region: "ap-southeast-1",
			Site:   "ap-southeast-1a",
			Hosts: model.Hosts{
				"router-ap-{{ .ScaleIndex }}": {
					Scope:        model.Scope{Tags: model.Tags{"edge-router"}},
					InstanceType: "c5.large",
					Components: model.Components{
						"router": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "router.yml",
							ConfigName:     "router.yml",
							PublicIdentity: "router-ap-{{ .Host.ScaleIndex }}",
						},
					},
				},
				"server": {
					Scope:        model.Scope{Tags: model.Tags{"server"}},
					InstanceType: "c5.large",
				},
			},
		},
	},

	Actions: model.ActionBinders{
		"bootstrap": stallActions.NewBootstrapAction(),
		"start":     stallActions.NewStartAction(),
		"stop": func(m *model.Model) model.Action {
			return component.StopInParallel("*", 15)
		},
		"syncModelEdgeState": stallActions.NewSyncModelEdgeStateAction(),
		"clean": func(m *model.Model) model.Action {
			return actions.Workflow(
				component.StopInParallel("*", 15),
				host.GroupExec("*", 25, "rm -f logs/*"),
			)
		},
	},

	Infrastructure: model.InfrastructureStages{
		aws_ssh_key.Express(),
		terraform_0.Express(),
		semaphore_0.Restart(90 * time.Second),
	},

	Configuration: model.ConfigurationStages{
		zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric()),
		config.Component(),
		zitilab.DefaultZitiBinaries(),
	},

	Distribution: model.DistributionStages{
		distribution.DistributeSshKey("*"),
		distribution.Locations("*", "logs"),
		rsync.Rsync(25),
	},

	Disposal: model.DisposalStages{
		terraform.Dispose(),
		aws_ssh_key2.Dispose(),
	},
}

func main() {
	m.AddActivationActions("stop", "bootstrap", "start", "syncModelEdgeState")
	// m.VarConfig.EnableDebugLogger()

	model.AddBootstrapExtension(&zitilab.Bootstrap{})
	model.AddBootstrapExtension(binding.AwsCredentialsLoader)
	model.AddBootstrapExtension(aws_ssh_key.KeyManager)

	fablab.InitModel(m)
	fablab.Run()
}
