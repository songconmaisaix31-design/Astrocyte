package events

import (
	"bytes"
	"encoding/json"
	"runtime"
	"testing"
)

func semanticFixture(name string, argument any) []byte {
	b, _ := json.Marshal(map[string]any{"type": "response_item", "timestamp": "2026-10-01T00:00:00Z", "payload": map[string]any{"type": "function_call", "name": name, "arguments": argument}})
	return b
}

func TestSemanticUnitsSeparateWrapperCallsWithoutShellDuplication(t *testing.T) {
	raw := semanticFixture("exec", `const results=await Promise.allSettled([tools.exec_command({cmd:"rg a; rg b; git status"}),tools.apply_patch("literal")]);text(results);`)
	rows, children, err := MapNativeSemanticTools("codex", "parent", "/fixture", raw, 1, NativeAssociation{}, 0)
	if err != nil || len(rows) != 2 || len(children) != 2 {
		t.Fatalf("unexpected units: %d/%d %v", len(rows), len(children), err)
	}
	for i, kind := range []string{"search", "edit_file"} {
		if rows[i].Event == nil || rows[i].Event.Kind != kind || !bytes.Equal(rows[i].RawBytes, raw) || rows[i].Observation.Tokens != nil || rows[i].Event.Topic != "" {
			t.Fatalf("child %d lost semantics/raw/unknown scope", i)
		}
		signal, e := NormalizeValidationSignal(rows[i])
		if e != nil || signal == nil || signal.Strength != 1 || signal.Weight != .6 {
			t.Fatal("weight multiplied incorrectly")
		}
	}
	if rows[0].OriginKey == rows[1].OriginKey || children[0].NonCognitiveSegments != 1 {
		t.Fatal("identity or segment reconciliation")
	}
}

func TestSemanticParserRetainsDynamicAndIgnoresQuotedCalls(t *testing.T) {
	cases := []struct {
		code    string
		buckets []string
	}{
		{`text("tools.apply_patch('fake')"); // tools.exec_command({cmd:'rg fake'})`, []string{"noncognitive_control"}},
		{`await tools.exec_command({cmd:"go test ./..."});await tools.apply_patch("patch");`, []string{"noncognitive_shell", "cognitive_edit"}},
		{`const x=await tools.exec_command({cmd:'rg a'});if(x){text(x);}`, []string{"cognitive_search"}},
		{`if(false){await tools.apply_patch("patch")}`, []string{"unresolved_structure"}},
		{`const f=()=>{tools.apply_patch("patch")};`, []string{"unresolved_structure"}},
		{`await tools.apply_patch(dynamic);await tools.exec_command({cmd:"rg x"})`, []string{"unresolved_structure", "cognitive_search"}},
		{"await tools.apply_patch(`patch ${secret}`)", []string{"unresolved_structure"}},
		{`await tools.exec("await tools.apply_patch('nested')")`, []string{"cognitive_edit"}},
		{`await tools.exec_command({cmd:'git commit -m attempt'})`, []string{"noncognitive_shell"}},
		{`await tools[toolName]({})`, []string{"unresolved_structure"}},
	}
	for i, c := range cases {
		children := SemanticTools("codex", semanticFixture("exec", c.code), 0)
		if len(children) != len(c.buckets) {
			t.Fatalf("case %d child count %d", i, len(children))
		}
		for j, b := range c.buckets {
			if children[j].Bucket != b {
				t.Fatalf("case %d child %d: %s", i, j, children[j].Bucket)
			}
		}
	}
}

func TestSemanticDirectPaperAndSearchAliases(t *testing.T) {
	for _, c := range []struct {
		name         string
		args         any
		bucket, kind string
	}{
		{"Read", map[string]string{"file_path": "/fixture/readme.md"}, "noncognitive_read", "unknown"},
		{"Read", map[string]string{"file_path": "/fixture/paper.pdf"}, "cognitive_paper", "read_paper"},
		{"mcp__morph_research__search_evidence", map[string]string{"query": "fixture"}, "cognitive_search", "search"},
	} {
		raw := semanticFixture(c.name, c.args)
		rows, children, err := MapNativeSemanticTools("codex", "parent", "/fixture", raw, 1, NativeAssociation{}, 0)
		if err != nil || len(children) != 1 || children[0].Bucket != c.bucket {
			t.Fatal("classification mismatch")
		}
		if c.kind == "unknown" {
			if rows[0].Event != nil {
				t.Fatal("false cognitive")
			}
		} else if rows[0].Event == nil || rows[0].Event.Kind != c.kind {
			t.Fatal("cognitive alias unmapped")
		}
	}
}

