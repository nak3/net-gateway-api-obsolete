module github.com/nak3/net-gateway-api

go 1.15

require (
	go.uber.org/zap v1.17.0
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/code-generator v0.20.7
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/net-kourier v0.24.0
	knative.dev/networking v0.0.0-20210708015022-4e655b7fa1c3
	knative.dev/pkg v0.0.0-20210706174620-fe90576475ca
	knative.dev/serving v0.24.0
	sigs.k8s.io/gateway-api v0.2.0
	sigs.k8s.io/yaml v1.2.0
)
