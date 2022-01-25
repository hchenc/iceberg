package filters

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespaceDeletePredicate struct {
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (r NamespaceDeletePredicate) Create(e event.CreateEvent) bool {
	return false
}
func (r NamespaceDeletePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
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
func (r NamespaceDeletePredicate) Generic(e event.GenericEvent) bool {
	return false
}

type NameDeletePredicate struct {
	//include namespaces has higher priority
	IncludeNames []string
	ExcludeNames []string
}

func (r NameDeletePredicate) Create(e event.CreateEvent) bool {
	return false
}
func (r NameDeletePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
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
func (r NameDeletePredicate) Generic(e event.GenericEvent) bool {
	return false
}
