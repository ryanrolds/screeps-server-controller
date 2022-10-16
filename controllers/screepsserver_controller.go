/*
Copyright 2022.

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

package controllers

import (
	"bytes"
	"context"
	"fmt"

	template "text/template"

	botClient "github.com/ryanrolds/gh_bot/v1/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	_ "embed"

	screepsv1 "github.com/ryanrolds/screeps-server-controller/api/v1"
	"github.com/ryanrolds/screeps-server-controller/internal/screeps"
)

type dependency struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type privateServerDetails struct {
	// CR details
	Name      string
	Namespace string

	// Private Server
	SteamApiKey    string
	ServerPassword string
	Port           int32
	CliPort        int32
	Image          string

	// Dependencies
	Mongo dependency
	Redis dependency
}

// ScreepsServerReconciler reconciles a ScreepsServer object
type ScreepsServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=screeps.pedanticorderliness.com,resources=screepsservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=screeps.pedanticorderliness.com,resources=screepsservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=screeps.pedanticorderliness.com,resources=screepsservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScreepsServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ScreepsServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting reconciliation")

	// Fetch the ScreepsServer CR
	screepsServer := screepsv1.ScreepsServer{}
	if err := r.Get(ctx, req.NamespacedName, &screepsServer); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("server CR not found, probably deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "error fetching server CR")
		return ctrl.Result{}, err
	}

	// Get private server details, will be used to configure other services
	config, err := r.privateServerDetails(ctx, &screepsServer)
	if err != nil {
		logger.Error(err, "unable to get private server details")
		return ctrl.Result{}, err
	}

	// Ensure Private Server is available
	status, err := r.ensurePrivateServer(ctx, &screepsServer, config)
	if err != nil {
		logger.Error(err, "unable to ensure private server deployment")
		return ctrl.Result{}, err
	}

	screepsServer.Status = *status

	// Update status
	if err := r.Status().Update(ctx, &screepsServer); err != nil {
		logger.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}

	logger.Info("Finished reconciliation")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScreepsServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&screepsv1.ScreepsServer{}).
		Complete(r)
}

func (r *ScreepsServerReconciler) privateServerDetails(ctx context.Context,
	screepsServer *screepsv1.ScreepsServer) (privateServerDetails, error) {
	psd := privateServerDetails{}

	// Fetch ScreepsServer secrets
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "screeps-server-secrets",
		Namespace: screepsServer.Namespace,
	}, secret)
	if err != nil {
		return psd, err
	}

	// Populate private server details
	psd.Name = screepsServer.Name
	psd.Namespace = screepsServer.Namespace
	psd.SteamApiKey = string(secret.Data["stream-api-key"])
	psd.ServerPassword = string(secret.Data["server-password"])
	psd.Port = 21025
	psd.CliPort = 21026

	return psd, nil
}

func (r *ScreepsServerReconciler) ensurePrivateServer(
	ctx context.Context,
	screepsServer *screepsv1.ScreepsServer,
	config privateServerDetails,
) (*screepsv1.ScreepsServerStatus, error) {
	logger := log.FromContext(ctx)

	tmpl, err := template.New("private-server-config").Parse(screeps.PrivateServerConfigTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, err
	}

	configMapName := screepsServer.Name + "-private-server-config"

	// Check if Private Server ConfigMap is present
	privateServerConfig := corev1.ConfigMap{}
	privateServerConfigName := types.NamespacedName{
		Name:      configMapName,
		Namespace: screepsServer.Namespace,
	}
	err = r.Get(ctx, privateServerConfigName, &privateServerConfig)
	if err != nil {
		// normal error handling
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		// If not present, create
		privateServerConfig = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: screepsServer.Namespace,
			},
			Data: map[string]string{
				".screepsrc": buf.String(),
			},
		}
		if err := ctrl.SetControllerReference(screepsServer, &privateServerConfig, r.Scheme); err != nil {
			return nil, err
		}

		if err := r.Create(ctx, &privateServerConfig); err != nil {
			return nil, err
		}
	}

	app := fmt.Sprintf(`%s-private-server`, screepsServer.Name)
	replicas := int32(1)
	image := fmt.Sprintf(`docker.pedanticorderliness.com/screeps-private-server:%s`, screepsServer.Spec.Tag)

	privateServerName := types.NamespacedName{
		Name:      app,
		Namespace: screepsServer.Namespace,
	}
	privateServerDeployment := appsv1.Deployment{}
	err = r.Get(ctx, privateServerName, &privateServerDeployment)
	if err != nil {
		// normal error handling
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf(`problem getting deployment for private server: %w`, err)
		}

		// If not present, create
		privateServerDeployment = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app,
				Namespace: screepsServer.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": app,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": app,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  app,
								Image: image,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      `screepsrc`,
										MountPath: `/pserver/.screepsrc`,
										SubPath:   `.screepsrc`,
									},
								},
							},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: `regcred`,
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: `screepsrc`,
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMapName,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if err := ctrl.SetControllerReference(screepsServer, &privateServerDeployment, r.Scheme); err != nil {
			return nil, fmt.Errorf(`problem setting controller reference on private server deployment: %w`, err)
		}

		if err := r.Create(ctx, &privateServerDeployment); err != nil {
			return nil, fmt.Errorf(`problem creating private server deployment: %w`, err)
		}
	}

	privateServerService := corev1.Service{}
	privateServerServiceName := types.NamespacedName{
		Name:      app,
		Namespace: screepsServer.Namespace,
	}

	err = r.Get(ctx, privateServerServiceName, &privateServerService)
	if err != nil {
		// normal error handling
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		// If not present, create
		privateServerService = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app,
				Namespace: screepsServer.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Selector: map[string]string{
					"app": app,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "game",
						Port:       21025,
						TargetPort: intstr.FromInt(21025),
					},
					{
						Name:       "cli",
						Port:       21026,
						TargetPort: intstr.FromInt(21026),
					},
				},
			},
		}
		if err := ctrl.SetControllerReference(screepsServer, &privateServerService, r.Scheme); err != nil {
			return nil, err
		}

		if err := r.Create(ctx, &privateServerService); err != nil {
			return nil, err
		}
	}

	service := corev1.Service{}
	serviceName := types.NamespacedName{
		Name:      app,
		Namespace: screepsServer.Namespace,
	}
	err = r.Get(ctx, serviceName, &service)
	if err != nil {
		return nil, err
	}

	status := screepsv1.StatusCreated
	host := ""

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			status = screepsv1.StatusRunning
			host = ingress.IP

			if screepsServer.Status.Status != screepsv1.StatusRunning {
				// TODO common on GH PR of running server and IP
				botClient := botClient.New(&botClient.Config{})

				comment := fmt.Sprintf(`Server is now running at %s`, host)
				err := botClient.CommentOnPR(ctx, "screeps-bot-choreographer", comment,
					screepsServer.Spec.PullRequest.Issue)
				if err != nil {
					logger.Error(err, `problem commenting on PR`)
				} else {
					logger.Info("created comment")
				}
			}
		}
	}

	serverStatus := screepsv1.ScreepsServerStatus{
		Status:      status,
		ServiceHost: host,
		ServicePort: 21025,
	}

	return &serverStatus, nil
}
