package config

import (
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"net"
	"os"
	"os/user"
	"path"
)

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "integrate"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/iceberg"
)

type IntegrationConfig struct {
	HarborOptions    *HarborOptions    `json:"harbor_options" yaml:"HarborOptions"`
	GitlabOptions    *GitlabOptions    `json:"gitlab_options" yaml:"GitlabOptions"`
	IntegrateOptions *IntegrateOptions `json:"integrate_options" yaml:"IntegrateOptions"`
}

type HarborOptions struct {
	User     string `json:"user" yaml:"User"`
	Password string `json:"password" yaml:"Password"`
	Host     string `json:"host" yaml:"Host"`
}

type GitlabOptions struct {
	Password string `json:"password" yaml:"Password"`
	Port     string `json:"port" yaml:"Port"`
	Token    string `json:"token" yaml:"Token"`
	User     string `json:"user" yaml:"User"`
	Version  string `json:"version" yaml:"Version"`
	Host     string `json:"host" yaml:"Host"`
}

type IntegrateOptions struct {
	IntegrateOptions []IntegrateOption
}

type IntegrateOption struct {
	CiConfigPath string `json:"ci_config_path" yaml:"CiConfigPath"`
	Pipeline     string `json:"pipeline" yaml:"Pipeline"`
	Template     string `json:"template" yaml:"Template"`
}

type KubernetesOptions struct {
	// kubeconfig
	KubeConfig *rest.Config

	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfigPath string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes apiserver public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

func (h *HarborOptions) Validate() error {
	return nil
}

func (g *GitlabOptions) Validate() error {
	return nil
}

func (i *IntegrateOptions) Validate() error {
	return nil
}

func (k *KubernetesOptions) Validate() error {
	if len(k.KubeConfigPath) != 0 {
		if config, err := clientcmd.BuildConfigFromFlags("", k.KubeConfigPath); err == nil {
			k.KubeConfig = config
			return nil
		} else {
			return err
		}
	}
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	tlsClientConfig := rest.TLSClientConfig{}
	if _, err := certutil.NewPool(rootCAFile); err != nil {
		klog.Errorf("Expected to load root CA config from %s, but got err: %v", rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = rootCAFile
	}

	k.KubeConfig = &rest.Config{
		// TODO: switch to using cluster DNS.
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
		BearerTokenFile: tokenFile,
	}

	return nil
}

// NewKubernetesOptions returns a `zero` instance
func NewKubernetesConfig() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	// make it be easier for those who wants to run api-server locally
	homePath := homedir.HomeDir()
	if homePath == "" {
		// try os/user.HomeDir when $HOME is unset.
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}

	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfigPath = userHomeConfig
	}
	return
}

func (k *KubernetesOptions) AddFlags(fs *pflag.FlagSet, c *KubernetesOptions) {
	fs.StringVar(&k.KubeConfigPath, "kubeconfig", c.KubeConfigPath, ""+
		"Path for kubernetes kubeconfig file, if left blank, will use "+
		"in cluster way.")
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*IntegrationConfig, error) {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := &IntegrationConfig{}

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil

}
