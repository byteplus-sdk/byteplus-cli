package cmd

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestFormatServiceShortUsesExplorerAsset(t *testing.T) {
	restore := replaceExplorerDescriptionsAsset(func() ([]byte, error) {
		return []byte(`{"services":{"custom":{"service_en":"Custom BytePlus service"}},"apis":{}}`), nil
	})
	defer restore()

	got := formatServiceShort("custom")
	if got != "Custom BytePlus service" {
		t.Fatalf("formatServiceShort() = %q, want asset description", got)
	}
}

func TestFormatActionLongUsesExplorerAsset(t *testing.T) {
	restore := replaceExplorerDescriptionsAsset(func() ([]byte, error) {
		return []byte(`{"services":{},"apis":{"custom":{"DescribeThing":{"name_en":"Describe thing","description_en":"Returns the requested thing."}}}}`), nil
	})
	defer restore()

	short := formatActionShort("custom", "DescribeThing")
	if short != "Describe thing" {
		t.Fatalf("formatActionShort() = %q, want English api name", short)
	}

	got := formatActionLong("custom", "DescribeThing")
	if !strings.Contains(got, "Returns the requested thing.") {
		t.Fatalf("formatActionLong() = %q, want asset api description", got)
	}
}

func TestExplorerDescriptionsFallbackWhenAssetMissing(t *testing.T) {
	restore := replaceExplorerDescriptionsAsset(func() ([]byte, error) {
		return nil, errors.New("asset missing")
	})
	defer restore()

	got := formatServiceShort("sts")
	if !strings.Contains(got, "Security Token Service") {
		t.Fatalf("formatServiceShort() = %q, want fallback service description", got)
	}
}

func replaceExplorerDescriptionsAsset(fn func() ([]byte, error)) func() {
	oldFn := loadExplorerDescriptionsAsset
	oldOnce := explorerDescriptionsOnce
	oldData := explorerDescriptions

	loadExplorerDescriptionsAsset = fn
	explorerDescriptionsOnce = sync.Once{}
	explorerDescriptions = explorerDescriptionsData{}

	return func() {
		loadExplorerDescriptionsAsset = oldFn
		explorerDescriptionsOnce = oldOnce
		explorerDescriptions = oldData
	}
}
