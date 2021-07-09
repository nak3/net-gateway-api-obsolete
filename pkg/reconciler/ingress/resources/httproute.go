package resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	"knative.dev/networking/pkg/apis/networking"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/kmeta"
)

// MakeHTTPRoute creates HTTPRoute to set up routing rules.
func MakeHTTPRoute(
	ctx context.Context,
	ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
	gateway gwv1alpha1.Gateway,
) (*gwv1alpha1.HTTPRoute, error) {

	spec, err := makeHTTPRouteSpec(ctx, ing, rule, gateway)
	if err != nil {
		return nil, err
	}

	return &gwv1alpha1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HTTPRouteName(rule.Hosts),
			Namespace: ing.Namespace,
			Labels: kmeta.UnionMaps(ing.Labels, map[string]string{
				HTTPRouteNamespaceLabelKey: ing.Namespace,
				HTTPRouteVisibilityKey:     Visibility(rule.Visibility),
				// Do not usee HTTPRoute name as it exceeds 63 byte.
				networking.IngressLabelKey: ing.Name,
			}),
			Annotations: kmeta.FilterMap(ing.GetAnnotations(), func(key string) bool {
				return key == corev1.LastAppliedConfigAnnotation
			}),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(ing)},
		},
		Spec: spec,
	}, nil
}

func makeHTTPRouteSpec(
	ctx context.Context,
	ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
	gateway gwv1alpha1.Gateway,
) (gwv1alpha1.HTTPRouteSpec, error) {

	hostnames := []gwv1alpha1.Hostname{}
	for _, hostname := range rule.Hosts {
		hostnames = append(hostnames, gwv1alpha1.Hostname(hostname))
	}

	rules := makeHTTPRouteRule(rule)

	gatewayRef := gwv1alpha1.GatewayReference{
		Namespace: gateway.Namespace,
		Name:      gateway.Name,
	}

	return gwv1alpha1.HTTPRouteSpec{
		Hostnames: hostnames,
		Rules:     rules,
		Gateways: gwv1alpha1.RouteGateways{
			Allow:       gwv1alpha1.GatewayAllowFromList,
			GatewayRefs: []gwv1alpha1.GatewayReference{gatewayRef},
		},
	}, nil
}

func makeHTTPRouteRule(rule *netv1alpha1.IngressRule) []gwv1alpha1.HTTPRouteRule {
	rules := []gwv1alpha1.HTTPRouteRule{}

	for _, path := range rule.HTTP.Paths {
		var forwards []gwv1alpha1.HTTPRouteForwardTo
		var preFilters []gwv1alpha1.HTTPRouteFilter
		if path.AppendHeaders != nil {
			preFilters = []gwv1alpha1.HTTPRouteFilter{{
				Type: gwv1alpha1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gwv1alpha1.HTTPRequestHeaderFilter{
					Set: path.AppendHeaders,
				}}}
		}

		for _, split := range path.Splits {
			name := split.IngressBackend.ServiceName
			forward := gwv1alpha1.HTTPRouteForwardTo{
				Port:        portNumPtr(split.ServicePort.IntValue()),
				ServiceName: &name,
				Weight:      int32(split.Percent),
				Filters: []gwv1alpha1.HTTPRouteFilter{{
					Type: gwv1alpha1.HTTPRouteFilterRequestHeaderModifier,
					RequestHeaderModifier: &gwv1alpha1.HTTPRequestHeaderFilter{
						Set: split.AppendHeaders,
					}},
				}}
			forwards = append(forwards, forward)
		}

		rule := gwv1alpha1.HTTPRouteRule{
			ForwardTo: forwards,
			Filters:   preFilters,
		}
		rules = append(rules, rule)
	}
	return rules
}

func portNumPtr(port int) *gwv1alpha1.PortNumber {
	pn := gwv1alpha1.PortNumber(port)
	return &pn
}

// HTTPRouteName returns the name for the HTTPRoute
// for the given host and visibility.
func HTTPRouteName(hosts []string) string {
	return hosts[len(hosts)-1]
}
