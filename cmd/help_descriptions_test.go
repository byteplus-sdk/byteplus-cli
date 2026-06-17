package cmd

import (
	"strings"
	"testing"
)

func TestUsageTemplatesIncludeFixedFlags(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "root", text: rootUsageTemplate()},
		{name: "service", text: serviceUsageTemplate()},
		{name: "action", text: actionUsageTemplate("", []string{"InstanceId string"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range expectedFixedFlagsForTest() {
				if !strings.Contains(tt.text, want) {
					t.Fatalf("%s usage missing %q:\n%s", tt.name, want, tt.text)
				}
			}
		})
	}
}

func expectedFixedFlagsForTest() []string {
	return []string{"---profile", "---region", "---endpoint", "---debug", "---debug-log-file"}
}
