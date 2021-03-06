package app

import (
	"context"
	"fmt"
	"github.com/hchenc/iceberg/cmd/controller-manager/app/options"
	"github.com/hchenc/iceberg/pkg/clients/clientset"
	"github.com/hchenc/iceberg/pkg/config"
	"github.com/hchenc/iceberg/pkg/constants"
	icecontroller "github.com/hchenc/iceberg/pkg/controllers"
	"github.com/hchenc/iceberg/pkg/utils/term"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewControllerManagerCommandOptions() *cobra.Command {

	s := options.NewControllerManagerConfigOptions()
	conf, err := config.TryLoadFromDisk()
	if err == nil {
		s = &options.ControllerManagerConfig{
			KubeOptions:      s.KubeOptions,
			HarborOptions:    conf.HarborOptions,
			GitlabOptions:    conf.GitlabOptions,
			IntegrateOptions: conf.IntegrateOptions,
			LeaderElect:      s.LeaderElect,
			LeaderElection:   s.LeaderElection,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	cmd := &cobra.Command{
		Use:   "controller-manager",
		Short: "",
		Long:  "Iceberg controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err = run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	return cmd
}

func run(s *options.ControllerManagerConfig, ctx context.Context) error {
	scheme := runtime.NewScheme()

	mgrOptions := manager.Options{
		Scheme: scheme,
		Port:   9443,
	}
	if s.LeaderElect {
		mgrOptions.LeaderElection = s.LeaderElect
		mgrOptions.LeaderElectionNamespace = constants.DevopsNamespace
		mgrOptions.LeaderElectionID = "iceberg-controller-manager-leader-election"
		mgrOptions.LeaseDuration = &s.LeaderElection.LeaseDuration
		mgrOptions.RetryPeriod = &s.LeaderElection.RetryPeriod
		mgrOptions.RenewDeadline = &s.LeaderElection.RenewDeadline
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())
	// Use 8443 instead of 443 cause we need root permission to bind port 443
	mgr, err := manager.New(s.KubeOptions.KubeConfig, mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}
	cs := clientset.NewClientSetForControllerManagerConfigOptions(s)
	controller := icecontroller.NewControllerOrDie(cs, mgr)

	if err = controller.Reconcile(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}
	return nil
}
