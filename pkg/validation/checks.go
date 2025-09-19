package validation

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

func (v *ValidationRun) createVolume() error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-storage-validation",
			Namespace:    v.Configuration.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: ptr.To(v.Configuration.StorageClass),
			Resources: corev1.VolumeResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse(DefaultPVCSize),
				},
			},
		},
	}

	// need to create pvc
}
