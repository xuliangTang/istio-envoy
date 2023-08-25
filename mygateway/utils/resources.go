package utils

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"istio-envoy/mygateway/utils/helpers"
	v1 "k8s.io/api/networking/v1"
	"log"
	"sigs.k8s.io/yaml"
)

const (
	TestIngress = "mygateway/test/ingress.yaml"
)

var (
	clusters, routeConfigs, listeners []types.Resource
)

func NewEmptySnapshot(version string) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {},
			resource.RouteType:    {},
			resource.ListenerType: {},
		},
	)
	return snap
}

func AddSnapshot(version string, ingress *v1.Ingress) *cache.Snapshot {
	// 渲染listener
	lis := &listener.Listener{}
	err := NewTplGenerator[*listener.Listener]().
		GetOutput(ingress, "listener", lis)
	if err != nil {
		log.Fatalln(err)
	}
	listeners = append(listeners, lis)

	// 渲染routeConfig
	rc := &route.RouteConfiguration{}
	err = NewTplGenerator[*route.RouteConfiguration]().
		GetOutput(ingress, "route", rc)
	if err != nil {
		log.Fatalln(err)
	}
	routeConfigs = append(routeConfigs, rc)

	// 渲染clusters
	cls, err := NewTplGenerator[*cluster.Cluster]().
		GetOutputs(ingress, "clusters", func() *cluster.Cluster {
			return &cluster.Cluster{}
		})
	if err != nil {
		log.Fatalln(err)
	}
	for _, c := range cls {
		clusters = append(clusters, c)
	}

	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  clusters,
			resource.RouteType:    routeConfigs,
			resource.ListenerType: listeners,
		},
	)
	return snap
}

// GenerateSnapshot 解析测试yaml生成快照
func GenerateSnapshot(version string) *cache.Snapshot {
	// 把ingress yaml反序列化为对象
	ingBytes := helpers.MustLoadFile(TestIngress)
	ingress := &v1.Ingress{}
	if err := yaml.Unmarshal(ingBytes, ingress); err != nil {
		log.Fatalln(err)
	}

	// 渲染listener
	lis := &listener.Listener{}
	err := NewTplGenerator[*listener.Listener]().
		GetOutput(ingress, "listener", lis)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染routeConfig
	routeConfig := &route.RouteConfiguration{}
	err = NewTplGenerator[*route.RouteConfiguration]().
		GetOutput(ingress, "route", routeConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染clusters
	var resList []types.Resource
	clusters, err := NewTplGenerator[*cluster.Cluster]().
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
