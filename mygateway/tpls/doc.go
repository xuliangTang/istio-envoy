package tpls

import (
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/cache/v3" // 要加这个引入，否则下面jsonpb.unmarshal会报错
)
