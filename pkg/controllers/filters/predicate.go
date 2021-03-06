package filters

import (
	"strings"
)

var (
	DefaultIncludeNamespaces = []string{
		"fat",
		"sit",
		"uat",
	}
	DefaultExcludeNamespaces = []string{
		"system",
		"kube",
	}
	DefaultExcludeNames = []string{
		"system",
		"admin",
	}
)

func checkIndexKey(array []string, indexKey string) (bool, bool) {
	exists, verified := false, false

	if len(array) != 0 {
		verified = true
		for _, ns := range array {
			exists = exists || strings.Contains(indexKey, ns)
		}
		return exists, verified
	}
	return exists, verified
}

func checkLabels(labels map[string]string, target map[string]string, force bool) bool {
	result := false

	for k, v := range target {
		if value, exists := labels[k]; exists {
			if force {
				result = result || exists && (value == v)
			} else {
				result = result || exists
			}

		}
	}
	return result
}
