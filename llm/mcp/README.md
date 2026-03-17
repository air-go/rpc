# MCP (Model Context Protocol) 模块

基于 `github.com/mark3labs/mcp-go` 包实现的本地 MCP 客户端和服务端，支持零网络开销的进程内通信，并提供可扩展的 Hook 机制用于集成 OpenTelemetry 链路追踪和 Prometheus 指标采集。

## 特性

- ✅ **本地通信**: 使用 `NewInProcessClient` 实现零网络开销的进程内通信
- ✅ **Hook 抽象**: 通过 `ServerToolHook` 和 `ClientToolHook` 接口定义统一的 Hook 规范
- ✅ **链路追踪**: 完整的 OpenTelemetry 分布式追踪支持
- ✅ **指标采集**: Prometheus 指标采集，包含调用次数和耗时统计
- ✅ **可扩展性**: HookChain 支持动态注册和组合多个 Hook
- ✅ **线程安全**: HookChain 使用读写锁，支持并发执行

## 结构结构

```
llm/mcp/
├── server/                      # 服务端实现
│   ├── server.go               # 服务端核心实现
│   └── hook/
│       ├── interface.go         # 服务端 Hook 接口定义
│       ├── chain.go             # Hook Chain 实现
│       ├── otel.go              # OpenTelemetry Hook
│       ├── prometheus.go        # Prometheus Hook
│       ├── chain_test.go        # Hook Chain 单元测试
│       ├── otel_test.go         # Otel Hook 单元测试
│       └── prometheus_test.go   # Prometheus Hook 单元测试
├── client/                      # 客户端实现
│   ├── client.go               # 客户端核心实现
│   └── hook/
│       ├── interface.go         # 客户端 Hook 接口定义
│       ├── chain.go             # Hook Chain 实现
│       ├── otel.go              # OpenTelemetry Hook
│       ├── prometheus.go        # Prometheus Hook
│       ├── chain_test.go        # Hook Chain 单元测试
│       ├── otel_test.go         # Otel Hook 单元测试
│       └── prometheus_test.go   # Prometheus Hook 单元测试
└── example/                     # 使用示例
    └── main.go                 # 完整的客户端+服务端示例
```

## 快速开始

### 1. 创建服务端

```go
import (
    "github.com/air-go/rpc/llm/mcp/server"
    serverHook "github.com/air-go/rpc/llm/mcp/server/hook"
    "github.com/mark3labs/mcp-go/mcp"
)

// 创建 HookChain
serverHookChain := serverHook.NewHookChain()

// 注册 OpenTelemetry 追踪 Hook
serverHookChain.RegisterHook(serverHook.NewOtelServerHook())

// 注册 Prometheus 指标 Hook
serverHookChain.RegisterHook(serverHook.NewPrometheusServerHook())

// 创建服务端
mcpServer := server.NewServer("my-server", "1.0.0",
    server.WithHooks(serverHookChain),
)

// 注册工具
tool := mcp.NewTool("echo",
    mcp.WithDescription("Echo back the input message"),
    mcp.WithString("message",
        mcp.Description("Message to echo back"),
        mcp.Required(),
    ),
)

mcpServer.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    message := req.GetString("message", "")
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.TextContent{
                Type: "text",
                Text: fmt.Sprintf("Echo: %s", message),
            },
        },
    }, nil
})
```

### 2. 创建客户端（进程内连接）

```go
import (
    "github.com/air-go/rpc/llm/mcp/client"
    clientHook "github.com/air-go/rpc/llm/mcp/client/hook"
)

// 配置客户端 Hooks
clientHookChain := clientHook.NewClientHookChain()
clientHookChain.RegisterHook(clientHook.NewOtelClientHook())
clientHookChain.RegisterHook(clientHook.NewPrometheusClientHook())

// 创建进程内客户端（零网络开销）
// serverName 是必传参数
mcpClient, err := client.NewInProcessClient(
    mcpServer.GetMCPServer(),
    "my-server",
    client.WithHooks(clientHookChain),
)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
defer mcpClient.Close()
```

### 3. 初始化并调用工具

```go
import "github.com/mark3labs/mcp-go/mcp"

ctx := context.Background()

// 初始化客户端
initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
    Params: mcp.InitializeParams{
        ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
        Capabilities:    mcp.ClientCapabilities{},
        ClientInfo: mcp.Implementation{
            Name:    "my-client",
            Version: "1.0.0",
        },
    },
})
if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}

// 调用工具
result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
    Params: mcp.CallToolParams{
        Name: "echo",
        Arguments: map[string]interface{}{
            "message": "Hello MCP!",
        },
    },
})
if err != nil {
    log.Fatalf("Failed to call tool: %v", err)
}

// 处理结果
for _, content := range result.Content {
    if textContent, ok := content.(mcp.TextContent); ok {
        fmt.Printf("Result: %s\n", textContent.Text)
    }
}
```

## 完整示例

查看 [example/main.go](example/main.go) 获取完整的客户端+服务端使用示例，包含：

- ✅ 服务端创建和工具注册
- ✅ 客户端进程内连接
- ✅ OpenTelemetry 链路追踪集成
- ✅ Prometheus 指标采集集成
- ✅ 多个工具示例（echo、add、greet、calculate）
- ✅ 并发调用演示

