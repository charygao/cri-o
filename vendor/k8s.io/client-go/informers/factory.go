/*
Copyright The Kubernetes Authors.

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

package informers

import (
	reflect "reflect"
	sync "sync"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	admissionregistration "k8s.io/client-go/informers/admissionregistration"
	apps "k8s.io/client-go/informers/apps"
	auditregistration "k8s.io/client-go/informers/auditregistration"
	autoscaling "k8s.io/client-go/informers/autoscaling"
	batch "k8s.io/client-go/informers/batch"
	certificates "k8s.io/client-go/informers/certificates"
	coordination "k8s.io/client-go/informers/coordination"
	core "k8s.io/client-go/informers/core"
	events "k8s.io/client-go/informers/events"
	extensions "k8s.io/client-go/informers/extensions"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	networking "k8s.io/client-go/informers/networking"
	node "k8s.io/client-go/informers/node"
	policy "k8s.io/client-go/informers/policy"
	rbac "k8s.io/client-go/informers/rbac"
	scheduling "k8s.io/client-go/informers/scheduling"
	settings "k8s.io/client-go/informers/settings"
	storage "k8s.io/client-go/informers/storage"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

// SharedInformerOption defines the functional option type for SharedInformerFactory.
type SharedInformerOption func(*sharedInformerFactory) *sharedInformerFactory

type sharedInformerFactory struct {
	client           kubernetes.Interface
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	lock             sync.Mutex
	defaultResync    time.Duration
	customResync     map[reflect.Type]time.Duration

	informers map[reflect.Type]cache.SharedIndexInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[reflect.Type]bool
}

// WithCustomResyncConfig sets a custom resync period for the specified informer types.
func WithCustomResyncConfig(resyncConfig map[v1.Object]time.Duration) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		for k, v := range resyncConfig {
			factory.customResync[reflect.TypeOf(k)] = v
		}
		return factory
	}
}

// WithTweakListOptions sets a custom filter on all listers of the configured SharedInformerFactory.
func WithTweakListOptions(tweakListOptions internalinterfaces.TweakListOptionsFunc) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.tweakListOptions = tweakListOptions
		return factory
	}
}

// WithNamespace limits the SharedInformerFactory to the specified namespace.
func WithNamespace(namespace string) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.namespace = namespace
		return factory
	}
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all namespaces.
func NewSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

// NewFilteredSharedInformerFactory constructs a new instance of sharedInformerFactory.
// Listers obtained via this SharedInformerFactory will be subject to the same filters
// as specified here.
// Deprecated: Please use NewSharedInformerFactoryWithOptions instead
func NewFilteredSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync, WithNamespace(namespace), WithTweakListOptions(tweakListOptions))
}

// NewSharedInformerFactoryWithOptions constructs a new instance of a SharedInformerFactory with additional options.
func NewSharedInformerFactoryWithOptions(client kubernetes.Interface, defaultResync time.Duration, options ...SharedInformerOption) SharedInformerFactory {
	factory := &sharedInformerFactory{
		client:           client,
		namespace:        v1.NamespaceAll,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
		customResync:     make(map[reflect.Type]time.Duration),
	}

	// Apply all options
	for _, opt := range options {
		factory = opt(factory)
	}

	return factory
}

// Start initializes all requested informers.
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
			go informer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

// WaitForCacheSync waits for all started informers' cache were synced.
func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	informers := func() map[reflect.Type]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[reflect.Type]cache.SharedIndexInformer{}
		for informerType, informer := range f.informers {
			if f.startedInformers[informerType] {
				informers[informerType] = informer
			}
		}
		return informers
	}()

	res := map[reflect.Type]bool{}
	for informType, informer := range informers {
		res[informType] = cache.WaitForCacheSync(stopCh, informer.HasSynced)
	}
	return res
}

// InternalInformerFor returns the SharedIndexInformer for obj using an internal
// client.
func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}

	resyncPeriod, exists := f.customResync[informerType]
	if !exists {
		resyncPeriod = f.defaultResync
	}

	informer = newFunc(f.client, resyncPeriod)
	f.informers[informerType] = informer

	return informer
}

// SharedInformerFactory provides shared informers for resources in all known
// API group versions.
type SharedInformerFactory interface {
	internalinterfaces.SharedInformerFactory
	ForResource(resource schema.GroupVersionResource) (GenericInformer, error)
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	Admissionregistration() admissionregistration.Interface
	Apps() apps.Interface
	Auditregistration() auditregistration.Interface
	Autoscaling() autoscaling.Interface
	Batch() batch.Interface
	Certificates() certificates.Interface
	Coordination() coordination.Interface
	Core() core.Interface
	Events() events.Interface
	Extensions() extensions.Interface
	Networking() networking.Interface
	Node() node.Interface
	Policy() policy.Interface
	Rbac() rbac.Interface
	Scheduling() scheduling.Interface
	Settings() settings.Interface
	Storage() storage.Interface
}

func (f *sharedInformerFactory) Admissionregistration() admissionregistration.Interface {
	return admissionregistration.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Apps() apps.Interface {
	return apps.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Auditregistration() auditregistration.Interface {
	return auditregistration.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Autoscaling() autoscaling.Interface {
	return autoscaling.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Batch() batch.Interface {
	return batch.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Certificates() certificates.Interface {
	return certificates.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Coordination() coordination.Interface {
	return coordination.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Core() core.Interface {
	return core.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Events() events.Interface {
	return events.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Extensions() extensions.Interface {
	return extensions.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Networking() networking.Interface {
	return networking.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Node() node.Interface {
	return node.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Policy() policy.Interface {
	return policy.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Rbac() rbac.Interface {
	return rbac.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Scheduling() scheduling.Interface {
	return scheduling.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Settings() settings.Interface {
	return settings.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Storage() storage.Interface {
	return storage.New(f, f.namespace, f.tweakListOptions)
}
