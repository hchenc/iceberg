package syncer

import (
	"fmt"
	"net/url"
	"testing"
)

func TestClientService(t *testing.T) {
	//gitLabGenerator,_ := NewGitLabGenerator()
	//service := NewGenerateService(gitLabGenerator)
	asd, err := url.Parse("asd")
	defer fmt.Println("finish")
	fmt.Println(asd, err)
	return
}
