package controller

import (
	"fmt"
	infrav1alpha1 "github.com/bitcoin-sv/cars-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileService is the cars service reconciler
func (r *CarsReconciler) ReconcileService(log logr.Logger) (bool, error) {
	cars := infrav1alpha1.Cars{}
	if err := r.Get(r.Context, r.NamespacedName, &cars); err != nil {
		return false, err
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cars",
			Namespace: r.NamespacedName.Namespace,
			Labels:    getAppLabels(),
		},
	}
	_, err := controllerutil.CreateOrUpdate(r.Context, r.Client, &svc, func() error {
		return r.updateService(&svc, &cars)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CarsReconciler) updateService(svc *corev1.Service, cars *infrav1alpha1.Cars) error {
	err := controllerutil.SetControllerReference(cars, svc, r.Scheme)
	if err != nil {
		return err
	}
	svc.Spec = *defaultCarsServiceSpec()
	return nil
}

func defaultCarsServiceSpec() *corev1.ServiceSpec {
	labels := map[string]string{
		"app": "cars",
	}
	ipFamily := corev1.IPFamilyPolicySingleStack
	return &corev1.ServiceSpec{
		Selector:       labels,
		ClusterIP:      "None",
		IPFamilyPolicy: &ipFamily,
		IPFamilies: []corev1.IPFamily{
			corev1.IPv4Protocol,
		},
		Ports: []corev1.ServicePort{
			{
				Name:       "cars-tcp",
				Port:       int32(CarsPort),
				TargetPort: intstr.FromInt32(CarsPort),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}
}

// ReconcileIngress is the ingress
func (r *CarsReconciler) ReconcileIngress(log logr.Logger) (bool, error) {
	cars := infrav1alpha1.Cars{}
	if err := r.Get(r.Context, r.NamespacedName, &cars); err != nil {
		return false, err
	}
	// Skip if domain isn't set
	if cars.Spec.Domain == "" {
		return false, nil
	}
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cars",
			Namespace: r.NamespacedName.Namespace,
			Labels:    getAppLabels(),
		},
	}
	_, err := controllerutil.CreateOrUpdate(r.Context, r.Client, &ingress, func() error {
		return r.updateIngress(&ingress, &cars)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CarsReconciler) updateIngress(ingress *networkingv1.Ingress, cars *infrav1alpha1.Cars) error {
	err := controllerutil.SetControllerReference(cars, ingress, r.Scheme)
	if err != nil {
		return err
	}
	if cars.Spec.ClusterIssuer != "" {
		if ingress.Annotations == nil {
			ingress.Annotations = make(map[string]string)
		}
		ingress.Annotations["cert-manager.io/cluster-issuer"] = cars.Spec.ClusterIssuer
	}
	ingress.Spec = *defaultCarsIngressSpec(cars)
	return nil
}

func defaultCarsIngressSpec(cars *infrav1alpha1.Cars) *networkingv1.IngressSpec {
	pathType := networkingv1.PathTypeImplementationSpecific
	class := "nginx"
	return &networkingv1.IngressSpec{
		IngressClassName: &class,
		TLS: []networkingv1.IngressTLS{
			{
				Hosts: []string{
					fmt.Sprintf("%s.%s", cars.Name, cars.Spec.Domain),
				},
				SecretName: "cars-tls",
			},
		},
		Rules: []networkingv1.IngressRule{
			{
				Host: fmt.Sprintf("%s.%s", cars.Name, cars.Spec.Domain),
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "cars",
										Port: networkingv1.ServiceBackendPort{
											Number: int32(CarsPort),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
