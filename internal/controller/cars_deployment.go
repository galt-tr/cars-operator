package controller

import (
	infrav1alpha1 "github.com/bitcoin-sv/cars-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileDeployment is the cars deployment reconciler
func (r *CarsReconciler) ReconcileDeployment(log logr.Logger) (bool, error) {
	cars := infrav1alpha1.Cars{}
	if err := r.Get(r.Context, r.NamespacedName, &cars); err != nil {
		return false, err
	}
	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cars",
			Namespace: r.NamespacedName.Namespace,
			Labels:    getAppLabels(),
		},
	}
	_, err := controllerutil.CreateOrUpdate(r.Context, r.Client, &dep, func() error {
		return r.updateDeployment(&dep, &cars)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CarsReconciler) updateDeployment(dep *appsv1.Deployment, cars *infrav1alpha1.Cars) error {
	err := controllerutil.SetControllerReference(cars, dep, r.Scheme)
	if err != nil {
		return err
	}
	dep.Spec = *defaultCarsDeploymentSpec()

	if cars.Spec.Image != "" {
		dep.Spec.Template.Spec.Containers[0].Image = cars.Spec.Image
	}

	return nil
}

func defaultCarsDeploymentSpec() *appsv1.DeploymentSpec {
	labels := map[string]string{
		"app":        "cars",
		"deployment": "cars",
	}
	envFrom := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "cars-environment",
				},
			},
		},
	}
	env := []corev1.EnvVar{}
	return &appsv1.DeploymentSpec{
		Replicas: ptr.To(int32(1)),
		Selector: metav1.SetAsLabelSelector(labels),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Time{},
				Labels:            labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						EnvFrom:         envFrom,
						Env:             env,
						Image:           DefaultImage,
						ImagePullPolicy: corev1.PullAlways,
						Name:            "cars",
						// Make sane defaults, and this should be configurable
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("500Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: CarsPort,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					/*{
						Name: "cars-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "cars-config",
								},
							},
						},
					},*/
				},
			},
		},
	}
}
