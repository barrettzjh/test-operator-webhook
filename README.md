# test-operator-webhook
build test-operator with webhook

**基于test-operator**

```shell
git clone https://github.com/barrettzjh/test-operator.git

目录结构
# tree
.
├── api
│   └── v1
│       ├── groupversion_info.go
│       ├── test_types.go
│       └── zz_generated.deepcopy.go
├── bin
│   └── manager
├── config
│   ├── certmanager
│   │   ├── certificate.yaml
│   │   ├── kustomization.yaml
│   │   └── kustomizeconfig.yaml
│   ├── crd
│   │   ├── bases
│   │   │   └── test.test.zhangjinhui.online_tests.yaml
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_tests.yaml
│   │       └── webhook_in_tests.yaml
│   ├── default
│   │   ├── kustomization.yaml
│   │   ├── manager_auth_proxy_patch.yaml
│   │   ├── manager_webhook_patch.yaml
│   │   └── webhookcainjection_patch.yaml
│   ├── manager
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ├── prometheus
│   │   ├── kustomization.yaml
│   │   └── monitor.yaml
│   ├── rbac
│   │   ├── auth_proxy_client_clusterrole.yaml
│   │   ├── auth_proxy_role_binding.yaml
│   │   ├── auth_proxy_role.yaml
│   │   ├── auth_proxy_service.yaml
│   │   ├── kustomization.yaml
│   │   ├── leader_election_role_binding.yaml
│   │   ├── leader_election_role.yaml
│   │   ├── role_binding.yaml
│   │   ├── role.yaml
│   │   ├── test_editor_role.yaml
│   │   └── test_viewer_role.yaml
│   ├── samples
│   │   └── test_v1_test.yaml
│   └── webhook
│       ├── kustomization.yaml
│       ├── kustomizeconfig.yaml
│       └── service.yaml
├── controllers
│   ├── suite_test.go
│   └── test_controller.go
├── cover.out
├── Dockerfile
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
├── main.go
├── Makefile
├── PROJECT
└── README.md

16 directories, 46 files

```

**部署cert-manager**

```shell
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml
```

**添加webhook**

```shell
# kubebuilder create webhook --group test --version v1 --kind Test --defaulting --programmatic-validation
Writing scaffold for you to edit...
api/v1/test_webhook.go


此时新创建了一个go文件。
```

**修改webhook逻辑**

```go
package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

//定义日志
var testlog = logf.Log.WithName("test-resource")

//定义启动方法
func (r *Test) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//注意 默认生成的marker是v1版本的，需要做修改，可对比该示例
// +kubebuilder:webhook:path=/mutate-test-test-zhangjinhui-online-v1-test,mutating=true,failurePolicy=fail,groups=test.test.zhangjinhui.online,resources=tests,verbs=create;update,versions=v1,name=mtest.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

//申明defaulter
var _ webhook.Defaulter = &Test{}

//默认数据操作
func (r *Test) Default() {
	testlog.Info("default", "name", r.Name)
	//定义如果Port和TargetPort为空时，给予定义默认值
	if r.Spec.Port == 0{
		r.Spec.Port = 80
	}

	if r.Spec.TargetPort == 0{
		r.Spec.TargetPort = 80
        testlog.Info("set default value 80 for TargetPort")
	}
}

//注意 默认生成的marker是v1版本的，需要做修改，可对比该示例
// +kubebuilder:webhook:verbs=create;update,path=/validate-test-test-zhangjinhui-online-v1-test,mutating=false,failurePolicy=fail,groups=test.test.zhangjinhui.online,resources=tests,versions=v1,name=vtest.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

//申明validator
var _ webhook.Validator = &Test{}

//创建validate函数
func (r *Test) ValidateCreate() error {
	testlog.Info("validate create", "name", r.Name)
	return r.validateTest()
}

//更新validate函数
func (r *Test) ValidateUpdate(old runtime.Object) error {
	testlog.Info("validate update", "name", r.Name)
	return r.validateTest()
}

//删除validate函数
func (r *Test) ValidateDelete() error {
	testlog.Info("validate delete", "name", r.Name)
	return nil
}

//判断规则是否通过，如果有不通过则返回apierrors
//在create 和 update 中执行该函数
func (r *Test) validateTest() error {
	var allErrs field.ErrorList
	if err := r.validateNodePort(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "test.test.zhangjinhui.online", Kind: "Test"},
		r.Name, allErrs)
}

//定义准入规则， 不允许nodePort为40000
func (r *Test) validateNodePort() *field.Error {
	if r.Spec.NodePort == 40000 {
		return field.Invalid(field.NewPath("spec").Child("schedule"), r.Spec.NodePort, "nodePort 不能为40000")
	}
	return nil
}
```

**修改config**

**config/default/kustomization.yaml**

```yaml
# Adds namespace to all resources.
namespace: test-kubebuilder-system

# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: test-kubebuilder-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

bases:
- ../crd
- ../rbac
- ../manager
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in 
# crd/kustomization.yaml
- ../webhook
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'. 'WEBHOOK' components are required.
- ../certmanager
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'. 
#- ../prometheus

patchesStrategicMerge:
  # Protect the /metrics endpoint by putting it behind auth.
  # If you want your controller-manager to expose the /metrics
  # endpoint w/o any authn/z, please comment the following line.
- manager_auth_proxy_patch.yaml

# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in 
# crd/kustomization.yaml
- manager_webhook_patch.yaml

# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'.
# Uncomment 'CERTMANAGER' sections in crd/kustomization.yaml to enable the CA injection in the admission webhooks.
# 'CERTMANAGER' needs to be enabled to use ca injection
- webhookcainjection_patch.yaml

# the following config is for teaching kustomize how to do var substitution
vars:
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER' prefix.
- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
  fieldref:
    fieldpath: metadata.namespace
- name: CERTIFICATE_NAME
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
- name: SERVICE_NAMESPACE # namespace of the service
  objref:
    kind: Service
    version: v1
    name: webhook-service
  fieldref:
    fieldpath: metadata.namespace
- name: SERVICE_NAME
  objref:
    kind: Service
    version: v1
    name: webhook-service

```

