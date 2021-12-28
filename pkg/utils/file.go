package utils

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
)

func WriteConfigTo(obj interface{}, fpath string) error {
	data, _ := yaml.Marshal(obj)
	err := ioutil.WriteFile(fpath, data, 0666)
	return err
}