### 运行示例

```bash
cd agent/mcp/example
go run main.go
```

## Hook 系统

### Hook 接口

#### 服务端 Hook

```go
type ServerToolHook interface {
    BeforeCall(ctx context.Context, hookCtx *HookContext) error
    AfterCall(ctx context.Context, hookCtx *HookContext)
    OnError(ctx context.Context, hookCtx *HookContext)
}
```

#### 客户端 Hook

```go
type ClientToolHook interface {
    BeforeCall(ctx context.Context, hookCtx *ClientHookContext) error
    AfterCall(ctx context.Context, hookCtx *ClientHookContext)
    OnError(ctx context.Context, hookCtx *ClientHookContext)
}
```

### HookContext

HookContext 提供完整的上下文信息和可扩展的元数据：

```go
type HookContext struct {
    ToolName string                        // 工具名称
    Request  *mcp.CallToolRequest          // 调用请求
    Response *mcp.CallToolResult          // 调用响应
    Error    error                        // 错误信息
    Metadata map[string]interface{}       // 可扩展的元数据
}
```

### HookChain

HookChain 管理多个 Hook 的执行顺序，支持并发安全：

```go
chain := hook.NewHookChain()

// 注册多个 Hook
chain.RegisterHook(hook1)
chain.RegisterHook(hook2)
chain.RegisterHook(hook3)

// 执行顺序：BeforeCall1 -> BeforeCall2 -> BeforeCall3 -> handler -> AfterCall1 -> AfterCall2 -> AfterCall3
```

## 可观测性集成

### OpenTelemetry 链路追踪

自动为每个工具调用创建 Span，记录：

- 工具名称
- 参数信息（args count）
- 响应信息（content length）
- 错误信息

Prometheus 指标自动记录：

#### 服务端指标

| 指标名称 | 类型 | 标签 | 说明 |
|---------|------|------|------|
| `mcp_server_tool_calls_total` | Counter | tool_name, status | 工具调用总数 |
| `mcp_server_tool_duration_seconds` | Histogram | tool_name | 工具执行耗时 |

#### 客户端指标

| 指标名称 | 类型 | 标签 | 说明 |
|---------|------|------|------|
| `mcp_client_tool_calls_total` | Counter | server_name, tool_name, status | 客户端工具调用总数 |
| `mcp_client_tool_duration_seconds` | Histogram | server_name, tool_name | 客户端工具调用耗时 |

## 测试

运行所有单元测试：

```bash
# 测试服务端
go test ./agent/mcp/server/... -v

# 测试客户端
go test ./agent/mcp/client/... -v

# 测试所有
go test ./agent/mcp/... -v
```

测试覆盖率：88个测试用例，全部通过

## 依赖

- `github.com/mark3labs/mcp-go@v0.45.0` - MCP 协议实现
- `go.opentelemetry.io/otel` - OpenTelemetry 链路追踪
- `github.com/prometheus/client_golang` - Prometheus 指标采集

## 注意事项

### Prometheus Hook 重复注册问题

由于 Prometheus 使用 `promauto` 自动注册到默认 registry，在单元测试中创建多个 Hook 实例会导致重复注册错误。在实际使用中，建议：

1. **生产环境**: 每个服务端/客户端只创建一个 Prometheus Hook 实例
2. **测试环境**: 使用单例模式或自定义 registry 避免重复注册

### Hook 执行顺序

1. **BeforeCall**: 按注册顺序依次执行，任一返回错误则中断
2. **Handler**: 执行实际工具处理逻辑
3. **AfterCall/OnError**: 按注册顺序依次执行

### 本地通信 vs 网络通信

本实现专注于进程内通信（InProcess），提供零网络开销的优势。如需网络通信（stdio/http），可以扩展 `Client` 支持外部 transport。

## 扩展性

### 添加自定义 Hook

```go
// 实现 Hook 接口
type CustomHook struct {
    customField string
}

func (h *CustomHook) BeforeCall(ctx context.Context, hookCtx *HookContext) error {
    hookCtx.Metadata["custom_data"] = h.customField
    return nil
}

func (h *CustomHook) AfterCall(ctx context.Context, hookCtx *HookContext) {
    // 处理调用后逻辑
}

func (h *CustomHook) OnError(ctx context.Context, hookCtx *HookContext) {
    // 处理错误
}

// 注册到 HookChain
chain.RegisterHook(&CustomHook{customField: "value"})
```

### 常见的自定义 Hook 实现

- 日志记录 Hook
- 限流 Hook
- 缓存 Hook
- 认证授权 Hook
- 监控告警 Hook

## 最佳实践

1. **Hook 顺序**: 建议先注册 Otel Hook（用于追踪），再注册 Prometheus Hook（用于指标）
2. **元数据传递**: 使用 `HookContext.Metadata` 在 Hook 生命周期中传递数据
3. **错误处理**: 在 `OnError` Hook 中记录错误，但不要吞掉错误
4. **性能考虑**: Hook 应该轻量快速，避免阻塞主流程
5. **测试覆盖**: 自定义 Hook 应该有对应的单元测试

## 许可证

本模块遵循项目的开源许可证。
