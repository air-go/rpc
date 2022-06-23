<!--
 * @Descripttion:
 * @Author: weihaoyu
-->

# rpc
Go 微服务框架，同时支持 gRPC 和 HTTP，封装各种常用组件，开箱即用，专注业务。
<br><br>
如果您对本框架有任何意见或建议，欢迎随时通过以下方式反馈和完善：
1. 提 issues 反馈
2. 通过下方的联系方式直接联系我
3. 提 PR 共同维护
<br><br>

## 联系我
QQ群：909211071
<br>
个人QQ：444216978
<br>
微信：AirGo___
<br><br>

## 目前已支持
✅ &nbsp;多格式配置读取
<br>
✅ &nbsp;服务优雅关闭
<br>
✅ &nbsp;进程结束资源自动回收
<br>
✅ &nbsp;日志抽象和标准字段统一（请求、DB、Redis、RPC）
<br>
✅ &nbsp;DB
<br>
✅ &nbsp;RabbitMQ
<br>
✅ &nbsp;Redis
<br>
✅ &nbsp;分布式缓存（解决缓存穿透、击穿、雪崩）
<br>
✅ &nbsp;分布式链路追踪
<br>
✅ &nbsp;分布式锁
<br>
✅ &nbsp;服务注册
<br>
✅ &nbsp;服务发现
<br>
✅ &nbsp;负载均衡
<br>
✅ &nbsp;HTTP-RPC 超时传递
<br>
✅ &nbsp;端口多路复用
<br>
✅ &nbsp;gRPC
<br><br>

## 后续逐渐支持
日志收集
<br>
监控告警
<br>
限流
<br>
熔断
<br><br>

# 目录结构
```
- rpc
  - bootstrap //应用启动
  - client
    - grpc //grpc客户端
    - http //http客户端
  - server
    - grpc //grpc服务端
    - http //http服务端
  - library //基础组件库，不建议修改
    - app //app
    - apollo //阿波罗
    - cache //分布式缓存
    - config //配置加载
    - cron //任务调度
    - etcd //etcd
    - grpc //grpc封装
    - jaeger //jaeger分布式链路追踪
    - job //离线任务
    - lock //分布式锁
    - logger //日志
    - orm //db orm
    - queue //消息队列
    - redis //redis
    - registry //注册中心
    - reliablequeue //可靠消息队列
    - selector //负载均衡器
    - servicer //下游服务
  .gitignore
  Dockerfile
  LICENSE
  Makefile
  README.md
  go.mod
  go.sum
```
<br>

## Example
<a href="https://github.com/air-go/rpc-example/blob/master/http/main.go">HTTP</a>
<br>
<a href="https://github.com/air-go/rpc-example/blob/master/grpc/main.go">gRPC</a>
<br>
<a href="https://github.com/air-go/rpc-example/blob/master/trace">完整业务架构</a>
<br>
