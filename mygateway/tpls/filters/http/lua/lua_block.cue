package lua

lua: {
	block: string | *""		// 完整的lua script
	request: string | *""		// envoy_on_request函数体
	response: string | *""	// envoy_on_response函数体
}

lua_filter_config: {
		name: "envoy.filters.http.lua"
		typed_config: {
			"@type": "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
			if lua.block != "" {
				inline_code: lua.block
			}
			if lua.block == "" {
				inline_code: """
function envoy_on_response(response)
\(lua.response)
end
function envoy_on_request(request)
\(lua.request)
end
"""
			}
		}
}