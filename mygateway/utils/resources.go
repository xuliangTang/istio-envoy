package utils

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"istio-envoy/mygateway/tpls"
	"istio-envoy/mygateway/utils/helpers"
	v1 "k8s.io/api/networking/v1"
	"log"
	"sigs.k8s.io/yaml"
)

const (
	TestIngress = "mygateway/test/ingress.yaml"
)

// GenerateSnapshot 解析cue生成快照
func GenerateSnapshot(version string) *cache.Snapshot {
	// 把ingress yaml反序列化为对象
	ingBytes := helpers.MustLoadFile(TestIngress)
	ingress := &v1.Ingress{}
	if err := yaml.Unmarshal(ingBytes, ingress); err != nil {
		log.Fatalln(err)
	}

	// 渲染listener
	lis := &listener.Listener{}
	err := tpls.NewTplGenerator[*listener.Listener]().
		GetOutput(ingress, "listener", lis)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染routeConfig
	routeConfig := &route.RouteConfiguration{}
	err = tpls.NewTplGenerator[*route.RouteConfiguration]().
		GetOutput(ingress, "route", routeConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染clusters
	var resList []types.Resource
	clusters, err := tpls.NewTplGenerator[*cluster.Cluster]().
		GetOutputs(ingress, "clusters", func() *cluster.Cluster {
			return &cluster.Cluster{}
		})
	if err != nil {
		log.Fatalln(err)
	}
	for _, c := range clusters {
		resList = append(resList, c)
	}

	// 构建snapshot
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  resList,
			resource.RouteType:    {routeConfig},
			resource.ListenerType: {lis},
		},
	)
	return snap
}
