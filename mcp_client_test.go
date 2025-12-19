package main

import (
	"context"
	"fmt"
	"testing"
)

func Test_mcp_client(t *testing.T){
	ctx := context.Background()
	client := NewMCPClient(ctx, "uvx", []string{"mcp-server-fetch"}, nil)
	err := client.Start()
	if err != nil{
		fmt.Printf("Fail to start MCP client: %v",err)
		return
	}
	err = client.SetTools()
	if err != nil{
		fmt.Printf("Fail to start MCP client: %v",err)
		return
	}
	tools := client.Gettools()
	if len(tools)==0{
		fmt.Printf("tools is empty")
		return
	}
	fmt.Printf("MCP client started successfully with tools: %v",tools)
}