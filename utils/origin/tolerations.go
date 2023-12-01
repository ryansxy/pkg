package origin

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	policyv1alpha1 "github.com/k-cloud-labs/pkg/apis/policy/v1alpha1"
)

type Tolerations struct {
	Value []corev1.Toleration
}

func (t *Tolerations) GetJsonPatch(rawObj *unstructured.Unstructured, Replace bool, operator policyv1alpha1.OverriderOperator) (*OverrideOption, error) {
	if operator == policyv1alpha1.OverriderOpAdd || operator == policyv1alpha1.OverriderOpRemove {
		return nil, errors.New("unsupported operator type error")
	}

	k := rawObj.GetKind()

	op := &OverrideOption{
		Op: string(policyv1alpha1.OverriderOpReplace),
	}

	if k == PodKind {
		op.Path = "/spec/tolerations"
	} else if k == DeploymentKind {
		op.Path = "/spec/template/spec/tolerations"
	} else {
		return nil, errors.New("unsupported kind type error")
	}

	op.Value = t.Value
	return op, nil
}
