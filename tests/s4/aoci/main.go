package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 本工具只驱动Agent自有公开仓库的实际MCP，不生成语义或修改候选。
func run() error {
	binary := flag.String("binary", "", "trusted AOCI executable")
	root := flag.String("root", "", "owned public worktree")
	tool := flag.String("tool", "", "official tool name or list")
	request := flag.String("request", "", "ignored local JSON arguments file")
	output := flag.String("output", "", "ignored local MCP response file")
	flag.Parse()
	if *binary == "" || *root == "" || *tool == "" || *output == "" {
		return errors.New("explicit tool paths and output are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	remote, err := exec.CommandContext(ctx, "git", "-C", *root, "remote", "get-url", "origin").Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(remote)) != "https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015.git" {
		return errors.New("only the owned public project is authorized")
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "astrocyte-s4-owned-public-authoring", Version: "1"}, nil)
	command := exec.CommandContext(ctx, *binary, "--repo", *root, "mcp")
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		return err
	}
	defer session.Close()
	var response any
	var toolFailed bool
	if *tool == "list" {
		result, err := session.ListTools(ctx, nil)
		if err != nil {
			return err
		}
		response = result
		fmt.Println("official_tools", len(result.Tools))
	} else {
		args := map[string]any{}
		if *request != "" {
			data, err := os.ReadFile(*request)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(data, &args); err != nil {
				return err
			}
		}
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: *tool, Arguments: args})
		if err != nil {
			return err
		}
		response = result
		toolFailed = result.IsError
		fmt.Println("official_tool", *tool, "is_error", result.IsError, "content_blocks", len(result.Content))
	}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if toolFailed {
		return errors.New("official MCP tool returned an error; inspect the ignored response")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "AOCI verification failed; inspect the local response and explicit tool configuration")
		os.Exit(1)
	}
}
