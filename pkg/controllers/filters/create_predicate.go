package filters

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"strings"
)

//TODO
type NamespaceCreatePredicate struct {
}

func (r NamespaceCreatePredicate) Create(e event.CreateEvent) bool {
	name := e.Meta.GetNamespace()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else if strings.Contains(name, "sit") || strings.Contains(name, "fat") || strings.Contains(name, "uat") {
		return true
	} else {
		return false
	}
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
