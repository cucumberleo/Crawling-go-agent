package main

import (
	"context"

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
type MCP_Client struct{
	Ctx context.Context
	Client *client.Client
	Cmd string
	Tools []mcp.Tool
	Args []string
	Env []string
}


func NewMCPClient(ctx context.Context, cmd string, args,env []string) *MCP_Client{
	// 建立标准输入输出传输协议
	stdioProto := transport.NewStdio(cmd,env,args...)
	client := client.NewClient(stdioProto)
	// 这边传参
	return &MCP_Client{
		Ctx: ctx,
		Client: client,
		Cmd: cmd,
		Env: env,
		Args: args,
	}
}

// start mcp client
func (m *MCP_Client)Start() error{
	if err := m.Client.Start(m.Ctx); err!=nil{
		return err
	}
	mcpInitReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name: "mcp-client-go",
				Version: "0.1.0",
			},
		},
	}
	if _, err := m.Client.Initialize(m.Ctx,mcpInitReq); err!=nil{
		return err
	}
	return nil
}

// 设置工具
func (m *MCP_Client) SetTools() error{
	toolsReq := mcp.ListToolsRequest{}
	tools, err := m.Client.ListTools(m.Ctx,toolsReq)
	if err!=nil{
		return err
	}
	m.Tools = tools.Tools
	return nil
}

// 调用工具
func (m *MCP_Client) Calltool(name string, args any) (string, error){
	res, err := m.Client.CallTool(m.Ctx,mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: name,
			Arguments: args,
		},
	})
	if err != nil{
		return "",err
	}
	return mcp.GetTextFromContent(res.Content), nil
}

// 获取工具
func (m *MCP_Client) Gettools() []mcp.Tool{
	return m.Tools
}