func TestSemanticLiteralConstDataExpressionsAndShellAST(t *testing.T) {
	for i, code := range []string{
		"const target='x';const cmd=['rg '+target,'git status'].join(';');await tools.exec_command({cmd});",
		"const target='x';await tools.exec_command({cmd:`rg ${target}`});",
		`await tools.exec_command({cmd:"$env:GOFLAGS='-p=1';go test ./...;if ($LASTEXITCODE) { exit 1 }"});`,
		`const script="python - <<'PY'\nprint('not a command: rg fake')\nPY\n";await tools.exec_command({cmd:script});`,
	} {
		children := SemanticTools("codex", semanticFixture("exec", code), 0)
		want := "cognitive_search"
		if i >= 2 {
			want = "noncognitive_shell"
		}
		if i == 2 && runtime.GOOS != "windows" {
			want = "unresolved_structure"
		} // no host PowerShell AST on this platform
		if len(children) != 1 || children[0].Bucket != want {
			t.Fatalf("literal grammar case %d bucket %v", i, children)
		}
	}
}

func TestStaticQuoteHelperConstantFoldingDoesNotEvaluateCode(t *testing.T) {
	code := `const quote = s => "'" + s.replace(/'/g,"''") + "'";const term="fixture";await tools.exec_command({cmd: "rg " + quote(term)});text(/tools.apply_patch("fake")/);`
	rows, children, err := MapNativeSemanticTools("codex", "parent", "/fixture", semanticFixture("exec", code), 1, NativeAssociation{}, 0)
	if err != nil || len(children) != 1 || children[0].Bucket != "cognitive_search" || rows[0].Event == nil {
		t.Fatalf("literal quote helper failed: err=%v children=%d bucket=%s reason=%s", err, len(children), children[0].Bucket, children[0].Reason)
	}
	malicious := `const quote = s => eval(s);await tools.exec_command({cmd:quote("rg fixture")});`
	if got := SemanticTools("codex", semanticFixture("exec", malicious), 0); len(got) != 1 || got[0].Bucket != "unresolved_structure" {
		t.Fatal("arbitrary helper evaluated")
	}
}

func TestInvalidWrapperNeverPromotesPrefixCalls(t *testing.T) {
	code := "await tools.apply_patch('patch');await tools.exec_command({cmd:`unterminated"
	rows, children, err := MapNativeSemanticTools("codex", "parent", "/fixture", semanticFixture("exec", code), 1, NativeAssociation{}, 0)
	if err != nil || len(rows) != 1 || len(children) != 1 || rows[0].Event != nil || children[0].Bucket != "unresolved_structure" {
		t.Fatal("invalid wrapper promoted a prefix or duplicated failures")
	}
}

func TestLiteralObjectPropertiesRequireUnmutatedBindings(t *testing.T) {
	for i, code := range []string{`const config={cmd:'rg fixture'};await tools.exec_command({cmd:config.cmd});`, `const config={cmd:'rg fixture'};config.cmd=dynamic;await tools.exec_command({cmd:config.cmd});`} {
		got := SemanticTools("codex", semanticFixture("exec", code), 0)
		want := "cognitive_search"
		if i == 1 {
			want = "unresolved_structure"
		}
		if len(got) != 1 || got[0].Bucket != want {
			t.Fatal("mutable constant guessed")
		}
	}
}

func TestNativeShellSyntaxBridgePreservesUTF8(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("native Windows syntax bridge not available; lexical grammar tested separately")
	}
	// Synthetic Unicode command name; ParseInput inspects it but never executes.
	result, available := parseShellAST("脱敏命令 -Value '引用中'")
	if !available || result.ParseErrors != 0 || len(result.Commands) != 1 || result.Commands[0].Head != "脱敏命令" {
		t.Fatal("syntax bridge corrupted UTF8 stdin")
	}
}

