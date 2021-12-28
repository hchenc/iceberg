package clientset

import (
	"context"
	"github.com/hchenc/application/pkg/client/clientset/versioned"
	"github.com/hchenc/iceberg/cmd/controller-manager/app/options"
	versioned2 "github.com/hchenc/pager/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type ClientSet struct {
	Ctx context.Context

	Kubeclient *kubernetes.Clientset

	AppClient *versioned.Clientset

	PagerClient *versioned2.Clientset

	GitlabClient *GitlabClient

	HarborClient *HarborClient
}

func NewClientSetForControllerManagerConfigOptions(conf *options.ControllerManagerConfig) *ClientSet {

	var cs ClientSet

	cs.Ctx = context.Background()

	cs.Kubeclient = kubernetes.NewForConfigOrDie(conf.KubeOptions.KubeConfig)

	cs.AppClient = versioned.NewForConfigOrDie(conf.KubeOptions.KubeConfig)

	cs.PagerClient = versioned2.NewForConfigOrDie(conf.KubeOptions.KubeConfig)

	cs.GitlabClient = NewGitlabClient(conf.GitlabOptions, conf.IntegrateOptions)

	cs.HarborClient = NewHarborClient(conf.HarborOptions, cs.Ctx)

	return &cs
}
