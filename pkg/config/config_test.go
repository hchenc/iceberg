package config

import (
	"fmt"
	"testing"
)

func TestNewKubernetesConfig(t *testing.T) {
	config := NewKubernetesConfig()
	fmt.Println(config)
}
