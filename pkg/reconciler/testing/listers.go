/*
Copyright 2020 The Knative Authors

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

package ingress

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	networking "knative.dev/networking/pkg/apis/networking/v1alpha1"
	fakeservingclientset "knative.dev/networking/pkg/client/clientset/versioned/fake"
	networkinglisters "knative.dev/networking/pkg/client/listers/networking/v1alpha1"
	"knative.dev/pkg/reconciler/testing"

	fakegatewayapiclientset "github.com/nak3/net-gateway-api/pkg/client/gatewayapi/clientset/versioned/fake"
	gwlisters "github.com/nak3/net-gateway-api/pkg/client/gatewayapi/listers/apis/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeservingclientset.AddToScheme,
	fakegatewayapiclientset.AddToScheme,
	fakekubeclientset.AddToScheme,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewListers(objs []runtime.Object) Listers {
	scheme := NewScheme()

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}
	return scheme
}

func (*Listers) NewScheme() *runtime.Scheme {
	return NewScheme()
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetNetworkingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeservingclientset.AddToScheme)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l *Listers) GetGatewayAPIObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakegatewayapiclientset.AddToScheme)
}

// GetIngressLister get lister for Ingress resource.
func (l *Listers) GetIngressLister() networkinglisters.IngressLister {
	return networkinglisters.NewIngressLister(l.IndexerFor(&networking.Ingress{}))
}

// GetHTTPRouteLister get lister for HTTPProxy resource.
func (l *Listers) GetHTTPRouteLister() gwlisters.HTTPRouteLister {
	return gwlisters.NewHTTPRouteLister(l.IndexerFor(&gwv1alpha1.HTTPRoute{}))
}

// GetGatewayLister get lister for HTTPProxy resource.
func (l *Listers) GetGatewayLister() gwlisters.GatewayLister {
	return gwlisters.NewGatewayLister(l.IndexerFor(&gwv1alpha1.Gateway{}))
}

// GetEndpointsLister get lister for K8s Endpoints resource.
func (l *Listers) GetEndpointsLister() corev1listers.EndpointsLister {
	return corev1listers.NewEndpointsLister(l.IndexerFor(&corev1.Endpoints{}))
}
