//go:build windows

package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
)

// 固定 Parser.ParseInput 只构建语法树，stdout 只返回结构 JSON。
const shellASTProgram = `$ErrorActionPreference='Stop'
[Console]::InputEncoding=[System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding=[System.Text.UTF8Encoding]::new($false)
while ($null -ne ($line=[Console]::In.ReadLine())) {
  $request=ConvertFrom-Json -InputObject $line
  $tokens=$null;$errors=$null
  $ast=[System.Management.Automation.Language.Parser]::ParseInput([string]$request.cmd,[ref]$tokens,[ref]$errors)
  $commands=@($ast.FindAll({param($n) $n -is [System.Management.Automation.Language.CommandAst]},$true) | Sort-Object {$_.Extent.StartOffset} | ForEach-Object { $name=$_.GetCommandName(); @{head=[string]$name;dynamic=($null -eq $name)} })
  @{commands=$commands;parse_errors=@($errors).Count;error_ids=@($errors | ForEach-Object {$_.ErrorId})} | ConvertTo-Json -Depth 5 -Compress | ForEach-Object {[Console]::Out.WriteLine($_)}
}`

var shellASTProcess struct {
	sync.Mutex
	started bool
	command *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Scanner
	err     error
}

func parseShellAST(code string) (shellASTResult, bool) {
	p := &shellASTProcess
	p.Lock()
	defer p.Unlock()
	if !p.started {
		p.started = true
		binary, e := exec.LookPath("pwsh")
		if e != nil {
			p.err = e
			return shellASTResult{ParseErrors: 1}, true
		}
		p.command = exec.Command(binary, "-NoProfile", "-NonInteractive", "-Command", shellASTProgram)
		p.command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		p.input, p.err = p.command.StdinPipe()
		if p.err != nil {
			return shellASTResult{ParseErrors: 1}, true
		}
		output, e := p.command.StdoutPipe()
		if e != nil {
			p.err = e
			return shellASTResult{ParseErrors: 1}, true
		}
		p.command.Stderr = io.Discard
		p.output = bufio.NewScanner(output)
		p.output.Buffer(make([]byte, 4096), 8*1024*1024)
		p.err = p.command.Start()
	}
	if p.err != nil {
		return shellASTResult{ParseErrors: 1}, true
	}
	data, e := json.Marshal(map[string]string{"cmd": code})
	if e != nil {
		return shellASTResult{ParseErrors: 1}, true
	}
	if _, e = fmt.Fprintln(p.input, string(data)); e != nil {
		p.err = e
		return shellASTResult{ParseErrors: 1}, true
	}
	if !p.output.Scan() {
		p.err = fmt.Errorf("syntax parser unavailable")
		return shellASTResult{ParseErrors: 1}, true
	}
	var r shellASTResult
	if json.Unmarshal(p.output.Bytes(), &r) != nil {
		return shellASTResult{ParseErrors: 1}, true
	}
	return r, true
}
