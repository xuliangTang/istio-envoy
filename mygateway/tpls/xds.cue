#input: {
	 listen_port: int | *8080
}
input: #input

output: {
	listener: {
		name: "0.0.0.0_8080"
		address: {
			 socket_address: {
			 	 address: "0.0.0.0",
		     port_value: input.listen_port
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
															cluster_name: "gateway_cluster"
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
				{
					name: "myhost"
					domains: [ "envoy.virtuallain.com:32180" ]
					routes: [
						{
							 match: {
									prefix: "/"
							 }
							 route: {
								 cluster: "prod_cluster"
							 }
						}
					]
				}
		]
	}

	clusters: [
      {
      	  name: "prod_cluster"
					connect_timeout: "1s"
					type: "LOGICAL_DNS"
					lb_policy: "ROUND_ROBIN"
					load_assignment: {
						cluster_name: "prod_cluster"
						endpoints: [
							{
								lb_endpoints: [
									{
										endpoint: {
											address: {
													socket_address: {
													  	address: "prodsvc.default.svc.cluster.local"
															port_value: 80
													}
											}
									  }
									}]
							}]
					}
      }
  ]

}