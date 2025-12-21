package main

import (
	"context"
	"fmt"

	llm "ds-agent/LLM"
	mcpClient "ds-agent/MCP"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
)

/*
McpClient: Mcp客户
LLM: Gpt接口
其他同openai_ai.go中解释
*/

type CrawlAgent struct {
	Ctx       context.Context
	McpClient []*mcpClient.MCP_Client
	LLM       *llm.OpenAI_ai
	Model     string
	RagCtx    string
}

// 构建agent
func NewAgent(ctx context.Context, model string, mcpClient []*mcpClient.MCP_Client) *CrawlAgent {
	// 1. initialize tools
	tools := make([]mcp.Tool, 0)
	for i, mcpCli := range mcpClient {
		fmt.Printf("正在启动 MCP 客户端 %d (%s)...\n", i+1, mcpCli.Cmd)
		// try to start mcp client
		err := mcpCli.Start()
		if err != nil {
			fmt.Printf("错误：启动 MCP 客户端失败: %v\n", err)
			continue
		}
		fmt.Printf("MCP 客户端 %d 启动成功\n", i+1)

		// try to set tools
		fmt.Printf("正在获取 MCP 客户端 %d 的工具列表...\n", i+1)
		err = mcpCli.SetTools()
		if err != nil {
			fmt.Printf("错误：获取工具列表失败: %v\n", err)
			continue
		}
		toolCount := len(mcpCli.Gettools())
		fmt.Printf("MCP 客户端 %d 共有 %d 个工具\n", i+1, toolCount)
		// put tool in tools list
		tools = append(tools, mcpCli.Gettools()...)
	}
	fmt.Printf("总共获取到 %d 个工具\n", len(tools))
	// 2. Initialize llm with tools
	fmt.Println("正在初始化 LLM 客户端...")
	llm := llm.CreateNewClient(ctx, model, llm.WithTools(tools))
	fmt.Println("LLM 客户端初始化完成")
	return &CrawlAgent{
		McpClient: mcpClient,
		LLM:       llm,
		Model:     model,
		Ctx:       ctx,
	}
}

// Start toolcalling
func (a *CrawlAgent) StartAction(prompt string) string {
	if a.LLM == nil {
		fmt.Println("LLM is not initialized")
		return ""
	}
	// find which mcpclient the tool belong to
	// O(n3) remain optimize
	fmt.Println("正在发送初始提示词到 LLM...")
	response, call_tools := a.LLM.Chat(prompt)
	fmt.Printf("收到 LLM 响应，工具调用数量: %d\n", len(call_tools))

	iteration := 0
	for len(call_tools) > 0 {
		iteration++
		fmt.Printf("\n=== 工具调用循环第 %d 轮 ===\n", iteration)
		for i, calltool := range call_tools {
			fmt.Printf("处理工具调用 %d/%d: %s\n", i+1, len(call_tools), calltool.Function.Name)
			found := false
			for _, mcpclient := range a.McpClient {
				tools := mcpclient.Gettools()
				for _, tool := range tools {
					if tool.Name == calltool.Function.Name {
						found = true
						fmt.Printf("找到对应的 MCP 客户端，正在调用工具: %s\n", tool.Name)
						// 打印工具参数（Arguments 是 string 类型）
						fmt.Printf("工具参数: %s\n", calltool.Function.Arguments)
						tool_res, err := mcpclient.Calltool(calltool.Function.Name, calltool.Function.Arguments)
						if err != nil {
							fmt.Printf("错误：工具调用失败: %v\n", err)
							continue
						}
						fmt.Printf("工具调用成功，结果长度: %d 字符\n", len(tool_res))
						// 打印工具返回结果的前200个字符，用于调试
						previewLen := len(tool_res)
						if previewLen > 200 {
							previewLen = 200
						}
						fmt.Printf("工具返回结果预览（前%d字符）: %s\n", previewLen, tool_res[:previewLen])
						a.LLM.Message = append(a.LLM.Message, openai.ToolMessage(tool_res, calltool.ID))
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				fmt.Printf("警告：未找到工具 %s 对应的 MCP 客户端\n", calltool.Function.Name)
			}
		}
		fmt.Println("正在发送工具结果到 LLM...")
		// Fix:防止无限循环
		response, call_tools = a.LLM.Chat("")
		fmt.Printf("收到 LLM 响应，工具调用数量: %d\n", len(call_tools))
	}
	fmt.Println("\n所有工具调用完成，正在关闭连接...")
	a.Close()
	return response
}

// close connection
func (a *CrawlAgent) Close() {
	for _, mcpCli := range a.McpClient {
		_ = mcpCli.Client.Close()
	}
}
