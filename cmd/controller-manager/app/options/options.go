package options

import (
	"flag"
	"github.com/hchenc/iceberg/pkg/config"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"strings"
	"time"
)

type ControllerManagerConfig struct {
	KubeOptions      *config.KubernetesOptions
	HarborOptions    *config.HarborOptions
	GitlabOptions    *config.GitlabOptions
	IntegrateOptions []*config.IntegrateOption
	LeaderElect      bool
	LeaderElection   *leaderelection.LeaderElectionConfig
}

func NewControllerManagerConfigOptions() *ControllerManagerConfig {

	return &ControllerManagerConfig{
		KubeOptions:      config.NewKubernetesConfig(),
		HarborOptions:    nil,
		GitlabOptions:    nil,
		IntegrateOptions: nil,
		LeaderElect:      false,
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
	}
}

func (c *ControllerManagerConfig) Validate() []error {
	var errs []error
	errs = append(errs, c.KubeOptions.Validate()...)
	errs = append(errs, c.GitlabOptions.Validate()...)
	errs = append(errs, c.HarborOptions.Validate()...)

	for _, ingegrateOption := range c.IntegrateOptions {
		errs = append(errs, ingegrateOption.Validate()...)
	}
	return errs
}

func (c *ControllerManagerConfig) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	c.KubeOptions.AddFlags(fss.FlagSet("kubernetes"), c.KubeOptions)
	fs := fss.FlagSet("leaderelection")

	fs.BoolVar(&c.LeaderElect, "leader-elect", c.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager "+
		"deployed with multiple replicas.")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}
