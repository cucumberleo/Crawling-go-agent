package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// define LLM structure
/*
	Modelname:模型名称
	SystemPrompt:系统提示词
	RagCtx: RAG向量信息
	Tools：可以调用的工具信息
	Messgage: 历史对话信息
	LLM：调用大模型api的实例
*/
type OpenAI_ai struct {
	Ctx          context.Context
	Modelname    string
	SystemPrompt string
	RagCtx       string
	Tools        []mcp.Tool
	Message      []openai.ChatCompletionMessageParamUnion
	LLM          openai.Client
}

type LLM_Option func(*OpenAI_ai)

// init model internal systemprompt
func WithSystemPrompt(prompt string) LLM_Option {
	return func(llmp *OpenAI_ai) {
		llmp.SystemPrompt = prompt
	}
}

// init modelname
func WithModelname(model_name string) LLM_Option {
	return func(llmp *OpenAI_ai) {
		llmp.Modelname = model_name
	}
}

// init Rag context
func WithRagctx(ragctx string) LLM_Option {
	return func(llmp *OpenAI_ai) {
		llmp.RagCtx = ragctx
	}
}

// init Tools
func WithTools(tool []mcp.Tool) LLM_Option {
	return func(llmp *OpenAI_ai) {
		llmp.Tools = tool
	}
}

// init LLM client
func CreateNewClient(ctx context.Context, model_name string, opts ...LLM_Option) *OpenAI_ai {
	if model_name == "" {
		panic("Error: Model name is empty.")
	}
	// get api key
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_API_BASE_URL")
	if apiKey == "" {
		panic("Error:API key is empty")
	}

	// 构建gpt client 需要的option
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}
	// 创建LLM client
	cli := openai.NewClient(options...)
	// 根据参数构造llm结构体
	llm := &OpenAI_ai{
		Ctx:       ctx,
		Modelname: model_name,
		LLM:       cli,
	}
	for _, opt := range opts {
		opt(llm)
	}
	// 将本次系统提示词封装成系统message，加入整个对话历史
	if llm.SystemPrompt != "" {
		llm.Message = append(llm.Message, openai.SystemMessage(llm.SystemPrompt))
	}
	// rag信息同理
	if llm.RagCtx != "" {
		llm.Message = append(llm.Message, openai.SystemMessage(llm.RagCtx))
	}
	fmt.Println("Successfully init llm")
	return llm
}

// chat方法
// 第一个括号表示绑定在OpenAI_ai这个结构体指针上的函数
func (c *OpenAI_ai) Chat(prompt string) (result string, toolCall []openai.ToolCallUnion) {
	if prompt != "" {
		c.Message = append(c.Message, openai.UserMessage(prompt))
	}
	fmt.Printf("[LLM] 正在准备 API 请求，工具数量: %d\n", len(c.Tools))
	gpt_tools := MCPtool2OpenAItool(c.Tools)
	fmt.Println("[LLM] 正在发送请求到 OpenAI API...")
	stream := c.LLM.Chat.Completions.NewStreaming(c.Ctx, openai.ChatCompletionNewParams{
		Model:    c.Modelname,
		Messages: c.Message,
		Seed:     openai.Int(0),
		// 这边要做一个转换，mcp的tool不是openai的tool
		Tools: gpt_tools,
	})
	// 增加流错误检查
	if err := stream.Err(); err != nil {
		panic(fmt.Sprintf("Stream initialization error: %v", err))
	}
	fmt.Println("[LLM] 开始接收流式响应...")
	result = ""
	// finished := false
	var fullContent strings.Builder
	// 存放chunk的容器
	acc := openai.ChatCompletionAccumulator{}
	chunkCount := 0
	for stream.Next() {
		chunkCount++
		chunk := stream.Current()
		// fmt.Printf("Received chunk: [%s]\n", chunk.Choices[0].Delta.Content)
		// 流变成块存放进去
		acc.AddChunk(chunk)
		// 如果内容已经结束,结果就是当前的内容
		if tool, ok := acc.JustFinishedToolCall(); ok {
			fmt.Printf("[LLM] 检测到工具调用: %s\n", tool.Name)
			toolCall = append(toolCall, openai.ToolCallUnion{
				ID: tool.ID,
				Function: openai.FunctionToolCallFunction{
					Name:      tool.Name,
					Arguments: tool.Arguments,
				},
			})
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)
		}
		// if stream.Err() != nil{
		// 	panic(stream.Err())
		// }
	}
	if err := stream.Err(); err != nil {
		fmt.Printf("[LLM] 流式响应错误: %v\n", err)
	}
	fmt.Printf("[LLM] 接收完成，共收到 %d 个数据块，内容长度: %d 字符\n", chunkCount, fullContent.Len())
	// 将本次assistant消息补回历史，确保下一轮ToolMessage正确关联
	if len(acc.Choices) > 0 {
		c.Message = append(c.Message, acc.Choices[0].Message.ToParam())
	}
	result = fullContent.String()
	return result, toolCall
}

// 将mcp tools转化成openai tools
func MCPtool2OpenAItool(mcp_tool []mcp.Tool) []openai.ChatCompletionToolUnionParam {
	Openai_tool := make([]openai.ChatCompletionToolUnionParam, 0, len(mcp_tool))
	for _, tool := range mcp_tool {
		// 确保 required 字段总是一个数组，不能是 nil
		required := tool.InputSchema.Required
		if required == nil {
			required = []string{}
		}
		
		parms := openai.FunctionParameters{
			"type":       tool.InputSchema.Type,
			"properties": tool.InputSchema.Properties,
			"required":   required,
		}
		Openai_tool = append(Openai_tool, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  parms,
				},
			},
		})
	}
	return Openai_tool
}
