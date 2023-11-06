<!--
 * @Descripttion:
 * @Author: weihaoyu
-->
<div align="center">

<h1>rpc</h1>

[![workflows](https://github.com/air-go/rpc/workflows/Go/badge.svg)](https://github.com/air-go/rpc/actions?query=workflow%3AGo+branch%3Amaster)
[![Release](https://img.shields.io/github/v/release/air-go/rpc.svg?style=flat-square)](https://github.com/air-go/rpc/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/air-go/rpc)](https://goreportcard.com/report/github.com/air-go/rpc)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

<p> Go 微服务框架，同时支持 gRPC 和 HTTP，封装各种常用组件，开箱即用，专注业务 </p>

<img src="https://camo.githubusercontent.com/82291b0fe831bfc6781e07fc5090cbd0a8b912bb8b8d4fec0696c881834f81ac/68747470733a2f2f70726f626f742e6d656469612f394575424971676170492e676966" width="800"  height="3">

</div>
<br>

## 建议反馈
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

## 功能列表
✅ &nbsp;多格式配置读取
<br>
✅ &nbsp;服务优雅关闭
<br>
✅ &nbsp;进程结束资源自动回收
<br>
✅ &nbsp;日志抽象和标准字段统一（请求、DB、Redis、RPC）
<br>
✅ &nbsp;DB（ORM）
<br>
✅ &nbsp;Redis
<br>
✅ &nbsp;RabbitMQ
<br>
✅ &nbsp;Kafka
<br>
✅ &nbsp;Apollo配置中心
<br>
✅ &nbsp;cron定时任务
<br>
✅ &nbsp;单机和分布式限流
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
✅ &nbsp;通用链接池
<br>
✅ &nbsp;HTTP-RPC 超时传递
<br>
✅ &nbsp;端口多路复用
<br>
✅ &nbsp;gRPC
<br>
✅ &nbsp;Prometheus 监控
<br><br>

## 后续规划
日志收集
<br>
告警
<br>
限流
<br>
熔断
<br><br>

## 工程目录
```
- rpc
  - bootstrap //应用启动
  - client
    - grpc //grpc客户端
    - http //http客户端
  - library //基础组件库，不建议修改
    - app //app
    - apollo //阿波罗
    - cache //分布式缓存
    - config //配置加载
    - cron //任务调度
    - etcd //etcd
    - grpc //grpc封装
    - opentracing //opentracing分布式链路追踪
    - limiter //限流
    - lock //分布式锁
    - logger //日志
    - orm //db orm
    - otel //otel分布式链路追踪
    - pool //通用链接池
    - prometheus //prometheus监控
    - queue //消息队列
    - redis //redis
    - registry //注册中心
    - reliablequeue //可靠消息队列
    - selector //负载均衡器
    - servicer //下游服务
  - mock
    - third //三方单测mock
    - tools //常见mock工具封装
  - server
    - grpc //grpc服务端
    - http //http服务端  
  - third //三方依赖引入
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

---
[![Star History Chart](https://api.star-history.com/svg?repos=air-go/rpc&type=Date)](https://star-history.com/#air-go/rpc&Date)
