package main

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"fmt"
	"istio-envoy/mygateway/utils/helpers"
	v1 "k8s.io/api/networking/v1"
	"log"
	"sigs.k8s.io/yaml"
)

func main() {
	// 解析ingress
	ing := &v1.Ingress{}
	if err := yaml.Unmarshal([]byte(test_ingress), ing); err != nil {
		log.Fatalln(err)
	}

	// 读取xds模板
	xdsBytes := helpers.MustLoadFile("mygateway/tpls/xds.cue")
	cc := cuecontext.New()
	xdsCv := cc.CompileBytes(xdsBytes)

	// cue解析ingress时，IngressRuleValue是内嵌属性而且没有json tag，所以会取后面protobuf tag的第一段(bytes)
	// 所以解析后会多一段bytes: rules->bytes->http->paths
	ingCv := cc.Encode(ing)

	// 填充渲染
	retCv := xdsCv.FillPath(cue.ParsePath("input"), ingCv)
	retCv = retCv.LookupPath(cue.ParsePath("output"))

	helpers.SaveStr(fmt.Sprintf("%s", retCv), "./output.yaml")
}

const test_ingress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-myservicea
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: true
spec:
  rules:
    - host: myservicea.foo.org
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: myservicea
                port:
                  number: 80
  ingressClassName: nginx
`
