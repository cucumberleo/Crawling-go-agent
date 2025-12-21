package main

import (
	"context"
	mcpClient "ds-agent/MCP"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
)

func main() {
	// 添加超时控制（30分钟）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Println("开始初始化 MCP 客户端...")
	allowDir, _ := os.Getwd()
	fmt.Printf("当前目录: %s\n", allowDir)

	fmt.Println("正在创建 fetch MCP 客户端...")
	fetchMcpClient := mcpClient.NewMCPClient(ctx, "uvx", []string{"mcp-server-fetch"}, nil)

	fmt.Println("正在创建 filesystem MCP 客户端...")
	fileMcpClient := mcpClient.NewMCPClient(ctx, "npx", []string{"-y", "@modelcontextprotocol/server-filesystem", allowDir}, nil)

	fmt.Println("正在初始化 Agent...")
	agent := NewAgent(ctx, openai.ChatModelGPT3_5Turbo, []*mcpClient.MCP_Client{fetchMcpClient, fileMcpClient})

	fmt.Println("Agent 初始化完成，开始执行任务...")
	prompt := `你是一个内容获取与文件写入助手。
	你必须且只能使用我提供的 MCP 工具完成任务:
	1) 使用 mcp-server-fetch 获取公开网页的内容(仅访问单页遵守网站条款，不进行批量或大规模抓取))
	2)对获取的公开内容进行简要总结。
	3) 使用 @modelcontextprotocol/server-filesystem 将结果写入当前目录的文件。
	不要自行访问网络或返回拒绝，直接调用上述工具完成。
	访问 https://example.com/ 首页公开内容，提取简要摘要，并将结果写入当前目录的new.md(若存在则覆盖))。只使用提供的工具完成。`

	result := agent.StartAction(prompt)
	fmt.Println("Final result:", result)
}
