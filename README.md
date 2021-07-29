# Knative Gateway API Controller

[![GoDoc](https://godoc.org/knative.dev/net-gateway-api-controller?status.svg)](https://godoc.org/knative.dev/gateway-api-controller)
[![Go Report Card](https://goreportcard.com/badge/knative/net-gateway-api-controller)](https://goreportcard.com/report/knative/gateway-api-controller)

Knative Gateway API Controller is a controller to generate Gateway resources based on Knative Ingress.

## Usage

#### Deploy net-gateway-api CRD

```
kubectl apply -k 'github.com/kubernetes-sigs/net-gateway-api/config/crd?ref=v0.2.0'
```

#### Install Istio (v1.10 or later)

```
istioctl install -y
```

#### Install net-gateway-api controller

```
ko resolve -f test/config/ -f config/ | kubectl apply -f -
```

#### Install Knative Serving

```
kubectl apply --filename https://storage.googleapis.com/knative-nightly/serving/latest/serving-crds.yaml
kubectl apply --filename https://storage.googleapis.com/knative-nightly/serving/latest/serving-core.yaml
```

#### Apply ingress class `gateway-api.ingress.networking.knative.dev`

```
kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  -p '{"data":{"ingress.class":"gateway-api.ingress.networking.knative.dev"}}'
```

#### Deploy Knative Service

```
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: hello-example
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        name: user-container
EOF
```

Now httproute is created.

```
$ kubectl get httproutes.networking.x-k8s.io
NAME                                      HOSTNAMES
hello-example.default.example.com         ["hello-example.default.example.com"]
hello-example.default.svc.cluster.local   ["hello-example.default","hello-example.default.svc","hello-example.default.svc.cluster.local"]
```

#### Access to the knative service

```
$ curl -H "Host: hello-example.default.example.com" 172.20.0.2:30348
Hello World!
```

__NOTE__ `172.20.0.2:30348` needs to be replaced with your `istio-ingressgateway.istio-system` endpoint.

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
