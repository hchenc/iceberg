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
	var result bool

	if len(array) != 0 {
		for _, ns := range array {
			result = result || strings.Contains(indexKey, ns)
		}
		return result, true
	}
	return false, false
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
