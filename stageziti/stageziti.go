package stageziti

import (
	"github.com/openziti/fablab/kernel/lib/runlevel/2_kitting/devkit"
	"github.com/openziti/fablab/kernel/model"
)

func FetchZitiEdgeTunnel(version string) model.ConfigurationStage {
	return &ZitiEdgeTunnelFetchStage{
		Version: version,
	}
}

type ZitiEdgeTunnelFetchStage struct {
	Version string
	OS      string
	Arch    string
}

func (self *ZitiEdgeTunnelFetchStage) Configure(run model.Run) error {
	return devkit.FetchBinaries(func(binDir string) error {
		if self.OS == "" {
			self.OS = "linux"
		}
		if self.Arch == "" {
			self.Arch = "amd64"
		}
		panic("not yet implemented")
		//return getziti.InstallZitiEdgeTunnel(self.Version, self.OS, self.Arch, binDir, false)
	}).Configure(run)
}
