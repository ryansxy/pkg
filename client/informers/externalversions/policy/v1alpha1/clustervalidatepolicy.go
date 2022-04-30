/*
Copyright 2022 by k-cloud-labs org.

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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	policyv1alpha1 "github.com/k-cloud-labs/pkg/apis/policy/v1alpha1"
	versioned "github.com/k-cloud-labs/pkg/client/clientset/versioned"
	internalinterfaces "github.com/k-cloud-labs/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/k-cloud-labs/pkg/client/listers/policy/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterValidatePolicyInformer provides access to a shared informer and lister for
// ClusterValidatePolicies.
type ClusterValidatePolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterValidatePolicyLister
}

type clusterValidatePolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewClusterValidatePolicyInformer constructs a new informer for ClusterValidatePolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterValidatePolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterValidatePolicyInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredClusterValidatePolicyInformer constructs a new informer for ClusterValidatePolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterValidatePolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PolicyV1alpha1().ClusterValidatePolicies().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PolicyV1alpha1().ClusterValidatePolicies().Watch(context.TODO(), options)
			},
		},
		&policyv1alpha1.ClusterValidatePolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterValidatePolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterValidatePolicyInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterValidatePolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&policyv1alpha1.ClusterValidatePolicy{}, f.defaultInformer)
}

func (f *clusterValidatePolicyInformer) Lister() v1alpha1.ClusterValidatePolicyLister {
	return v1alpha1.NewClusterValidatePolicyLister(f.Informer().GetIndexer())
}
