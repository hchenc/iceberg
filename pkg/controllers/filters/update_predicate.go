package filters

import "sigs.k8s.io/controller-runtime/pkg/event"

type NamespaceUpdatePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (r NamespaceUpdatePredicate) Update(e event.UpdateEvent) bool {
	namespace := e.ObjectNew.GetNamespace()

	if exists, verified := checkIndexKey(r.IncludeNamespaces, namespace); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNamespaces, namespace); verified {
		return !exists
	}

	return false

}