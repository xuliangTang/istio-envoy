package common
import (
	"strconv"
)

#annotations: [string]: string
annotations: #annotations

ingress_prefix: "envoy.ingress.kubernetes.io"
ingress_annotations:{
	rewrite_target: ingress_prefix + "/" + "rewrite-target"

	// 限流 https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#extensions-filters-http-local-ratelimit-v3-localratelimit
	ratelimit_max: ingress_prefix + "/" + "ratelimit-max"
	ratelimit_perfill: ingress_prefix + "/" + "ratelimit-perfill"
	ratelimit_fillinteval: ingress_prefix + "/" + "ratelimit-fillinteval"

	// 跨域 https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/cors/v3/cors.proto#extension-envoy-filters-http-cors
	cors_enable: ingress_prefix + "/" + "cors-enable"
	cors_allow_origin: ingress_prefix + "/" + "cors-allow-origin"		// 精准匹配
	cors_allow_origin_prefix: ingress_prefix + "/" + "cors-allow-origin-prefix"	// 前缀
	cors_allow_origin_suffix: ingress_prefix + "/" + "cors-allow-origin-suffix"	// 后缀
	cors_allow_origin_regex: ingress_prefix + "/" + "cors-allow-origin-regex"		// 正则
	cors_allow_origin_contains: ingress_prefix + "/" + "cors-allow-origin-contains"	// 包含字符串
	cors_allow_origin_ignore_case: ingress_prefix + "/" + "cors-allow-origin-ignore-case"	// 是否忽略大小写
	cors_allow_methods: ingress_prefix + "/" + "cors-allow-methods"
	cors_allow_headers: ingress_prefix + "/"+"cors-allow-headers"
	cors_expose_headers: ingress_prefix + "/"+"cors-expose-headers"
	cors_max_age: ingress_prefix + "/" + "cors-max-age"
	cors_allow_credentials: ingress_prefix + "/" + "cors-allow-credentials"
}

#ratelimit: {
	max: int | *0  // 桶的数量，如果是 0 则不限流
	perfill: int | *1		// 填充数量
  fillinteval: string | *"1s"	// 填充间隔
}
ratelimit: #ratelimit & {
	if annotations != _|_ {
		 if annotations[ingress_annotations.ratelimit_max] != _|_ {
		 		max: strconv.Atoi(annotations[ingress_annotations.ratelimit_max])	// 因为fill时会被看作字符串，所以要转换为int
		 }
		 if annotations[ingress_annotations.ratelimit_perfill] != _|_{
		 		perfill: annotations[ingress_annotations.ratelimit_perfill]
		 }
		 if annotations[ingress_annotations.ratelimit_fillinteval] != _|_{
		 		fillinteval: annotations[ingress_annotations.ratelimit_fillinteval]
		 }
	}
}

vars: {
	if annotations != _|_ && annotations[ingress_annotations.rewrite_target] != _|_ {
		rewrite_value: annotations[ingress_annotations.rewrite_target]
	}

	if annotations != _|_ && annotations[ingress_annotations.cors_enable] == "true" {
		cors: {
		 	allow_origin_string_match: [
		 		{
					if annotations[ingress_annotations.cors_allow_origin] != _|_ {
						 exact: annotations[ingress_annotations.cors_allow_origin]
					}
					if annotations[ingress_annotations.cors_allow_origin_prefix] != _|_ {
						 prefix: annotations[ingress_annotations.cors_allow_origin_prefix]
					}
					if annotations[ingress_annotations.cors_allow_origin_suffix] != _|_ {
						 suffix: annotations[ingress_annotations.cors_allow_origin_suffix]
					}
					if annotations[ingress_annotations.cors_allow_origin_regex] != _|_ {
						 safe_regex: {
						 		google_re2: {}
						 		regex: annotations[ingress_annotations.cors_allow_origin_regex]
						 }
					}
					if annotations[ingress_annotations.cors_allow_origin_contains] != _|_ {
						 contains: annotations[ingress_annotations.cors_allow_origin_contains]
					}
					if annotations[ingress_annotations.cors_allow_origin_ignore_case] != _|_ {
						 ignore_case: annotations[ingress_annotations.cors_allow_origin_ignore_case]
					}
			  }
		  ]
		 	if annotations[ingress_annotations.cors_allow_methods] != _|_ {
				 allow_methods: annotations[ingress_annotations.cors_allow_methods]
			}
			if annotations[ingress_annotations.cors_allow_headers] != _|_ {
				 allow_headers: annotations[ingress_annotations.cors_allow_headers]
			}
     	if annotations[ingress_annotations.cors_max_age] != _|_ {
				 max_age: annotations[ingress_annotations.cors_max_age]
			}
	  	if annotations[ingress_annotations.cors_expose_headers] != _|_ {
				 expose_headers: annotations[ingress_annotations.cors_expose_headers]
			}
	   	if annotations[ingress_annotations.cors_allow_credentials] != _|_ {
				 allow_credentials: annotations[ingress_annotations.cors_allow_credentials]
			}
		}
	}
}