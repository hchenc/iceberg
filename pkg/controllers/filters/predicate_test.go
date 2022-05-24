package filters

import (
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNamespacesFilter(t *testing.T) {
	var nonStrings []string
	var name string
	var result, exists bool

	name = "sales-fat"
	result, exists = checkIndexKey(DefaultIncludeNamespaces, name)
	assert.Equal(t, exists, true)
	assert.Equal(t, result, true)

	name = "sales-sit"
	result, exists = checkIndexKey(DefaultIncludeNamespaces, name)
	assert.Equal(t, exists, true)
	assert.Equal(t, result, true)

	name = "sales-uat"
	result, exists = checkIndexKey(DefaultIncludeNamespaces, name)
	assert.Equal(t, exists, true)
	assert.Equal(t, result, true)

	name = "devops-system"
	result, exists = checkIndexKey(DefaultIncludeNamespaces, name)
	assert.Equal(t, exists, true)
	assert.Equal(t, result, false)

	name = "kube-system"
	result, exists = checkIndexKey(DefaultIncludeNamespaces, name)
	assert.Equal(t, exists, true)
	assert.Equal(t, result, false)

	name = "sales-fat"
	result, exists = checkIndexKey(nonStrings, name)
	assert.Equal(t, exists, false)
	assert.Equal(t, result, false)

	name = "sales-sit"
	result, exists = checkIndexKey(nonStrings, name)
	assert.Equal(t, exists, false)
	assert.Equal(t, result, false)

	name = "sales-uat"
	result, exists = checkIndexKey(nonStrings, name)
	assert.Equal(t, exists, false)
	assert.Equal(t, result, false)

	name = "devops-system"
	result, exists = checkIndexKey(nonStrings, name)
	assert.Equal(t, exists, false)
	assert.Equal(t, result, false)

	name = "kube-system"
	result, exists = checkIndexKey(nonStrings, name)
	assert.Equal(t, exists, false)
	assert.Equal(t, result, false)

	name = "test"
	exists, verified := checkIndexKey(DefaultExcludeNames, name)
	assert.Equal(t, verified, true)
	assert.Equal(t, exists, false)
}

func TestNamesFilter(t *testing.T) {
	var result bool

	version := map[string]string{
		constants.KubesphereVersion: constants.KubesphereInitVersion,
	}
	app := map[string]string{
		constants.KubesphereAppName: "123",
	}
	appVersion := map[string]string{
		constants.KubesphereAppName: "123",
		constants.KubesphereVersion: constants.KubesphereInitVersion,
	}

	labels1 := map[string]string{
		constants.KubesphereVersion:     constants.KubesphereInitVersion,
		constants.KubesphereInitVersion: constants.KubesphereVersion,
		constants.KubesphereAppName:     "demo",
	}
	labels2 := map[string]string{
		constants.KubesphereVersion:     constants.KubesphereInitVersion,
		constants.KubesphereInitVersion: constants.KubesphereVersion,
		constants.KubesphereAppName:     "123",
	}

	result = checkItemExist(labels1, version, true)
	assert.Equal(t, result, true)

	result = checkItemExist(labels1, app, true)
	assert.Equal(t, result, false)

	result = checkItemExist(labels1, appVersion, true)
	assert.Equal(t, result, true)

	result = checkItemExist(labels1, app, false)
	assert.Equal(t, result, true)

	result = checkItemExist(labels2, appVersion, true)
	assert.Equal(t, result, true)
}
