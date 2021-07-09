/*
Copyright 2018 The Knative Authors

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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	"knative.dev/networking/pkg/apis/networking"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/kmeta"

	"github.com/nak3/net-gateway-api/pkg/reconciler/ingress/config"
)

const (
	HTTPRouteNamespaceLabelKey = networking.GroupName + "/httprouteNamespace"
	HTTPRouteVisibilityKey     = networking.GroupName + "/httprouteVisibility"
)

// MakeGateway creates Gateway to map to HTTPRoute.
func MakeGateway(
	ctx context.Context,
	ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
) (*gwv1alpha1.Gateway, error) {

	gatewayConfig := config.FromContext(ctx).Gateway
	visibility := Visibility(rule.Visibility)
	gatewayNamespace := gatewayConfig.LookupGatewayNamespace(visibility)
	if gatewayNamespace == "" {
		gatewayNamespace = ing.Namespace
	}

	return &gwv1alpha1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HTTPRouteName(rule.Hosts),
			Namespace: gatewayNamespace,
			Labels:    ing.Labels,
			Annotations: kmeta.FilterMap(ing.GetAnnotations(), func(key string) bool {
				return key == corev1.LastAppliedConfigAnnotation
			}),
		},
		Spec: makeGatewaySpec(ctx, ing, rule, gatewayConfig),
	}, nil
}

// TODO: name
func Visibility(visibility netv1alpha1.IngressVisibility) string {
	switch visibility {
	case netv1alpha1.IngressVisibilityClusterLocal:
		return "cluster-local"
	case netv1alpha1.IngressVisibilityExternalIP:
		return ""
	}
	return ""
}

func makeGatewaySpec(
	ctx context.Context,
	ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
	gwConfig *config.Gateway,
) gwv1alpha1.GatewaySpec {

	var listeners []gwv1alpha1.Listener
	for _, host := range rule.Hosts {
		host := gwv1alpha1.Hostname(host)
		route := gwv1alpha1.RouteBindingSelector{
			Namespaces: gwv1alpha1.RouteNamespaces{
				From: gwv1alpha1.RouteSelectAll,
			},
			Selector: metav1.LabelSelector{MatchLabels: map[string]string{
				HTTPRouteNamespaceLabelKey: ing.Namespace,
				HTTPRouteVisibilityKey:     Visibility(rule.Visibility),
				// Do not usee HTTPRoute name as it exceeds 63 byte.
				networking.IngressLabelKey: ing.Name,
			}},
			Kind: "HTTPRoute",
		}
		listeners = append(listeners, gwv1alpha1.Listener{
			Hostname: &host,
			Port:     gwv1alpha1.PortNumber(80),
			Protocol: gwv1alpha1.HTTPProtocolType,
			Routes:   route})
	}

	visibility := Visibility(rule.Visibility)
	gwSpec := gwv1alpha1.GatewaySpec{
		GatewayClassName: gwConfig.LookupGatewayClass(visibility),
		Listeners:        listeners,
	}
	if ad := gwConfig.LookupAddress(visibility); ad != "" {
		gwSpec.Addresses = []gwv1alpha1.GatewayAddress{{Type: gwv1alpha1.NamedAddressType, Value: ad}}
	}
	return gwSpec
}
