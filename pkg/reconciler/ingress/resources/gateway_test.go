/*
Copyright 2021 The Knative Authors

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

package resources

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nak3/net-gateway-api/pkg/reconciler/ingress/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/reconciler"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

const (
	testNamespace    = "test-ns"
	testIngressName  = "test-ingress"
	testGatewayClass = "test-class"
)

var (
	externalHost      = gwv1alpha1.Hostname(testHosts[0])
	localHostShortest = gwv1alpha1.Hostname(testLocalHosts[0])
	localHostShort    = gwv1alpha1.Hostname(testLocalHosts[1])
	localHostFull     = gwv1alpha1.Hostname(testLocalHosts[2])

	testLocalHosts = []string{
		"hello-example.default",
		"hello-example.default.svc",
		"hello-example.default.svc.cluster.local",
	}
	testHosts = []string{"hello-example.default.example.com"}

	route = gwv1alpha1.RouteBindingSelector{
		Namespaces: gwv1alpha1.RouteNamespaces{
			From: gwv1alpha1.RouteSelectAll,
		},
		Selector: metav1.LabelSelector{MatchLabels: map[string]string{
			HTTPRouteNamespaceLabelKey: testNamespace,
			HTTPRouteVisibilityKey:     "",
			networking.IngressLabelKey: testIngressName,
		}},
		Kind: "HTTPRoute",
	}
	localRoute = gwv1alpha1.RouteBindingSelector{
		Namespaces: gwv1alpha1.RouteNamespaces{
			From: gwv1alpha1.RouteSelectAll,
		},
		Selector: metav1.LabelSelector{MatchLabels: map[string]string{
			HTTPRouteNamespaceLabelKey: testNamespace,
			HTTPRouteVisibilityKey:     "cluster-local",
			networking.IngressLabelKey: testIngressName,
		}},
		Kind: "HTTPRoute",
	}
)

type testConfigStore struct {
	config *config.Config
}

func (t *testConfigStore) ToContext(ctx context.Context) context.Context {
	return config.ToContext(ctx, t.config)
}

var testConfig = &config.Config{
	Gateway: &config.Gateway{
		Gateways: map[string]*config.GatewayConfig{
			"": {
				GatewayClass: testGatewayClass,
			},
			"cluster-local": {
				GatewayClass: testGatewayClass,
			},
		}},
}

var _ reconciler.ConfigStore = (*testConfigStore)(nil)

func TestMakeGateway(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ci       *v1alpha1.Ingress
		expected []*gwv1alpha1.Gateway
	}{
		{
			name: "ingress with single rule",
			ci: &v1alpha1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testIngressName,
					Namespace: testNamespace,
					Labels: map[string]string{
						networking.IngressLabelKey: "test-ingress",
					},
				},
				Spec: v1alpha1.IngressSpec{Rules: []v1alpha1.IngressRule{{
					Hosts:      testHosts,
					Visibility: v1alpha1.IngressVisibilityExternalIP,
					HTTP:       &v1alpha1.HTTPIngressRuleValue{},
				}}},
			},
			expected: []*gwv1alpha1.Gateway{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      LongestHost(testHosts),
					Namespace: testNamespace,
					Labels: map[string]string{
						networking.IngressLabelKey: testIngressName,
					},
					Annotations: map[string]string{},
				},
				Spec: gwv1alpha1.GatewaySpec{
					GatewayClassName: testGatewayClass,
					Listeners: []gwv1alpha1.Listener{{
						Hostname: &externalHost,
						Port:     gwv1alpha1.PortNumber(80),
						Protocol: gwv1alpha1.HTTPProtocolType,
						Routes:   route,
					}},
				},
			}},
		}, {
			name: "ingress with multiple visibility",
			ci: &v1alpha1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testIngressName,
					Namespace: testNamespace,
					Labels: map[string]string{
						networking.IngressLabelKey: "test-ingress",
					},
				},
				Spec: v1alpha1.IngressSpec{Rules: []v1alpha1.IngressRule{
					{
						Hosts:      testHosts,
						Visibility: v1alpha1.IngressVisibilityExternalIP,
						HTTP:       &v1alpha1.HTTPIngressRuleValue{},
					}, {
						Hosts:      testLocalHosts,
						Visibility: v1alpha1.IngressVisibilityClusterLocal,
						HTTP:       &v1alpha1.HTTPIngressRuleValue{},
					},
				}},
			},
			expected: []*gwv1alpha1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      LongestHost(testHosts),
						Namespace: testNamespace,
						Labels: map[string]string{
							networking.IngressLabelKey: testIngressName,
						},
						Annotations: map[string]string{},
					},
					Spec: gwv1alpha1.GatewaySpec{
						GatewayClassName: testGatewayClass,
						Listeners: []gwv1alpha1.Listener{{
							Hostname: &externalHost,
							Port:     gwv1alpha1.PortNumber(80),
							Protocol: gwv1alpha1.HTTPProtocolType,
							Routes:   route,
						}},
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      LongestHost(testLocalHosts),
						Namespace: testNamespace,
						Labels: map[string]string{
							networking.IngressLabelKey: testIngressName,
						},
						Annotations: map[string]string{},
					},
					Spec: gwv1alpha1.GatewaySpec{
						GatewayClassName: testGatewayClass,
						Listeners: []gwv1alpha1.Listener{{
							Hostname: &localHostShortest,
							Port:     gwv1alpha1.PortNumber(80),
							Protocol: gwv1alpha1.HTTPProtocolType,
							Routes:   localRoute,
						}, {
							Hostname: &localHostShort,
							Port:     gwv1alpha1.PortNumber(80),
							Protocol: gwv1alpha1.HTTPProtocolType,
							Routes:   localRoute,
						}, {
							Hostname: &localHostFull,
							Port:     gwv1alpha1.PortNumber(80),
							Protocol: gwv1alpha1.HTTPProtocolType,
							Routes:   localRoute,
						}},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tcs := &testConfigStore{config: testConfig}
			ctx := tcs.ToContext(context.Background())
			for i, rule := range tc.ci.Spec.Rules {
				rule := rule
				gw, err := MakeGateway(ctx, tc.ci, &rule)
				if err != nil {
					t.Fatal("MakeGateway failed:", err)
				}
				if diff := cmp.Diff(tc.expected[i], gw); diff != "" {
					t.Error("Unexpected metadata (-want +got):", diff)
				}
			}
		})
	}
}
