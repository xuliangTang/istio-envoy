package ratelimit

// 一般用于http_connection_mamanger
local_ratelimit: {
	name: "envoy.filters.http.local_ratelimit"
	typed_config: {
		"@type": "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit"
		 stat_prefix: "http_local_rate_limiter"
	}
}

local_ratelimit_prefilter_config: {
	max_tokens: int | string
	tokens_per_fill: int | string
	fill_interval: string
	prefix: string  // 用于构建下面一坨内容的前缀
}

local_ratelimit_prefilter: {
	"envoy.filters.http.local_ratelimit": {
		 "@type": "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit"
			stat_prefix: local_ratelimit_prefilter_config.prefix + "_local_rate_limiter"
			token_bucket: {
				max_tokens: local_ratelimit_prefilter_config.max_tokens
				tokens_per_fill: local_ratelimit_prefilter_config.tokens_per_fill
				fill_interval: local_ratelimit_prefilter_config.fill_interval
			}
			filter_enabled: {
				runtime_key: local_ratelimit_prefilter_config.prefix + "_rate_limit_enabled"
				default_value: {
					 numerator: 100
					 denominator: "HUNDRED"
				}
			}
			filter_enforced: {
				runtime_key: local_ratelimit_prefilter_config.prefix + "_rate_limit_enforced"
				default_value: {
					 numerator: 100
					 denominator: "HUNDRED"
				}
			}
			response_headers_to_add: [
				{
					append: false
					header: {
						key: "x-local-rate-limit"
						value: "true"
					}
				}
			]
	}
}