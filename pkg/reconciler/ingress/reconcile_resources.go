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

package ingress

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/controller"

	"github.com/nak3/net-gateway-api/pkg/reconciler/ingress/config"
	"github.com/nak3/net-gateway-api/pkg/reconciler/ingress/resources"
)

// reconcileHTTPRoute reconciles HTTPRoute.
func (c *Reconciler) reconcileHTTPRoute(
	ctx context.Context, ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
	gateway gwv1alpha1.Gateway,
) (*gwv1alpha1.HTTPRoute, error) {
	recorder := controller.GetEventRecorder(ctx)

	httproute, err := c.httprouteLister.HTTPRoutes(ing.Namespace).Get(resources.LongestHost(rule.Hosts))
	if apierrs.IsNotFound(err) {
		desired, err := resources.MakeHTTPRoute(ing, rule, gateway)
		if err != nil {
			return nil, err
		}
		httproute, err = c.gwapiclient.NetworkingV1alpha1().HTTPRoutes(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			recorder.Eventf(ing, corev1.EventTypeWarning, "CreationFailed", "Failed to create HTTPRoute: %v", err)
			return nil, fmt.Errorf("failed to create HTTPRoute: %w", err)
		}

		recorder.Eventf(ing, corev1.EventTypeNormal, "Created", "Created HTTPRoute %q", httproute.GetName())
		return httproute, nil
	} else if err != nil {
		return nil, err
	} else {
		desired, err := resources.MakeHTTPRoute(ing, rule, gateway)
		if err != nil {
			return nil, err
		}

		if !equality.Semantic.DeepEqual(httproute.Spec, desired.Spec) ||
			!equality.Semantic.DeepEqual(httproute.Annotations, desired.Annotations) ||
			!equality.Semantic.DeepEqual(httproute.Labels, desired.Labels) {

			// Don't modify the informers copy.
			origin := httproute.DeepCopy()
			origin.Spec = desired.Spec
			origin.Annotations = desired.Annotations
			origin.Labels = desired.Labels

			updated, err := c.gwapiclient.NetworkingV1alpha1().HTTPRoutes(origin.Namespace).Update(
				ctx, origin, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update HTTPRoute: %w", err)
			}
			return updated, nil
		}
	}

	return httproute, err
}

// reconcileGateway reconciles Gateway.
func (c *Reconciler) reconcileGateway(
	ctx context.Context, ing *netv1alpha1.Ingress,
	rule *netv1alpha1.IngressRule,
) (*gwv1alpha1.Gateway, error) {
	recorder := controller.GetEventRecorder(ctx)

	visibility := resources.Visibility(rule.Visibility)
	gatewayConfig := config.FromContext(ctx).Gateway
	ns := gatewayConfig.LookupGatewayNamespace(visibility)
	if ns == "" {
		ns = ing.Namespace
	}

	gateway, err := c.gatewayLister.Gateways(ns).Get(resources.LongestHost(rule.Hosts))
	if apierrs.IsNotFound(err) {
		desired, err := resources.MakeGateway(ctx, ing, rule)
		if err != nil {
			return nil, err
		}
		gateway, err = c.gwapiclient.NetworkingV1alpha1().Gateways(ns).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			recorder.Eventf(ing, corev1.EventTypeWarning, "CreationFailed", "Failed to create Gateway: %v", err)
			return nil, fmt.Errorf("failed to create Gateway: %w", err)
		}

		recorder.Eventf(ing, corev1.EventTypeNormal, "Created", "Created Gateway %q", gateway.GetName())
		return gateway, nil
	} else if err != nil {
		return nil, err
	} else {
		// TODO: namespace change
		desired, err := resources.MakeGateway(ctx, ing, rule)
		if err != nil {
			return nil, err
		}

		if !equality.Semantic.DeepEqual(gateway.Spec, desired.Spec) ||
			!equality.Semantic.DeepEqual(gateway.Annotations, desired.Annotations) ||
			!equality.Semantic.DeepEqual(gateway.Labels, desired.Labels) {

			// Don't modify the informers copy.
			origin := gateway.DeepCopy()
			origin.Spec = desired.Spec
			origin.Annotations = desired.Annotations
			origin.Labels = desired.Labels

			updated, err := c.gwapiclient.NetworkingV1alpha1().Gateways(origin.Namespace).Update(
				ctx, origin, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update Gateway: %w", err)
			}
			return updated, nil
		}
	}
	// unreachable
	return nil, nil
}
