package validation

import (
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Current validation requirements are
// * create a volume
// * create a snapshot
// * perform offline volume expansion
// * create a vmimage using the storage class specified
// * boot a vm using storage class
// * hotplug 2 volumes to a vm
// * create vm snapshots
// * perform live migration across nodes

func (v *ValidationRun) runChecks() error {
	return nil
}

func (v *ValidationRun) cleanupResources() error {
	if !*v.Configuration.SkipCleanup {
		logrus.Info("skipping object cleanup")
		return nil
	}
	for _, obj := range v.createdObjects {
		if err := v.clients.runtimeClient.Delete(v.ctx, obj); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			// just log error and move on and attempt to clean up remaining objects
			logrus.Errorf("error deleting object %s", obj.GetName())
		}
	}
	return nil
}
