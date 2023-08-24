package common
import (
	"strconv"
)

#annotations: [string]: string
annotations: #annotations

ingress_prefix: "envoy.ingress.kubernetes.io"
ingress_annotations:{
	rewrite_target: ingress_prefix + "/" + "rewrite-target"

	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#extensions-filters-http-local-ratelimit-v3-localratelimit
	ratelimit_max: ingress_prefix + "/" + "ratelimit-max"
	ratelimit_perfill: ingress_prefix + "/" + "ratelimit-perfill"
	ratelimit_fillinteval: ingress_prefix + "/" + "ratelimit-fillinteval"
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
}