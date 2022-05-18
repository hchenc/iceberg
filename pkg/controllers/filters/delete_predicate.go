package filters

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespaceDeletePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}
func (r NamespaceDeletePredicate) Delete(e event.DeleteEvent) bool {

	name := e.Object.GetName()

	if exists, verified := checkIndexKey(r.IncludeNamespaces, name); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNamespaces, name); verified {
		return !exists
	}

	return false
}

type NameDeletePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNames []string
	ExcludeNames []string
}

func (r NameDeletePredicate) Delete(e event.DeleteEvent) bool {
	name := e.Object.GetName()

	if exists, verified := checkIndexKey(r.IncludeNames, name); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNames, name); verified {
		return !exists
	}
	return false
}
