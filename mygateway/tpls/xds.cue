package tpls
import (
	"lain.com/mygateway/mygateway/tpls/common"
	"lain.com/mygateway/mygateway/tpls/filters/http/ratelimit"
)

input: {}

// 覆盖common里的值
comm: common & {
	annotations: input.metadata.annotations
}

output: {
	listener: {
		name: "0.0.0.0_8080"
		address: {
			 socket_address: {
			 	 address: "0.0.0.0",
		     port_value: "8080"
			 }
		}
		filter_chains: [
			 {
			   filters:
			   [
						{
							name: "envoy.filters.network.http_connection_manager"
							typed_config:{
								"@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
								stat_prefix: "gateway_http"
								codec_type: "AUTO"
								rds: {
									route_config_name: "gateway_route"	// 对应下方route的name
									config_source: {
										api_config_source: {
												api_type: "GRPC"
												transport_api_version: "V3"
												grpc_services: [
													{
															envoy_grpc:  {
															cluster_name: "xds_cluster"
														}
													}
												]
										}
									}
								}
								http_filters: [
									if comm.vars.cors != _|_ {
										{
											name: "envoy.filters.http.cors"
										}
									},
									if comm.ratelimit.max != _|_ && comm.ratelimit.max > 0 {	// 判断rps限流
										ratelimit.local_ratelimit
									},
									if comm.vars.lua_block != _|_ {
										name: "envoy.filters.http.lua"
										typed_config: {
											"@type": "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
											inline_code: comm.vars.lua_block
										}
									},
									{
										name: "envoy.filters.http.router"
									}
								]
							}
						}
		  	 ]
		  }
		]

		trafficDirection: "OUTBOUND"
	}

	route: {
		name: "gateway_route"
		virtual_hosts: [
				for _, rule in input.spec.rules {
						name: rule.host + "_name"
						domains: [rule.host],
						if comm.vars.cors != _|_ {  // 跨域设置
								cors: comm.vars.cors
						},
						routes: [
							for _, p in rule.bytes.http.paths {
								{
									 match: {
									 		if p.pathType == "Prefix" {
									 			if comm.vars.rewrite_value == _|_ {
									 				prefix: p.path
									 			}
									 			if comm.vars.rewrite_value != _|_ {	// 判断路径重写
									 				safe_regex:{
														 google_re2: {}
														 regex: p.path
											 	 	}
									 			}
									 		}

									 		if p.pathType == "Exact" {
												path: p.path
									 		}
									 }
									 route: {
										 	cluster: "OUTBOUND|" + p.backend.service.name + "|" + "\(p.backend.service.port.number)"
										  if comm.vars.rewrite_value != _|_ {
												 	regex_rewrite: {
												 		pattern: {
												 			google_re2: {}
                              regex: p.path
														}
                            substitution: comm.vars.rewrite_value
												 	}
											}
									 }
									 if comm.ratelimit.max != _|_ && comm.ratelimit.max > 0 {	// 判断rps限流
									 		typed_per_filter_config: {
											 		_f: ratelimit & {
											 			local_ratelimit_prefilter_config: {
											 		 		max_tokens:  comm.ratelimit.max
											 		 		tokens_per_fill: comm.ratelimit.perfill
											 		 		fill_interval:  comm.ratelimit.fillinteval
											 		 		prefix: rule.host
											 		 	}
											 		}
											 		_f.local_ratelimit_prefilter
											}
									 }

								}
							}
						]
				}
		]
	}

	clusters: [
			for _, rule in input.spec.rules {
				for _, p in rule.bytes.http.paths {
					{
							name: "OUTBOUND|" + p.backend.service.name + "|" + "\(p.backend.service.port.number)"
      				connect_timeout: "1s"
      				type: "STATIC"
      				lb_policy: "ROUND_ROBIN"
      				load_assignment: {
								cluster_name: "OUTBOUND|" + p.backend.service.name + "|" + "\(p.backend.service.port.number)"
								endpoints: [
									{
										lb_endpoints: [
											{
												endpoint: {
													address: {
															socket_address: {
																	address: "172.17.0.2"
																	port_value: 80
															}
													}
												}
											}]
									}]
					    }
					}
				}
			}
  ]

}