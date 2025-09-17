package validation

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/harvester/storage-validator/pkg/api"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	harvesterhciv1beta1 "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/rancher/wrangler/v3/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clients        HarvesterClient
}

type HarvesterClient struct {
	coreClient *kubernetes.Clientset
	hvClient   *harvesterhciv1beta1.Clientset
}

func (v *ValidationRun) Execute() error {
	// initialise context
	v.ctx = signals.SetupSignalContext()

	// read configuration file
	if err := v.readConfig(); err != nil {
		return err
	}

	// generate k8s clients
	if err := v.setupClients(); err != nil {
		return err
	}

	// run preflight checks
	logrus.Info("running preflight checks")
	if err := v.preFlightChecks(); err != nil {
		return err
	}

	// apply system wide defaults
	if err := v.applyValidatinoDefaults(); err != nil {
		return err
	}

	return nil
}

// readConfig will read the configuration file and prep the
func (v *ValidationRun) readConfig() error {
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

// running preflight checks
func (v *ValidationRun) preFlightChecks() error {
	if v.Configuration.ImageURL == "" {
		return errors.New("no imageURL specified, aborting run")
	}
	return nil
}

// ApplyDefaults will apply sane defaults for the storage validation configuration
func (v *ValidationRun) applyValidatinoDefaults() error {
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

	if v.Configuration.SkipCleanup == nil {
		v.Configuration.SkipCleanup = &[]bool{true}[0]
	}

	// verify and apply default storageClass if one is not present
	if v.Configuration.StorageClass == "" {
		logrus.Warnf("no default storage class specified, looking up default storageclass")
		scList, err := v.clients.hvClient.StorageV1().StorageClasses().List(v.ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing storageclasses: %w", err)
		}
		for _, sc := range scList.Items {
			if val, ok := sc.Annotations[defaultSCAnnotation]; ok && val == "true" {
				v.Configuration.StorageClass = sc.Name
			}
		}
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

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error generating kubernetes client interface: %w", err)
	}

	harvesterClient, err := harvesterhciv1beta1.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error generating harvester client interfce: %w", err)
	}

	clients := HarvesterClient{
		coreClient: coreClient,
		hvClient:   harvesterClient,
	}
	v.cfg = cfg
	v.clients = clients
	return nil
}
