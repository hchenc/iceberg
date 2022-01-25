package filters

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespaceCreatePredicate struct {
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (r NamespaceCreatePredicate) Create(e event.CreateEvent) bool {
	name := e.Object.GetNamespace()

	if exists, verified := checkIndexKey(r.IncludeNamespaces, name); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNamespaces, name); verified {
		return !exists
	}

	return false
}
func (r NamespaceCreatePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (r NamespaceCreatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}
func (r NamespaceCreatePredicate) Generic(e event.GenericEvent) bool {
	return false
}

type NameCreatePredicate struct {
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
func (r NameCreatePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (r NameCreatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}
func (r NameCreatePredicate) Generic(e event.GenericEvent) bool {
	return false
}

type LabelCreatePredicate struct {
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
func (d LabelCreatePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (d LabelCreatePredicate) Delete(e event.DeleteEvent) bool {
	return false

}
func (d LabelCreatePredicate) Generic(e event.GenericEvent) bool {
	return false
}