func TestClosedForOfCandidatesAgreeWithoutDuplicatingEvents(t *testing.T) {
	for i, code := range []string{
		`for(const cmd of ['rg first','rg second']){await tools.exec_command({cmd});}`,
		`for(const [cmd,workdir] of [['rg first','/fixture/a'],['rg second','/fixture/b']]){await tools.exec_command({cmd,workdir});}`,
		`const commands=['go test ./...','git status'];for(const cmd of commands){text(await tools.exec_command({cmd}));}`,
		`for(const file_path of ['first.pdf','second.pdf']){await tools.Read({file_path});}`,
		`for(const [cmd,workdir] of [['git status',null],['go test ./...','/fixture']]){await tools.exec_command({cmd,workdir});}`,
		`for(const [label,cmd] of [['a','rg first'],['b','rg second']]) text(await tools.exec_command({cmd}));`,
	} {
		rows, children, err := MapNativeSemanticTools("codex", "parent", "/fixture", semanticFixture("exec", code), 1, NativeAssociation{}, 0)
		want := "cognitive_search"
		if i == 2 || i == 4 {
			want = "noncognitive_shell"
		}
		if i == 3 {
			want = "cognitive_paper"
		}
		if err != nil || len(children) != 1 || len(rows) != 1 || children[0].Bucket != want || rows[0].OriginKey != "parent/semantic/0.0" {
			t.Fatalf("closed loop case %d failed stable one-call semantics: count=%d error=%v", i, len(children), err)
		}
		if i < 2 && (rows[0].Event == nil || rows[0].Event.Kind != "search") {
			t.Fatal("closed search candidates lost mapping")
		}
	}
}

func TestClosedForOfRejectsAmbiguityMutationAndRuntimeControl(t *testing.T) {
	for i, code := range []string{
		`for(const cmd of ['rg fixture','git status']){await tools.exec_command({cmd});}`,
		`for(const cmd of await fetchCommands()){await tools.exec_command({cmd});}`,
		`for(const cmd of ['rg first','rg second']){if(ready){await tools.exec_command({cmd});}}`,
		`const commands=['rg fixture'];commands.push(dynamic);for(const cmd of commands){await tools.exec_command({cmd});}`,
		`for(const pair of [['rg first'],['rg second']]){pair[0]=dynamic;await tools.exec_command({cmd:pair[0]});}`,
		`for(const cmd of ['rg first','rg second']){const f=()=>tools.exec_command({cmd});}`,
		`for(const file_path of ['first.pdf','readme.md']){await tools.Read({file_path});}`,
		`if(ready){for(const cmd of ['rg first','rg second']){await tools.exec_command({cmd});}}`,
		`for(const cmd of ['rg first','rg second']){await tools.exec_command({cmd:unknown(cmd)});}`,
		`for(const cmd of ['rg first','rg second']) if(ready) await tools.exec_command({cmd});`,
		`if(ready) for(const cmd of ['rg first','rg second']){await tools.exec_command({cmd});}`,
		`function run(){for(const cmd of ['rg first','rg second']){await tools.exec_command({cmd});}}`,
		`if(ready) for(const config of [{cmd:'rg first'},{cmd:'rg second'}]){await tools.exec_command({cmd:config.cmd});}`,
	} {
		children := SemanticTools("codex", semanticFixture("exec", code), 0)
		if len(children) != 1 || children[0].Bucket != "unresolved_structure" {
			t.Fatalf("unsafe closed loop case %d promoted a call", i)
		}
	}
}

func TestObservedRepositoryIdentityDoesNotGuessWorktreeSpelling(t *testing.T) {
	for _, c := range []struct{ value, want string }{{"C:/repo/Astrocyte/.git", "Astrocyte/Attention Engine"}, {"Astrocyte", "Astrocyte/Attention Engine"}, {"https://github.com/owner/maimai-monster.git", "McPet"}, {"https://github.com/songconmaisaix31-design/mcpet", "McPet"}, {"git@github.com:owner/Morphogenesis.git", "Morphogenesis"}, {"attention-engine-w4", ""}} {
		if got := repositoryTopic(repositoryIdentity(c.value)); got != c.want {
			t.Fatalf("repository identity got %s", got)
		}
	}
	o := ObserveRepository("/nonexistent/Astrocyte/attention-engine")
	if o.Topic != "" || o.Status != "missing_cwd" {
		t.Fatal("guessed unavailable repository")
	}
}
