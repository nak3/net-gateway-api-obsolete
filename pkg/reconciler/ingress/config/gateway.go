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

package config

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/network"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

const (
	// GatewayConfigName is the config map name for the gateway configuration.
	GatewayConfigName = "config-gateway"

	// DefaultGatewayClass is the gatewayclass name for the gateway.
	DefaultGatewayClass = "istio"

	// DefaultVisibilityKey is the key for config-gateway for the gateway configuration.
	DefaultVisibilityKey = "default"

	// DefaultIstioNamespace is the default namespace for the gateway.
	// The namespace of the gateway should not be required,
	// but gateway-api with Istio needs to deploy Gateway in the same namespace with istio service.
	DefaultIstioNamespace = "istio-system"
)

var (
	// DefaultLocalGatewayService holds the default local gateway service address.
	// Placeholder service points to the service.
	DefaultLocalGatewayService = "knative-local-gateway.istio-system.svc." + network.GetClusterDomainName()

	// DefaultPublicGatewayService is the default gateway service address.
	DefaultPublicGatewayService = "istio-ingressgateway.istio-system.svc." + network.GetClusterDomainName()
)

type GatewayConfig struct {
	GatewayClass string `json:"gatewayClass,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Address      string `json:"address,omitempty"`
}

// Gateway maps gateways to routes by matching the gateway's
// label selectors to the route's labels.
type Gateway struct {
	// Gateways map from gateway to label selector.  If a route has
	// labels matching a particular selector, it will use the
	// corresponding gateway.  If multiple selectors match, we choose
	// the most specific selector.
	Gateways map[string]*GatewayConfig
}

// NewGatewayFromConfigMap creates a Gateway from the supplied ConfigMap
func NewGatewayFromConfigMap(configMap *corev1.ConfigMap) (*Gateway, error) {
	c := Gateway{Gateways: map[string]*GatewayConfig{}}
	for k, v := range configMap.Data {
		if k == configmap.ExampleKey {
			continue
		}
		if k == DefaultVisibilityKey {
			k = ""
		}
		config := &GatewayConfig{}
		err := yaml.Unmarshal([]byte(v), config)
		if err != nil {
			return nil, err
		}
		c.Gateways[k] = config
	}
	// Add default gateway if empty key is not defined.
	if _, ok := c.Gateways[""]; !ok {
		c.Gateways[""] = &GatewayConfig{GatewayClass: DefaultGatewayClass, Namespace: DefaultIstioNamespace, Address: DefaultPublicGatewayService}
	}

	// Add default local gateway if cluster-local gateway is not defined.
	if _, ok := c.Gateways["cluster-local"]; !ok {
		c.Gateways["cluster-local"] = &GatewayConfig{GatewayClass: DefaultGatewayClass, Namespace: DefaultIstioNamespace, Address: DefaultLocalGatewayService}
	}
	return &c, nil
}

// LookupGatewayNamespace returns a gateway namespace given a visibility config.
func (c *Gateway) LookupGatewayNamespace(visibility string) string {
	if c.Gateways[visibility] == nil {
		return ""
	}
	return c.Gateways[visibility].Namespace
}

// LookupGatewayClass returns a gatewayclass given a visibility config.
func (c *Gateway) LookupGatewayClass(visibility string) string {
	if c.Gateways[visibility] == nil {
		// TODO: empty gatewayclass should be error?
		return ""
	}
	return c.Gateways[visibility].GatewayClass
}

// LookupAddress returns a gateway address given a visibility config.
// TODO: LookupGatewayServiceAddress ?
func (c *Gateway) LookupAddress(visibility string) string {
	if c.Gateways[visibility] == nil {
		return ""
	}
	return c.Gateways[visibility].Address
}
