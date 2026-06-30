// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type inT struct {
	X string `json:"x" jsonschema:"a value"`
}
type outT struct {
	Y string `json:"y" jsonschema:"a result"`
}

func handler(context.Context, *mcp.CallToolRequest, inT) (*mcp.CallToolResult, outT, error) {
	return nil, outT{}, nil
}

func TestRegister_ReadOnlySuppressesWriteTools(t *testing.T) {
	s := New("test", "v0", true) // read-only

	Register(s, ToolDef{Name: "read_tool"}, handler)
	Register(s, ToolDef{Name: "write_tool", Write: true}, handler)

	if got := s.ToolCount(); got != 1 {
		t.Fatalf("ToolCount = %d, want 1 (write tool suppressed)", got)
	}
}

func TestRegister_ReadWriteRegistersBoth(t *testing.T) {
	s := New("test", "v0", false)

	Register(s, ToolDef{Name: "read_tool"}, handler)
	Register(s, ToolDef{Name: "write_tool", Write: true}, handler)

	if got := s.ToolCount(); got != 2 {
		t.Fatalf("ToolCount = %d, want 2", got)
	}
}

func TestNoteToolset(t *testing.T) {
	s := New("test", "v0", false)
	s.NoteToolset("core")
	s.NoteToolset("wit")
	got := s.Toolsets()
	if len(got) != 2 || got[0] != "core" || got[1] != "wit" {
		t.Errorf("Toolsets = %v, want [core wit]", got)
	}
}

func TestList(t *testing.T) {
	lr := List([]int{1, 2, 3})
	if lr.Count != 3 || len(lr.Items) != 3 {
		t.Errorf("List() = %+v, want count 3", lr)
	}
}
