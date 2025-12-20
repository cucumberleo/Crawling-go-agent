package llm

import (
	"context"
	"fmt"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestGPT(t *testing.T) {
	ctx := context.Background()
	model_name := openai.ChatModelGPT3_5Turbo
	llm := CreateNewClient(ctx, model_name)
	prompt := "HELLO"
	result, call_tool := llm.Chat(prompt)
	if len(call_tool) != 0 {
		fmt.Println("Tool call:", call_tool)
	}
	fmt.Println("Result: ", result)
}
