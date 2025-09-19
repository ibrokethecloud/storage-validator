package validation

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	err := v.clients.runtimeClient.Create(v.ctx, pvc)
	if err != nil {
		return fmt.Errorf("error creating pvc: %w", err)
	}

	// store for cleanup later on
	v.createdObjects = append(v.createdObjects, pvc)

	// attach pvc to pod to ensure creation
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-storage-validation",
			Namespace:    v.Configuration.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "pvc-storage-validation",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				},
			},
		},
	}

	err = v.clients.runtimeClient.Create(v.ctx, pod)
	if err != nil {
		return fmt.Errorf("error creating pvc: %w", err)
	}
	v.createdObjects = append(v.createdObjects, pod)

	// reconcile until pod is running and ensure pvc is bound
	return nil
}

// fetch and verify that specified object is ready
// else keep retrying till verification times out
func (v *ValidationRun) checkObjectIsReady(obj client.Object, check func(obj client.Object) bool) (bool, error) {
	for {
		err := v.clients.runtimeClient.Get(v.ctx, client.ObjectKeyFromObject(obj), obj)
		if err != nil {
			return false, fmt.Errorf("error getting object %v: %w", client.ObjectKeyFromObject(obj), err)
		}

		ready := check(obj)
		if !ready {
			time.Sleep(5 * time.Second)
		} else {
			return true, nil
		}

	}
}
