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

// log is for logging in this package.
var testlog = logf.Log.WithName("test-resource")

func (r *Test) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//注意 默认生成的marker是v1版本的，需要做修改，可对比该示例
// +kubebuilder:webhook:path=/mutate-test-test-zhangjinhui-online-v1-test,mutating=true,failurePolicy=fail,groups=test.test.zhangjinhui.online,resources=tests,verbs=create;update,versions=v1,name=mtest.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &Test{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
//默认数据操作
func (r *Test) Default() {
	testlog.Info("default", "name", r.Name)
	//定义如果Port和TargetPort为空时，给与定义默认值
	if r.Spec.Port == 0{
		r.Spec.Port = 80
	}

	if r.Spec.TargetPort == 0{
		r.Spec.TargetPort = 80
		testlog.Info("set default value 80 for TargetPort")
	}
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//注意 默认生成的marker是v1版本的，需要做修改，可对比该示例
// +kubebuilder:webhook:verbs=create;update,path=/validate-test-test-zhangjinhui-online-v1-test,mutating=false,failurePolicy=fail,groups=test.test.zhangjinhui.online,resources=tests,versions=v1,name=vtest.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Validator = &Test{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Test) ValidateCreate() error {
	testlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateTest()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Test) ValidateUpdate(old runtime.Object) error {
	testlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validateTest()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Test) ValidateDelete() error {
	testlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
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

