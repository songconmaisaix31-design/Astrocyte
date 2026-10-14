package events

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

// Opt-in local source audit. Never embeds real source text in test output/Git.
// The normal suite cannot claim this check passed when the sidecar is absent.
func TestLocalV22CookedCommandByteParity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("local native Windows parser cross-check unavailable on this platform")
	}
	expectedPath, inputPath := os.Getenv("ASTROCYTE_V22_COOKED_LOCAL"), os.Getenv("ASTROCYTE_V22_FROZEN_LOCAL")
	if expectedPath == "" || inputPath == "" {
		t.Skip("explicit local frozen input and independent cooked sidecar required")
	}
	b, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal("local sidecar unavailable")
	}
	var expected []struct {
		EventID string `json:"event_id"`
		Bucket  string `json:"bucket"`
		Command string `json:"command"`
	}
	if json.Unmarshal(b, &expected) != nil {
		t.Fatal("invalid local sidecar")
	}
	wanted := map[string]map[string]string{}
	for _, row := range expected {
		if row.Bucket != "unresolved_structure" {
			continue
		}
		parts := strings.Split(row.EventID, "/semantic/")
		if len(parts) != 2 {
			t.Fatal("invalid safe source locator")
		}
		if wanted[parts[0]] == nil {
			wanted[parts[0]] = map[string]string{}
		}
		wanted[parts[0]][parts[1]] = row.Command
	}
	file, err := os.Open(inputPath)
	if err != nil {
		t.Fatal("frozen input unavailable")
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	scan.Buffer(make([]byte, 4096), 64*1024*1024)
	checked := 0
	for scan.Scan() {
		var v struct {
			Source          string `json:"source"`
			FileIndex, Line int
			Blocks          []int  `json:"blocks"`
			RawBase64       string `json:"raw_base64"`
		}
		// Explicit tags for the frozen file ordinal.
		var envelope map[string]json.RawMessage
		if json.Unmarshal(scan.Bytes(), &envelope) != nil {
			t.Fatal("invalid frozen envelope")
		}
		_ = json.Unmarshal(envelope["source"], &v.Source)
		_ = json.Unmarshal(envelope["file_index"], &v.FileIndex)
		_ = json.Unmarshal(envelope["line"], &v.Line)
		_ = json.Unmarshal(envelope["blocks"], &v.Blocks)
		_ = json.Unmarshal(envelope["raw_base64"], &v.RawBase64)
		for _, block := range v.Blocks {
			parent := safeParentLocator(v.Source, v.FileIndex, v.Line, block)
			paths := wanted[parent]
			if paths == nil {
				continue
			}
			raw, err := base64.StdEncoding.DecodeString(v.RawBase64)
			if err != nil {
				t.Fatal("invalid frozen encoding")
			}
			children, err := SemanticTools(v.Source, raw, block)
			if err != nil {
				t.Fatal(err)
			}
			for _, child := range children {
				want, ok := paths[child.Path]
				if !ok {
					continue
				}
				got := str(object(child.args), "cmd")
				if child.Tool != "exec_command" || len(child.args) == 0 {
					continue
				}
				checked++
				original, _ := parseShellAST(got)
				t.Logf("safe event %s/semantic/%s original_errors=%d", parent, child.Path, original.ParseErrors)
				if original.ParseErrors != 0 {
					t.Error("independently valid cooked input became a native syntax failure")
				}
				if got != want {
					at := 0
					for at < len(got) && at < len(want) && got[at] == want[at] {
						at++
					}
					t.Errorf("safe event %s/semantic/%s cooked byte mismatch at %d lengths %d/%d", parent, child.Path, at, len(got), len(want))
				}
			}
		}
	}
	if scan.Err() != nil {
		t.Fatal("frozen input read error")
	}
	if checked == 0 {
		t.Fatal("no independent source pairs checked")
	}
	t.Logf("local source pairs checked=%d", checked)
}

func safeParentLocator(source string, file, line, block int) string {
	return fmt.Sprintf("native:%s:%d:%d:block:%d", source, file, line, block)
}
