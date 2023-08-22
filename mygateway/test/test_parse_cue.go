package main

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"fmt"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/cache/v3" // 要加这个引入，否则下面jsonpb.unmarshal会报错
	"github.com/golang/protobuf/jsonpb"
	"github.com/tidwall/gjson"
	"istio-envoy/mygateway/utils/helpers"
	"log"
)

const input = `
{
   "listen_port": 8080
}
`

func main() {
	cc := cuecontext.New()
	cv := cc.CompileBytes(helpers.MustLoadFile("mygateway/tpls/xds.cue"))
	inputCv := cc.CompileString(input)
	cv.FillPath(cue.ParsePath("input"), inputCv)
	b, err := cv.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染listener
	listenerJson := gjson.Get(string(b), "listener")
	lis := &listener.Listener{}
	if err = jsonpb.UnmarshalString(listenerJson.String(), lis); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(lis)

	// 渲染route_config
	routeJson := gjson.Get(string(b), "route")
	routeObj := &route.RouteConfiguration{}
	if err = jsonpb.UnmarshalString(routeJson.String(), routeObj); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(routeObj)

	// 渲染clusters
	clustersJson := gjson.Get(string(b), "clusters")
	var clusters []*cluster.Cluster
	for _, item := range clustersJson.Array() {
		clusterObj := &cluster.Cluster{}
		if err = jsonpb.UnmarshalString(item.String(), clusterObj); err != nil {
			log.Fatalln(err)
		}
		clusters = append(clusters, clusterObj)
	}
	fmt.Println(clusters)
}
