package filters

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"strings"
)
//TODO
type NamespaceDeletePredicate struct {

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
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else {
		return true
	}
}
func (r NamespaceDeletePredicate) Generic(e event.GenericEvent) bool {
	return false
}