package overridemanager

import (
	"encoding/json"
	"sort"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	policyv1alpha1 "github.com/k-cloud-labs/pkg/apis/policy/v1alpha1"
	"github.com/k-cloud-labs/pkg/client/listers/policy/v1alpha1"
	"github.com/k-cloud-labs/pkg/util"
)

// OverrideManager managers override policies operation
type OverrideManager interface {
	// ApplyOverridePolicies overrides the object if one or more matched override policies exist.
	// For cluster scoped resource:
	// - Apply ClusterOverridePolicy by policies name in ascending
	// For namespaced scoped resource, apply order is:
	// - First apply ClusterOverridePolicy;
	// - Then apply OverridePolicy;
	ApplyOverridePolicies(rawObj *unstructured.Unstructured, operation string) (appliedCOPs *AppliedOverrides, appliedOPs *AppliedOverrides, err error)
}

// GeneralOverridePolicy is an abstract object of ClusterOverridePolicy and OverridePolicy
type GeneralOverridePolicy interface {
	// GetName returns the name of OverridePolicy
	GetName() string
	// GetNamespace returns the namespace of OverridePolicy
	GetNamespace() string
	// GetOverridePolicySpec returns the OverridePolicySpec of OverridePolicy
	GetOverridePolicySpec() policyv1alpha1.OverridePolicySpec
}

// overrideOption define the JSONPatch operator
type overrideOption struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type policyOverriders struct {
	name       string
	namespace  string
	overriders policyv1alpha1.Overriders
}

type overrideManagerImpl struct {
	opLister  v1alpha1.OverridePolicyLister
	copLister v1alpha1.ClusterOverridePolicyLister
}

func NewOverrideManager(copLister v1alpha1.ClusterOverridePolicyLister, opLister v1alpha1.OverridePolicyLister) OverrideManager {
	return &overrideManagerImpl{
		opLister:  opLister,
		copLister: copLister,
	}
}

func (o *overrideManagerImpl) ApplyOverridePolicies(rawObj *unstructured.Unstructured, operation string) (*AppliedOverrides, *AppliedOverrides, error) {
	var (
		appliedCOPs *AppliedOverrides
		appliedOPs  *AppliedOverrides
		err         error
	)

	appliedCOPs, err = o.applyClusterOverridePolicies(rawObj, operation)
	if err != nil {
		klog.Errorf("Failed to apply cluster override policies. error: %v", err)
		return nil, nil, err
	}

	if rawObj.GetNamespace() != "" {
		// Apply namespace scoped override policies
		appliedOPs, err = o.applyOverridePolicies(rawObj, operation)
		if err != nil {
			klog.Errorf("Failed to apply namespaced override policies. error: %v", err)
			return nil, nil, err
		}
	}

	return appliedCOPs, appliedOPs, nil
}

func (o *overrideManagerImpl) applyClusterOverridePolicies(rawObj *unstructured.Unstructured, operation string) (*AppliedOverrides, error) {
	cops, err := o.copLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list cluster override policies, error: %v", err)
		return nil, err
	}

	if len(cops) == 0 {
		return nil, nil
	}

	items := make([]GeneralOverridePolicy, 0, len(cops))
	for i := range cops {
		items = append(items, cops[i])
	}

	matchingPolicyOverriders := o.getOverridersFromOverridePolicies(items, rawObj, operation)
	if len(matchingPolicyOverriders) == 0 {
		klog.V(2).Infof("No cluster override policy for resource: %s/%s", rawObj.GetNamespace(), rawObj.GetName())
		return nil, nil
	}

	appliedOverrides := &AppliedOverrides{}
	for _, p := range matchingPolicyOverriders {
		if err := applyPolicyOverriders(rawObj, p.overriders); err != nil {
			klog.Errorf("Failed to apply cluster overriders(%s) for resource(%s/%s), error: %v", p.name, rawObj.GetNamespace(), rawObj.GetName(), err)
			return nil, err
		}
		klog.V(2).Infof("Applied cluster overriders(%s) for resource(%s/%s)", p.name, rawObj.GetNamespace(), rawObj.GetName())
		appliedOverrides.Add(p.name, p.overriders)
	}

	return appliedOverrides, nil
}

