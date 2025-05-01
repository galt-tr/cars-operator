package controller

import (
	infrav1alpha1 "github.com/bitcoin-sv/cars-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileMysqlPVC is the mysql PVC
func (r *CarsReconciler) ReconcileMysqlPVC(log logr.Logger) (bool, error) {
	cars := infrav1alpha1.Cars{}
	if err := r.Get(r.Context, r.NamespacedName, &cars); err != nil {
		return false, err
	}
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-data",
			Namespace: r.NamespacedName.Namespace,
			Labels:    getAppLabels(),
		},
	}
	// Check if PVC is already created so that we can copy the existing spec values
	// This is how we properly support resizing
	existingPvcNamespacedName := types.NamespacedName{
		Namespace: pvc.Namespace,
		Name:      pvc.Name,
	}
	existingPVC := &corev1.PersistentVolumeClaim{}
	err := r.Get(r.Context, existingPvcNamespacedName, existingPVC)
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, err
	}

	// If in cluster PVC is not found; nil it out to not confuse the create or update section
	if k8serrors.IsNotFound(err) {
		existingPVC = nil
	}

	_, err = controllerutil.CreateOrUpdate(r.Context, r.Client, &pvc, func() error {
		return r.updatePVC(&pvc, existingPVC, &cars)
	})

	// Ignore forbidden errors
	if err != nil && !k8serrors.IsForbidden(err) {
		return false, err
	}
	return true, nil
}

func (r *CarsReconciler) updatePVC(pvc *corev1.PersistentVolumeClaim, inClusterPVC *corev1.PersistentVolumeClaim, cars *infrav1alpha1.Cars) error {
	err := controllerutil.SetControllerReference(cars, pvc, r.Scheme)
	if err != nil {
		return err
	}
	if inClusterPVC == nil {
		pvc.Spec = *defaultPVCSpec()
	} else {
		pvc.Spec = *inClusterPVC.Spec.DeepCopy()
	}

	// If storage class is configured, use it
	if cars.Spec.StorageClass != "" {
		pvc.Spec.StorageClassName = &cars.Spec.StorageClass
	}
	// If storage resources are configured, use them
	if cars.Spec.StorageResources != nil {
		pvc.Spec.Resources = *cars.Spec.StorageResources
	}
	if cars.Spec.StorageVolume != "" {
		pvc.Spec.VolumeName = cars.Spec.StorageVolume
	}
	return nil
}

func defaultPVCSpec() *corev1.PersistentVolumeClaimSpec {
	emptyStorageClass := ""
	return &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		StorageClassName: &emptyStorageClass,
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				"storage": resource.MustParse("5Gi"),
			},
		},
	}
}
