input: {}

nginx_prefix: "nginx.ingress.kubernetes.io"
nginx_annotations:{
	rewrite_target: nginx_prefix + "/" + "rewrite-target"
}

annotations: input.metadata.annotations

vars: {
	if annotations == _|_ {
		rewrite: false
		rewrite_value: ""
	}

	if annotations != _|_ && annotations[nginx_annotations.rewrite_target] != _|_ {
		rewrite: true
		rewrite_value: annotations[nginx_annotations.rewrite_target]
	}
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
						routes: [
							for _, p in rule.bytes.http.paths {
								{
									 match: {
									 		if p.pathType == "Prefix" {
									 			if vars.rewrite == false {
									 				prefix: p.path
									 			}
									 			if vars.rewrite == true {	// 判断路径重写
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
										  if vars.rewrite == true{
												 	regex_rewrite: {
												 		pattern: {
												 			google_re2: {}
                              regex: p.path
														}
                            substitution: vars.rewrite_value
												 	}
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