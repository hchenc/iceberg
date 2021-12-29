package filters

import (
	"github.com/hchenc/iceberg/pkg/constants"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"strings"
)

var (
	DefaultIncludeNamespaces = []string{
		constants.FAT,
		constants.SIT,
		constants.UAT,
	}
	DefaultExcludeNamespaces = []string{
		"system",
		"kube",
	}
)
//TODO
type NamespacePredicate struct {

	ExcludeNamespaces []string
	IncludeNamespaces []string
}

func (n NamespacePredicate) filter(namespace string) (string, bool) {
	//include namespace have higher priority
	for _, ns := range n.IncludeNamespaces{
		if strings.Contains(namespace, ns) {
			return "",true
		}
	}

	for _, ns := range n.ExcludeNamespaces{
		if strings.Contains(namespace, ns) {
			return "",false
		}
	}
	return namespace, false
}


func (n NamespacePredicate) Create(e event.CreateEvent) bool {
	name := e.Object.GetNamespace()
	if ns, result := n.filter(name); len(ns) == 0{
		return result
	} else {
		return false
	}
}
func (n NamespacePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (n NamespacePredicate) Delete(e event.DeleteEvent) bool {
	name := e.Object.GetName()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else {
		return true
	}
}
func (n NamespacePredicate) Generic(e event.GenericEvent) bool {
	return false
}