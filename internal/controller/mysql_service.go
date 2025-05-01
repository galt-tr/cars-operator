package controller

import (
	infrav1alpha1 "github.com/bitcoin-sv/cars-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileMysqlService is the mysql service reconciler
func (r *CarsReconciler) ReconcileMysqlService(log logr.Logger) (bool, error) {
	cars := infrav1alpha1.Cars{}
	if err := r.Get(r.Context, r.NamespacedName, &cars); err != nil {
		return false, err
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: r.NamespacedName.Namespace,
			Labels:    getAppLabels(),
		},
	}
	_, err := controllerutil.CreateOrUpdate(r.Context, r.Client, &svc, func() error {
		return r.updateMysqlService(&svc, &cars)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CarsReconciler) updateMysqlService(svc *corev1.Service, cars *infrav1alpha1.Cars) error {
	err := controllerutil.SetControllerReference(cars, svc, r.Scheme)
	if err != nil {
		return err
	}
	svc.Spec = *defaultMysqlServiceSpec()
	return nil
}

func defaultMysqlServiceSpec() *corev1.ServiceSpec {
	labels := map[string]string{
		"app": "mysql",
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
				Name:       "mysql-tcp",
				Port:       int32(MysqlPort),
				TargetPort: intstr.FromInt32(MysqlPort),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}
}
