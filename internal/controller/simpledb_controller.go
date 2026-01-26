/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "my.domain/db-operator/api/v1"
)

// SimpleDBReconciler reconciles a SimpleDB object
type SimpleDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC permissions
// +kubebuilder:rbac:groups=database.my.domain,resources=simpledbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.my.domain,resources=simpledbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.my.domain,resources=simpledbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *SimpleDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// 1. Fetch SimpleDB
	simpledb := &databasev1.SimpleDB{}
	if err := r.Get(ctx, req.NamespacedName, simpledb); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Fetch Deployment
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      simpledb.Name,
		Namespace: simpledb.Namespace,
	}, found)

	// 3. Create Deployment if missing
	if err != nil && errors.IsNotFound(err) {
		dep, err := r.deploymentForSimpleDB(simpledb)
		if err != nil {
			return ctrl.Result{}, err
		}

		l.Info("Creating Deployment", "name", dep.Name)
		if err := r.Create(ctx, dep); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Reconcile replicas
	size := simpledb.Spec.Replicas
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		if err := r.Update(ctx, found); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 5. Compute readiness
	ready := found.Status.UpdatedReplicas == size &&
		found.Status.ReadyReplicas == size &&
		found.Status.AvailableReplicas == size

	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "Creating",
		Message:            "Database is being created",
		LastTransitionTime: metav1.Now(),
	}

	if ready {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "Available"
		condition.Message = "Database is ready"
	}

	// 6. Avoid useless status updates
	existing := meta.FindStatusCondition(simpledb.Status.Conditions, "Ready")
	if existing != nil &&
		existing.Status == condition.Status &&
		existing.Reason == condition.Reason {
		return ctrl.Result{}, nil
	}

	meta.SetStatusCondition(&simpledb.Status.Conditions, condition)

	if err := r.Status().Update(ctx, simpledb); err != nil {
		l.Error(err, "Failed to update SimpleDB status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager wires controller to manager
func (r *SimpleDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.SimpleDB{}).
		Owns(&appsv1.Deployment{}).
		Named("simpledb").
		Complete(r)
}

// deploymentForSimpleDB builds Deployment
func (r *SimpleDBReconciler) deploymentForSimpleDB(m *databasev1.SimpleDB) (*appsv1.Deployment, error) {
	labels := labelsForSimpleDB(m.Name)
	replicas := m.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "simpledb",
							Image: m.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 5432,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_PASSWORD",
									Value: "changeme", // demo-only; real operators must use Secrets
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(m, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil

}

// labelsForSimpleDB returns common labels
func labelsForSimpleDB(name string) map[string]string {
	return map[string]string{
		"app":         "simpledb",
		"simpledb_cr": name,
	}
}
