package cmd

import (
	"strings"
	"testing"
)

func findFlagValue(flags []*Flag, name string) (string, bool) {
	for _, flag := range flags {
		if flag.Name == name {
			return flag.GetValue(), true
		}
	}
	return "", false
}

func TestParserSeparatesFixedAndDynamicFlags(t *testing.T) {
	testCtx := NewContext()
	args, err := NewParser([]string{
		"---profile", "prod",
		"---region", "ap-southeast-1",
		"--InstanceId", "i-123",
	}).ReadArgs(testCtx)
	if err != nil {
		t.Fatalf("ReadArgs returned error: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want none", args)
	}

	if got, ok := findFlagValue(testCtx.fixedFlags.GetFlags(), "profile"); !ok || got != "prod" {
		t.Fatalf("fixed profile flag = %q, exists=%v; want prod", got, ok)
	}
	if got, ok := findFlagValue(testCtx.fixedFlags.GetFlags(), "region"); !ok || got != "ap-southeast-1" {
		t.Fatalf("fixed region flag = %q, exists=%v; want ap-southeast-1", got, ok)
	}
	if got, ok := findFlagValue(testCtx.dynamicFlags.GetFlags(), "InstanceId"); !ok || got != "i-123" {
		t.Fatalf("dynamic InstanceId flag = %q, exists=%v; want i-123", got, ok)
	}
	if _, ok := findFlagValue(testCtx.dynamicFlags.GetFlags(), "-profile"); ok {
		t.Fatal("---profile must not be parsed as dynamic API flag -profile")
	}
}

func TestParserReturnsErrorWhenTrailingFlagHasNoValue(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "dynamic flag", args: []string{"--InstanceId"}, wantErr: "--InstanceId must set value."},
		{name: "fixed flag", args: []string{"---profile"}, wantErr: "---profile must set value."},
		{name: "fixed before dynamic", args: []string{"---profile", "--InstanceId"}, wantErr: "---profile must set value."},
		{name: "dynamic before fixed", args: []string{"--InstanceId", "---profile"}, wantErr: "--InstanceId must set value."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParser(tt.args).ReadArgs(NewContext())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParserReadArgsRejectsInvalidContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  *Context
	}{
		{name: "nil context", ctx: nil},
		{name: "empty context", ctx: &Context{}},
		{name: "missing dynamicFlags", ctx: &Context{fixedFlags: NewFlagSet()}},
		{name: "missing fixedFlags", ctx: &Context{dynamicFlags: NewFlagSet()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ReadArgs panicked: %v", r)
				}
			}()
			_, err := NewParser([]string{"--InstanceId", "i-123"}).ReadArgs(tt.ctx)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "invalid context") {
				t.Fatalf("error = %q, want invalid context", err.Error())
			}
		})
	}
}
