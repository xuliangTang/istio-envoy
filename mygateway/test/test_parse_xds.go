package main

import (
	"cuelang.org/go/cue"
	"fmt"
	"istio-envoy/mygateway/utils/helpers"
	v1 "k8s.io/api/networking/v1"
	"log"
	"sigs.k8s.io/yaml"
)

func main() {
	// 解析ingress
	ing := &v1.Ingress{}
	if err := yaml.Unmarshal(helpers.MustLoadFile("mygateway/test/ingress.yaml"), ing); err != nil {
		log.Fatalln(err)
	}

	// 读取xds模板
	xdsCv := helpers.MustLoadFileInstance("mygateway/tpls/xds.cue")

	// cue解析ingress时，IngressRuleValue是内嵌属性而且没有json tag，所以会取后面protobuf tag的第一段(bytes)
	// 所以解析后会多一段bytes: rules->bytes->http->paths
	ingCv := xdsCv.Context().Encode(ing)

	// 填充渲染
	retCv := xdsCv.FillPath(cue.ParsePath("input"), ingCv)
	retCv = retCv.LookupPath(cue.ParsePath("output"))

	helpers.SaveStr(fmt.Sprintf("%s", retCv), "./output.cue")
}

const test_ingress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-myservicea
  namespace: default
  annotations:
    envoy.ingress.kubernetes.io/rewrite-target: "/\\1"
    envoy.ingress.kubernetes.io/ratelimit-max: 5
    envoy.ingress.kubernetes.io/cors-enable: true
    envoy.ingress.kubernetes.io/cors-allow-methods: "GET, PUT, DELETE, POST, OPTIONS"
    envoy.ingress.kubernetes.io/cors-allow-origin: "*"
    envoy.ingress.kubernetes.io/cors-allow-origin-prefix: "abc"
    envoy.ingress.kubernetes.io/cors-allow-headers: "Authorization,test,ddd"
    envoy.ingress.kubernetes.io/cors-expose-headers: "sss"
    envoy.ingress.kubernetes.io/cors-max-age: 18000
    envoy.ingress.kubernetes.io/cors-allow-credentials: true
spec:
  rules:
    - host: myservicea.foo.org
      http:
        paths:
          - path: /v1/(.*?)
            pathType: Prefix
            backend:
              service:
                name: myservicea
                port:
                  number: 80
  ingressClassName: nginx
`