func (o *overrideManagerImpl) applyOverridePolicies(rawObj *unstructured.Unstructured, operation string) (*AppliedOverrides, error) {
	ops, err := o.opLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list override policies from namespace: %s, error: %v", rawObj.GetNamespace(), err)
		return nil, err
	}

	if len(ops) == 0 {
		return nil, nil
	}

	items := make([]GeneralOverridePolicy, 0, len(ops))
	for i := range ops {
		items = append(items, ops[i])
	}

	matchingPolicyOverriders := o.getOverridersFromOverridePolicies(items, rawObj, operation)
	if len(matchingPolicyOverriders) == 0 {
		klog.V(2).Infof("No override policy for resource(%s/%s)", rawObj.GetNamespace(), rawObj.GetName())
		return nil, nil
	}

	appliedOverriders := &AppliedOverrides{}
	for _, p := range matchingPolicyOverriders {
		if err := applyPolicyOverriders(rawObj, p.overriders); err != nil {
			klog.Errorf("Failed to apply overriders(%s/%s) for resource(%s/%s), error: %v", p.namespace, p.name, rawObj.GetNamespace(), rawObj.GetName(), err)
			return nil, err
		}
		klog.V(2).Infof("Applied overriders(%s/%s) for resource(%s/%s)", p.namespace, p.name, rawObj.GetNamespace(), rawObj.GetName())
		appliedOverriders.Add(p.name, p.overriders)
	}

	return appliedOverriders, nil
}

func (o *overrideManagerImpl) getOverridersFromOverridePolicies(policies []GeneralOverridePolicy, resource *unstructured.Unstructured, operation string) []policyOverriders {
	resourceMatchingPolicies := make([]GeneralOverridePolicy, 0)

	for _, policy := range policies {
		if policy.GetOverridePolicySpec().ResourceSelectors == nil {
			resourceMatchingPolicies = append(resourceMatchingPolicies, policy)
			continue
		}

		if util.ResourceMatchSelectors(resource, policy.GetOverridePolicySpec().ResourceSelectors...) {
			resourceMatchingPolicies = append(resourceMatchingPolicies, policy)
		}
	}

	matchingPolicyOverriders := make([]policyOverriders, 0)

	for _, policy := range resourceMatchingPolicies {
		for _, rule := range policy.GetOverridePolicySpec().OverrideRules {
			if len(rule.TargetOperations) == 0 || operationMatches(rule.TargetOperations, operation) {
				matchingPolicyOverriders = append(matchingPolicyOverriders, policyOverriders{
					name:       policy.GetName(),
					namespace:  policy.GetNamespace(),
					overriders: rule.Overriders,
				})
			}
		}
	}

	sort.Slice(matchingPolicyOverriders, func(i, j int) bool {
		return matchingPolicyOverriders[i].name < matchingPolicyOverriders[j].name
	})

	return matchingPolicyOverriders
}

// applyPolicyOverriders applies OverridePolicy/ClusterOverridePolicy overriders to target object
func applyPolicyOverriders(rawObj *unstructured.Unstructured, overriders policyv1alpha1.Overriders) error {
	if err := executeCue(rawObj, overriders.Cue); err != nil {
		return err
	}

	return applyJSONPatch(rawObj, parseJSONPatchesByPlaintext(overriders.Plaintext))
}

// applyJSONPatch applies the override on to the given unstructured object.
func applyJSONPatch(obj *unstructured.Unstructured, overrides []overrideOption) error {
	jsonPatchBytes, err := json.Marshal(overrides)
	if err != nil {
		return err
	}

	patch, err := jsonpatch.DecodePatch(jsonPatchBytes)
	if err != nil {
		return err
	}

	objectJSONBytes, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	patchedObjectJSONBytes, err := patch.Apply(objectJSONBytes)
	if err != nil {
		return err
	}

	err = obj.UnmarshalJSON(patchedObjectJSONBytes)
	return err
}

func parseJSONPatchesByPlaintext(overriders []policyv1alpha1.PlaintextOverrider) []overrideOption {
	patches := make([]overrideOption, 0, len(overriders))
	for i := range overriders {
		patches = append(patches, overrideOption{
			Op:    string(overriders[i].Operator),
			Path:  overriders[i].Path,
			Value: overriders[i].Value,
		})
	}
	return patches
}

func operationMatches(operations []string, target string) bool {
	for _, op := range operations {
		if op == target {
			return true
		}
	}

	return false
}

func executeCue(rawObj *unstructured.Unstructured, cue string) error {
	// todo: support cue
	return nil
}