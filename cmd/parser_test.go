package cmd

import (
	"strings"
	"testing"
)

func TestParserReadsFixedFlags(t *testing.T) {
	ctx := NewContext()
	parser := NewParser([]string{
		"---profile", "release",
		"---region", "ap-southeast-1",
		"---endpoint", "sts.byteplusapi.com",
		"--Limit", "10",
	})

	args, err := parser.ReadArgs(ctx)
	if err != nil {
		t.Fatalf("ReadArgs() error = %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("ReadArgs() args = %v, want empty", args)
	}
	if got := ctx.fixedFlags.GetByName("profile").GetValue(); got != "release" {
		t.Fatalf("profile fixed flag = %q, want release", got)
	}
	if got := ctx.fixedFlags.GetByName("region").GetValue(); got != "ap-southeast-1" {
		t.Fatalf("region fixed flag = %q, want ap-southeast-1", got)
	}
	if got := ctx.fixedFlags.GetByName("endpoint").GetValue(); got != "sts.byteplusapi.com" {
		t.Fatalf("endpoint fixed flag = %q, want sts.byteplusapi.com", got)
	}
	if got := ctx.dynamicFlags.GetByName("Limit").GetValue(); got != "10" {
		t.Fatalf("dynamic flag Limit = %q, want 10", got)
	}
}

func TestParserRejectsUnsupportedFixedFlag(t *testing.T) {
	ctx := NewContext()
	parser := NewParser([]string{"---debug", "true"})

	_, err := parser.ReadArgs(ctx)
	if err == nil {
		t.Fatal("ReadArgs() error = nil, want unsupported fixed flag error")
	}
	if !strings.Contains(err.Error(), "---debug is not supported") {
		t.Fatalf("ReadArgs() error = %q, want unsupported fixed flag message", err)
	}
}

func TestParserRequiresFixedFlagValue(t *testing.T) {
	ctx := NewContext()
	parser := NewParser([]string{"---region"})

	_, err := parser.ReadArgs(ctx)
	if err == nil {
		t.Fatal("ReadArgs() error = nil, want missing fixed flag value error")
	}
	if !strings.Contains(err.Error(), "---region must set value") {
		t.Fatalf("ReadArgs() error = %q, want missing value message", err)
	}
}
