package validation

import (
	"context"
	"fmt"
	"os"

	"github.com/harvester/storage-validator/pkg/api"
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
}

func (v *ValidationRun) Execute() error {
	// readConfiguration File
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
	return nil
}
