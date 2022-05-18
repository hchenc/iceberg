package filters

import (
	"github.com/hchenc/iceberg/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespaceCreatePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (r NamespaceCreatePredicate) Create(e event.CreateEvent) bool {
	namespace := e.Object.GetNamespace()

	if exists, verified := checkIndexKey(r.IncludeNamespaces, namespace); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNamespaces, namespace); verified {
		return !exists
	}

	return false
}


type NameCreatePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNames []string
	ExcludeNames []string
}

func (r NameCreatePredicate) Create(e event.CreateEvent) bool {
	name := e.Object.GetName()

	if exists, verified := checkIndexKey(r.IncludeNames, name); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNames, name); verified {
		return !exists
	}

	return false
}

type LabelCreatePredicate struct {
	filterPredicate
	Force         bool
	IncludeLabels map[string]string
	ExcludeLabels map[string]string
}

func (d LabelCreatePredicate) Create(e event.CreateEvent) bool {
	labels := e.Object.GetLabels()

	if result := checkLabels(labels, d.IncludeLabels, d.Force); result {
		return result
	}

	if result := checkLabels(labels, d.ExcludeLabels, d.Force); result {
		return !result
	}
	return false

}

type SecretCreatePredicate struct {
	filterPredicate
}

func (s SecretCreatePredicate) Create(e event.CreateEvent) bool {
	secret := e.Object.(*v1.Secret)
	if secret.Type != constants.AllowedSecretType {
		return false
	}
	return true
}