**config/certmanager/certificate.yaml**

```yaml
# The following manifests contain a self-signed issuer CR and a certificate CR.
# More document can be found at https://docs.cert-manager.io
# WARNING: Targets CertManager 0.11 check https://docs.cert-manager.io/en/latest/tasks/upgrading/index.html for 
# breaking changes
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: system
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: serving-cert  # this name should match the one appeared in kustomizeconfig.yaml
  namespace: system
spec:
  # $(SERVICE_NAME) and $(SERVICE_NAMESPACE) will be substituted by kustomize
  dnsNames:
  - $(SERVICE_NAME).$(SERVICE_NAMESPACE).svc
  - $(SERVICE_NAME).$(SERVICE_NAMESPACE).svc.cluster.local
  issuerRef:
    kind: Issuer
    name: selfsigned-issuer
  secretName: webhook-server-cert # this secret will not be prefixed, since it's not managed by kustomize
```

**config/default/webhookcainjection_patch.yaml**

```yaml
# This patch add annotation to admission webhook config and
# the variables $(CERTIFICATE_NAMESPACE) and $(CERTIFICATE_NAME) will be substituted by kustomize.
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
  annotations:
    cert-manager.io/inject-ca-from: $(CERTIFICATE_NAMESPACE)/$(CERTIFICATE_NAME)
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating-webhook-configuration
  annotations:
    cert-manager.io/inject-ca-from: $(CERTIFICATE_NAMESPACE)/$(CERTIFICATE_NAME)
```

**编译**

```shell
export IMG=harbor.yunyang.com.cn/library/test-kubebuilder:v0.27
make docker-build docker-push
...
```

**部署**

```shell
make deploy
...

# kubectl get pod -n test-kubebuilder-system
NAME                                                   READY   STATUS    RESTARTS   AGE
test-kubebuilder-controller-manager-7dbbb5b499-sxzph   2/2     Running   0          59s
```

**测试CRD**

```shell
# kubectl apply -f config/samples/test_v1_test.yaml 
test.test.test.zhangjinhui.online/test-sample created

# kubectl get pod,svc
NAME                                       READY   STATUS    RESTARTS   AGE
pod/test-sample-669d698cbd-5z5gx           1/1     Running   0          4s
pod/test-sample-669d698cbd-dcblc           1/1     Running   0          4s
pod/test-sample-669d698cbd-mgh7p           1/1     Running   0          4s

NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
service/test-sample           NodePort    10.0.136.215   <none>        80:30043/TCP     4s

# kubectl delete -f config/samples/test_v1_test.yaml 
test.test.test.zhangjinhui.online "test-sample" deleted

# kubectl get pod,svc
NAME                                       READY   STATUS    RESTARTS   AGE
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
```

**测试default**

```shell
修改config/samples/test_v1_test.yaml
apiVersion: test.test.zhangjinhui.online/v1
kind: Test
metadata:
  name: test-sample
spec:
  # Add fields here
  image: nginx:latest
  replicas: 3
  port: 80
  #targetPort: 80
  nodePort: 30043
  
# kubectl apply -f config/samples/test_v1_test.yaml
test.test.test.zhangjinhui.online/test-sample created

# kubectl logs -f test-kubebuilder-controller-manager-7dd58f5dbb-zxrmc -n test-kubebuilder-system -c manager
2021-03-26T03:06:45.890Z	INFO	test-resource	default	{"name": "test-sample"}
2021-03-26T03:06:45.890Z	INFO	test-resource	set default value 80 for TargetPort

这里不测试Port，前面在定义test_types.go时，没有设置omitempty
如果不配置Port的话
# kubectl apply -f config/samples/test_v1_test.yaml
error: error validating "config/samples/test_v1_test.yaml": error validating data: ValidationError(Test.spec): missing required field "port" in online.zhangjinhui.test.test.v1.Test.spec; if you choose to ignore these errors, turn validation off with --validate=false
```



**测试validate**

```shell
修改config/samples/test_v1_test.yaml
apiVersion: test.test.zhangjinhui.online/v1
kind: Test
metadata:
  name: test-sample
spec:
  # Add fields here
  image: nginx:latest
  replicas: 3
  port: 80
  targetPort: 80
  nodePort: 40000

# kubectl apply -f config/samples/test_v1_test.yaml 
Error from server (Test.test.test.zhangjinhui.online "test-sample" is invalid: spec.schedule: Invalid value: 40000: nodePort 不能为40000): error when creating "config/samples/test_v1_test.yaml": admission webhook "vtest.kb.io" denied the request: Test.test.test.zhangjinhui.online "test-sample" is invalid: spec.schedule: Invalid value: 40000: nodePort 不能为40000

# kubectl logs -f test-kubebuilder-controller-manager-7dbbb5b499-sxzph -n test-kubebuilder-system -c manager
2021-03-26T02:58:29.709Z	DEBUG	controller-runtime.webhook.webhooks	wrote response	{"webhook": "/validate-test-test-zhangjinhui-online-v1-test", "UID": "fd5e1107-83fc-4b5f-be30-bfd7536bc1f1", "allowed": false, "result": {}, "resultError": "got runtime.Object without object metadata: &Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:,Message:,Reason:Test.test.test.zhangjinhui.online \"test-sample\" is invalid: spec.schedule: Invalid value: 40000: nodePort 不能为40000,Details:nil,Code:403,}"}
```

此时一个简单的webhook就写好了