package controller

import (
	"context"
	"github.com/hchenc/iceberg/pkg/clients/clientset"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/syncer/gitlab"
	"github.com/hchenc/iceberg/pkg/syncer/harbor"
	"github.com/hchenc/iceberg/pkg/syncer/resource"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	application "github.com/hchenc/application/pkg/apis/app/v1beta1"
	iamv1alpha2 "github.com/hchenc/iceberg/pkg/apis/iam/v1alpha2"
	workspace "github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
)

const (
	RetryPeriod = 15
)

var (
	log = utils.GetLoggerEntry().WithFields(logrus.Fields{
		"component": "controller",
	})
)

var (
	reconcilerMap = make(map[string]Reconciler)

	projectGenerator     syncer.Generator
	groupGenerator       syncer.Generator
	namespaceGenerator   syncer.Generator
	applicationGenerator syncer.Generator
	userGenerator        syncer.Generator
	rolebindingGenerator syncer.Generator
	memberGenerator      syncer.Generator
	harborGenerator      syncer.Generator
	deploymentGenerator  syncer.Generator
	serviceGenerator     syncer.Generator
	volumeGenerator      syncer.Generator

	projectGeneratorService     syncer.GenerateService
	groupGeneratorService       syncer.GenerateService
	namespaceGeneratorService   syncer.GenerateService
	applicationGeneratorService syncer.GenerateService
	userGeneratorService        syncer.GenerateService
	rolebindingGeneratorService syncer.GenerateService
	memberGeneratorService      syncer.GenerateService
	harborGeneratorService      syncer.GenerateService
	deploymentGeneratorService  syncer.GenerateService
	serviceGeneratorService     syncer.GenerateService
	volumeGeneratorService      syncer.GenerateService
)

type Reconciler interface {
	SetUp(mgr manager.Manager)
}

type Reconcile func(mgr manager.Manager)

func (r Reconcile) SetUp(mgr manager.Manager) {
	r(mgr)
}

func RegisterReconciler(name string, f Reconcile) {
	reconcilerMap[name] = f
}

type Controller struct {
	Clientset *clientset.ClientSet

	ReconcilerMap map[string]Reconciler

	manager manager.Manager
}

func (c *Controller) Reconcile(ctx context.Context) {
	if err := c.manager.Start(ctx); err != nil {
		log.Fatalf("start reconciler failed for %s", err.Error())
	}
}

func NewControllerOrDie(cs *clientset.ClientSet, mgr manager.Manager) *Controller {
	c := &Controller{
		Clientset: cs,
		manager:   mgr,
	}
	c.ReconcilerMap = reconcilerMap

	runtime.Must(workspace.AddToScheme(mgr.GetScheme()))
	runtime.Must(application.AddToScheme(mgr.GetScheme()))
	runtime.Must(iamv1alpha2.AddToScheme(mgr.GetScheme()))
	runtime.Must(appsv1.AddToScheme(mgr.GetScheme()))
	runtime.Must(corev1.AddToScheme(mgr.GetScheme()))

	installGenerator(c.Clientset)
	installGeneratorService()

	for _, reconciler := range c.ReconcilerMap {
		reconciler.SetUp(mgr)
	}
	return c
}

func installGenerator(clientset *clientset.ClientSet) {
	projectGenerator = gitlab.NewGitLabProjectGenerator("", "", clientset.Ctx, clientset.GitlabClient, clientset.PagerClient)
	groupGenerator = gitlab.NewGroupGenerator("", clientset.Ctx, clientset.GitlabClient, clientset.PagerClient)
	userGenerator = gitlab.NewUserGenerator(clientset.Ctx, clientset.GitlabClient, clientset.PagerClient)
	memberGenerator = gitlab.NewMemberGenerator(clientset.Ctx, clientset.GitlabClient, clientset.PagerClient)

	namespaceGenerator = resource.NewNamespaceGenerator(clientset.Ctx, clientset.Kubeclient)
	applicationGenerator = resource.NewApplicationGenerator(clientset.Ctx, clientset.Kubeclient, clientset.AppClient)
	rolebindingGenerator = resource.NewRolebindingGenerator(clientset.Ctx, clientset.Kubeclient)
	deploymentGenerator = resource.NewDeploymentGenerator(clientset.Ctx, clientset.Kubeclient)
	serviceGenerator = resource.NewServiceGenerator(clientset.Ctx, clientset.Kubeclient)
	volumeGenerator = resource.NewVolumeGenerator(clientset.Ctx, clientset.Kubeclient)

	harborGenerator = harbor.NewHarborProjectGenerator("", "", clientset.HarborClient)
}

func installGeneratorService() {
	projectGeneratorService = syncer.NewGenerateService(projectGenerator)
	groupGeneratorService = syncer.NewGenerateService(groupGenerator)
	namespaceGeneratorService = syncer.NewGenerateService(namespaceGenerator)
	applicationGeneratorService = syncer.NewGenerateService(applicationGenerator)
	userGeneratorService = syncer.NewGenerateService(userGenerator)
	rolebindingGeneratorService = syncer.NewGenerateService(rolebindingGenerator)
	memberGeneratorService = syncer.NewGenerateService(memberGenerator)
	harborGeneratorService = syncer.NewGenerateService(harborGenerator)
	deploymentGeneratorService = syncer.NewGenerateService(deploymentGenerator)
	serviceGeneratorService = syncer.NewGenerateService(serviceGenerator)
	volumeGeneratorService = syncer.NewGenerateService(volumeGenerator)
}
