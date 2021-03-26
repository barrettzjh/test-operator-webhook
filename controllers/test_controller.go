/*


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
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	testv1 "zhangjinhui.online/m/api/v1"
)

// TestReconciler reconciles a Test object
type TestReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (c *TestReconciler) isCrdExist(ctx context.Context, req ctrl.Request)(testv1.Test ,error) {
	var crd testv1.Test
	err := c.Get(ctx, req.NamespacedName, &crd)
	return crd, err
}

func (c *TestReconciler) isDeployExist(ctx context.Context, req ctrl.Request)(v1.Deployment,error) {
	var deploy v1.Deployment
	err := c.Client.Get(ctx, req.NamespacedName, &deploy)
	return deploy, err
}

func (c *TestReconciler) isServiceExist(ctx context.Context, req ctrl.Request)(v12.Service, error) {
	var service v12.Service
	err:= c.Client.Get(ctx, req.NamespacedName, &service)
	return service, err
}

func (c *TestReconciler) createDeployment(ctx context.Context, test testv1.Test) {
	if err := c.Client.Create(ctx, getDeployment(test.Name, test.Namespace, test.Spec.Replicas, test.Spec.Image));err != nil{
		reqLogger.Info("创建deployment失败!"+err.Error())
	}
}

func (c *TestReconciler) createService(ctx context.Context, test testv1.Test) {
	if err := c.Client.Create(ctx, getService(test.Name, test.Namespace,test.Spec.Port, test.Spec.NodePort));err != nil{
		reqLogger.Info("创建service失败!"+err.Error())
	}
}

func (c *TestReconciler) updateDeployment(ctx context.Context, test testv1.Test, deployment v1.Deployment) {
	deployment.Spec.Replicas = &test.Spec.Replicas
	deployment.Spec.Template.Spec.Containers[0].Image = test.Spec.Image
	if err := c.Client.Update(ctx, &deployment);err != nil{
		reqLogger.Info("更新deployment失败!"+err.Error())
	}
}

func (c *TestReconciler) updateService(ctx context.Context, test testv1.Test, service v12.Service) {
	service.Spec.Ports[0].Port = test.Spec.Port
	service.Spec.Ports[0].NodePort = test.Spec.NodePort
	if err := c.Client.Update(ctx, &service);err != nil{
		reqLogger.Info("更新service失败!"+err.Error())
	}
}

func (c *TestReconciler) clear(ctx context.Context, req ctrl.Request) {
	//清理deploy
	if deploy, err := c.isDeployExist(ctx, req);err == nil{
		if DeployErr := c.Client.Delete(ctx, &deploy);DeployErr!=nil{
			reqLogger.Info("清理deploy失败， name=" + deploy.Name)
		}
	}
	//清理service
	if service, err := c.isServiceExist(ctx, req); err == nil{
		if ServiceErr := c.Client.Delete(ctx, &service);ServiceErr != nil{
			reqLogger.Info("清理service失败， name=" + service.Name)
		}
	}
}


// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=test.test.zhangjinhui.online,resources=tests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test.test.zhangjinhui.online,resources=tests/status,verbs=get;update;patch


var reqLogger  logr.Logger

func (c *TestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger = c.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	//判断该次Reconcile应该做什么操作
	var test testv1.Test
	//判断该cr还存不存在， 如果不存在，则清理资源
	test, err := c.isCrdExist(ctx, req)
	if err != nil{
		//报错则已经被删除， 此时需要清理该crd的资源，然后退出
		c.clear(ctx, req)
		return ctrl.Result{}, nil
	}


	//如果获取到该cr， 则更新，或创建资源
	//判断是否存在deployment，存在则更新deploy，不存在则创建
	if deploy, err := c.isDeployExist(ctx, req);err != nil{
		reqLogger.Info("没有获取到该deployment,正在创建该deployment.")
		c.createDeployment(ctx, test)
	}else {
		reqLogger.Info("获取到deployment,正在更新该deployment.")
		c.updateDeployment(ctx, test, deploy)
	}

	if service, err := c.isServiceExist(ctx, req);err != nil{
		reqLogger.Info("没有获取到该service,正在创建该service.")
		c.createService(ctx, test)
	}else {
		reqLogger.Info("获取到service,正在更新该service.")
		c.updateService(ctx, test, service)
	}
	// your logic here

	return ctrl.Result{}, nil
}

func (c *TestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Test{}).
		Complete(c)
}

func getDeployment(name string, namespace string, replicas int32, image string) *v1.Deployment {
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name:  name,
							Image: image,
							Ports: []v12.ContainerPort{
								{
									Name:          "http",
									Protocol:      v12.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}
func int32Ptr(i int32) *int32 { return &i }

func getService(name string, namespace string, port int32, nodePort int32) *v12.Service {
	service := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v12.ServiceSpec{
			Type: "NodePort",
			Ports: []v12.ServicePort{{
				Name:     "http",
				Port:     port,
				Protocol: "TCP",
				NodePort: nodePort,
			}},
			Selector: map[string]string{"app": name},
		},
	}
	return service
}


