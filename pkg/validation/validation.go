package validation

import (
	"context"
	"fmt"
	"os"

	harvester "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/storage-validator/pkg/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type ValidationRun struct {
	ConfigFile     string
	ctx            context.Context
	Configuration  *api.Configuration
	Result         *api.Result
	createdObjects []runtime.Object
	cfg            *rest.Config
	client         kubernetes.Interface
	hvClient       harvester.Interface
}

func (v *ValidationRun) Execute() error {
	// generate k8s clients
	if err := v.setupClients(); err != nil {
		return err
	}

	return nil
}

// ReadConfig will read the configuration file and prep the
func (v *ValidationRun) ReadConfig() error {
	contents, err := os.ReadFile(v.ConfigFile)
	if err != nil {
		return fmt.Errorf("error reading configFile %s: %w", v.ConfigFile, err)
	}

	configObj := &api.Configuration{}
	err = yaml.Unmarshal(contents, configObj)
	if err != nil {
		return fmt.Errorf("error unmarshalling configfile: %v", err)
	}
	v.Configuration = configObj
	return nil
}

// ApplyDefaults will apply sane defaults for the storage validation configuration
func (v *ValidationRun) ApplyValidatinoDefaults() error {
	if v.Configuration.VMConfig.CPU == 0 {
		v.Configuration.VMConfig.CPU = DefaultCPU
	}

	if v.Configuration.VMConfig.Memory == "" {
		v.Configuration.VMConfig.Memory = DefaultMem
	}

	if v.Configuration.VMConfig.Memory == "" {
		v.Configuration.VMConfig.Memory = DefaultMem
	}

	if v.Configuration.Timeout == nil {
		v.Configuration.Timeout = &[]int{DefaultTimeout}[0]
	}

	return nil
}

func (v *ValidationRun) setupClients() error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	cfg, err := kubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig %v", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error generating kubernetes client interface: %w", err)
	}
	v.client = client

	hvClient, err := harvester.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error generating harvester client interface: %w", err)
	}
	v.hvClient = hvClient
	return nil
}
