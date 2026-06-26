package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildActionInputExpandsJSONBodyFromFlatFlags(t *testing.T) {
	apiMeta := &ApiMeta{
		Request: &Meta{
			MetaTypes: map[string]*MetaType{
				"Items":    {TypeName: "array", TypeOf: "object"},
				"Enabled":  {TypeName: "boolean"},
				"Priority": {TypeName: "integer"},
				"Tags":     {TypeName: "array", TypeOf: "string"},
			},
			ChildMetas: map[string]*Meta{
				"Items": {
					MetaTypes: map[string]*MetaType{
						"Name":  {TypeName: "string"},
						"Count": {TypeName: "long"},
					},
				},
			},
		},
	}
	flags := []*Flag{
		{Name: "Items.1.Name", value: "first"},
		{Name: "Items.1.Count", value: "3"},
		{Name: "Items.2.Name", value: "second"},
		{Name: "Items.2.Count", value: "5"},
		{Name: "Enabled", value: "true"},
		{Name: "Priority", value: "9"},
		{Name: "Tags.1", value: "prod"},
	}

	got, fromBody, err := buildActionInput(flags, apiMeta, true)
	if err != nil {
		t.Fatalf("buildActionInput() error = %v", err)
	}
	if fromBody {
		t.Fatal("buildActionInput() fromBody = true, want false for flattened flags")
	}

	want := map[string]interface{}{
		"Items": []interface{}{
			map[string]interface{}{"Name": "first", "Count": int64(3)},
			map[string]interface{}{"Name": "second", "Count": int64(5)},
		},
		"Enabled":  true,
		"Priority": int64(9),
		"Tags":     []interface{}{"prod"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildActionInput() = %#v, want %#v", got, want)
	}
}

func TestBuildActionInputRejectsBodyWithFlatFlags(t *testing.T) {
	flags := []*Flag{
		{Name: "body", value: `{"Name":"demo"}`},
		{Name: "Name", value: "demo"},
	}

	_, _, err := buildActionInput(flags, nil, true)
	if err == nil {
		t.Fatal("buildActionInput() error = nil, want mutual exclusion error")
	}
	if !strings.Contains(err.Error(), "--body cannot be used together") {
		t.Fatalf("buildActionInput() error = %q, want mutual exclusion message", err)
	}
}

func TestBuildActionInputParsesJSONBodyObject(t *testing.T) {
	flags := []*Flag{{Name: "body", value: `{"Name":"demo"}`}}

	got, fromBody, err := buildActionInput(flags, nil, true)
	if err != nil {
		t.Fatalf("buildActionInput() error = %v", err)
	}
	if !fromBody {
		t.Fatal("buildActionInput() fromBody = false, want true")
	}

	want := &map[string]interface{}{"Name": "demo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildActionInput() = %#v, want %#v", got, want)
	}
}

