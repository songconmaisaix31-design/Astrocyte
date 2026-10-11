// Package mcp is the stdio MCP transport for the Astrocyte cognitive interface.
//
// It serves the five astrocyte.* tools from the task book §7 as a minimal
// newline-delimited JSON-RPC 2.0 server. The package carries no persistence and
// no attention-domain logic; the application layer supplies an MCPBackend that
// answers each tool with wire-shaped DTOs. It never orchestrates agents.
package mcp
