## Ingress Envoy Controller
基于 envoy 的网关控制面，由 operator 监听 ingressClassName 为 myenvoy 的 k8s ingress 资源，发生变更时自动同步 envoy 网关配置

### Install
```
kubectl apply -f https://raw.githubusercontent.com/xuliangTang/istio-envoy/main/mygateway/deploy/envoy.yaml
```

### Annotations
- `envoy.ingress.kubernetes.io/rewrite-target`: 路径重写
- `envoy.ingress.kubernetes.io/ratelimit-max`: 限流rps
- `envoy.ingress.kubernetes.io/ratelimit-perfill`: 限流令牌填充数量
- `envoy.ingress.kubernetes.io/ratelimit-fillinteval`: 限流令牌填充间隔
- `envoy.ingress.kubernetes.io/cors-enable`: 跨域开关
- `envoy.ingress.kubernetes.io/cors-allow-origin`: 允许的来源
- `envoy.ingress.kubernetes.io/cors-allow-origin-prefix`: 允许前缀来源
- `envoy.ingress.kubernetes.io/cors-allow-origin-suffix`: 允许后缀来源
- `envoy.ingress.kubernetes.io/cors-allow-origin-regex`: 允许正则的来源
- `envoy.ingress.kubernetes.io/cors-allow-origin-contains`: 允许包含该字符串的来源
- `envoy.ingress.kubernetes.io/cors-allow-origin-ignore-case`: 允许来源是否忽略大小写
- `envoy.ingress.kubernetes.io/cors-allow-methods`: 允许的请求方式
- `envoy.ingress.kubernetes.io/cors-allow-headers`: 允许的请求头
- `envoy.ingress.kubernetes.io/cors-expose-headers`: 允许暴露给客户端的HTTP头
- `envoy.ingress.kubernetes.io/cors-max-age`: 浏览器可以缓存该相应多少秒
- `envoy.ingress.kubernetes.io/cors-allow-credentials`: 是否允许跨域请求携带凭证(如 Cookie)
- `envoy.ingress.kubernetes.io/lua-block`: lua脚本的完整代码块
- `envoy.ingress.kubernetes.io/lua-request`: lua脚本的envoy_on_request函数体
- `envoy.ingress.kubernetes.io/lua-response`: lua脚本的envoy_on_response函数体