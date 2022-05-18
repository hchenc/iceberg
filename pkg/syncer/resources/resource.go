package resources

type assembleResourceFunc func(obj interface{}, namespace string) interface{}

func assembleResource(obj interface{}, namespace string, resourceFunc assembleResourceFunc) interface{} {
	return resourceFunc(obj, namespace)
}
