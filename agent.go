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
	for _, mcpCli := range mcpClient {
		// try to start mcp client
		err := mcpCli.Start()
		if err != nil {
			fmt.Println("Error starting Mcp Client:", err)
			continue
		}
		// try to set tools
		err = mcpCli.SetTools()
		if err != nil {
			fmt.Println("Error setting tools: ", err)
			continue
		}
		// put tool in tools list
		tools = append(tools, mcpCli.Gettools()...)
	}
	// 2. Initialize llm with tools
	llm := llm.CreateNewClient(ctx, model, llm.WithTools(tools))
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
	response, call_tools := a.LLM.Chat(prompt)
	for len(call_tools) > 0 {
		for _, calltool := range call_tools {
			for _, mcpclient := range a.McpClient {
				tools := mcpclient.Gettools()
				for _, tool := range tools {
					if tool.Name == calltool.Function.Name {
						fmt.Println("tool use:", tool)
						tool_res, err := mcpclient.Calltool(calltool.Function.Name, calltool.Function.Arguments)
						if err != nil {
							fmt.Println("Error tool calling", err)
							continue
						}
						a.LLM.Message = append(a.LLM.Message, openai.ToolMessage(tool_res, calltool.ID))
					}
				}
			}
		}
	}
	response, call_tools = a.LLM.Chat("")
	a.Close()
	return response
}

// close connection
func (a *CrawlAgent) Close() {
	for _, mcpCli := range a.McpClient {
		_ = mcpCli.Client.Close()
	}
}
