package mcpClient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

/*
Client: MCP 客户
Cmd：终端输入的命令
Env：环境
Args：所需的参数
Tools：工具

*/
// mcp client结构体
type MCP_Client struct {
	Ctx    context.Context
	Client *client.Client
	Cmd    string
	Tools  []mcp.Tool
	Args   []string
	Env    []string
}

func NewMCPClient(ctx context.Context, cmd string, args, env []string) *MCP_Client {
	// 建立标准输入输出传输协议
	stdioProto := transport.NewStdio(cmd, env, args...)
	client := client.NewClient(stdioProto)
	// 这边传参
	return &MCP_Client{
		Ctx:    ctx,
		Client: client,
		Cmd:    cmd,
		Env:    env,
		Args:   args,
	}
}

// start mcp client
func (m *MCP_Client) Start() error {
	fmt.Printf("[MCP] 正在启动客户端进程: %s %v\n", m.Cmd, m.Args)
	startTime := time.Now()
	if err := m.Client.Start(m.Ctx); err != nil {
		fmt.Printf("[MCP] 启动客户端进程失败: %v\n", err)
		return err
	}
	fmt.Printf("[MCP] 客户端进程启动成功 (耗时: %v)，正在初始化协议...\n", time.Since(startTime))
	
	// 检查上下文是否已取消
	if m.Ctx.Err() != nil {
		fmt.Printf("[MCP] 上下文已取消: %v\n", m.Ctx.Err())
		return m.Ctx.Err()
	}
	
	initStartTime := time.Now()
	mcpInitReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcp-client-go",
				Version: "0.1.0",
			},
		},
	}
	fmt.Printf("[MCP] 发送初始化请求...\n")
	if _, err := m.Client.Initialize(m.Ctx, mcpInitReq); err != nil {
		fmt.Printf("[MCP] 协议初始化失败 (耗时: %v): %v\n", time.Since(initStartTime), err)
		return err
	}
	fmt.Printf("[MCP] 协议初始化成功 (耗时: %v)\n", time.Since(initStartTime))
	return nil
}

// 设置工具
func (m *MCP_Client) SetTools() error {
	fmt.Printf("[MCP] 正在获取工具列表...\n")
	toolsReq := mcp.ListToolsRequest{}
	tools, err := m.Client.ListTools(m.Ctx, toolsReq)
	if err != nil {
		fmt.Printf("[MCP] 获取工具列表失败: %v\n", err)
		return err
	}
	m.Tools = tools.Tools
	fmt.Printf("[MCP] 成功获取 %d 个工具\n", len(m.Tools))
	return nil
}

// 调用工具
func (m *MCP_Client) Calltool(name string, args any) (string, error) {
	fmt.Printf("[MCP] 正在调用工具: %s\n", name)
	var arguments map[string]any
	switch v := args.(type) {
	case map[string]any:
		arguments = v
	case string:
		err := json.Unmarshal([]byte(v), &arguments)
		if err != nil {
			fmt.Printf("[MCP] 解析工具参数失败: %v\n", err)
			return "", err
		}
	default:
	}
	res, err := m.Client.CallTool(m.Ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: arguments,
		},
	})
	if err != nil {
		fmt.Printf("[MCP] 工具调用失败: %v\n", err)
		return "", err
	}
	result := mcp.GetTextFromContent(res.Content)
	fmt.Printf("[MCP] 工具调用成功，结果长度: %d 字符\n", len(result))
	return result, nil
}

// 获取工具
func (m *MCP_Client) Gettools() []mcp.Tool {
	return m.Tools
}
