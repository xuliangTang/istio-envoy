package utils

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"istio-envoy/mygateway/tpls"
	"log"
)

const input = `
{
   "listen_port": 8080
}
`

// GenerateSnapshot 解析cue生成快照
func GenerateSnapshot(version string) *cache.Snapshot {
	lis := &listener.Listener{}
	err := tpls.NewTplGenerator[*listener.Listener]().
		GetOutput(input, "listener", false, lis)
	if err != nil {
		log.Fatalln(err)
	}

	routeConfig := &route.RouteConfiguration{}
	err = tpls.NewTplGenerator[*route.RouteConfiguration]().
		GetOutput(input, "route", false, routeConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// 构建snapshot
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.RouteType:    {routeConfig},
			resource.ListenerType: {lis},
		},
	)
	return snap
}
