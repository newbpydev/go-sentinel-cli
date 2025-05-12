package models

import (
	"encoding/json"
	"testing"

	runner "github.com/newbpydev/go-sentinel/internal/runner"
)



func TestConvertRunnerTestEventToAPIModel(t *testing.T) {
	in := runner.TestEvent{
		Time:    "2025-05-12T10:00:00Z",
		Action:  "pass",
		Package: "github.com/newbpydev/go-sentinel/internal/ui",
		Test:    "TestFoo",
		Output:  "",
		Elapsed: 0.12,
	}
	api := APITestEvent{
		Time:    in.Time,
		Action:  in.Action,
		Package: in.Package,
		Test:    in.Test,
		Output:  in.Output,
		Elapsed: in.Elapsed,
	}
	b, err := json.Marshal(api)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out APITestEvent
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Action != "pass" || out.Test != "TestFoo" || out.Elapsed != 0.12 {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestAPITestEventHandlesAllTypes(t *testing.T) {
	cases := []runner.TestEvent{
		{Action: "run", Test: "TestBar"},
		{Action: "fail", Test: "TestBaz", Output: "failure msg"},
		{Action: "output", Output: "some output"},
	}
	for _, c := range cases {
		api := APITestEvent{
			Time:    c.Time,
			Action:  c.Action,
			Package: c.Package,
			Test:    c.Test,
			Output:  c.Output,
			Elapsed: c.Elapsed,
		}
		b, err := json.Marshal(api)
		if err != nil {
			t.Errorf("marshal: %v", err)
		}
		var out APITestEvent
		if err := json.Unmarshal(b, &out); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		if out.Action != c.Action || out.Test != c.Test || out.Output != c.Output {
			t.Errorf("mismatch: got %+v want %+v", out, c)
		}
	}
}

func TestAPITestEventNestedSerialization(t *testing.T) {
	parent := APITestEvent{
		Time:    "2025-05-12T10:00:00Z",
		Action:  "run",
		Package: "pkg",
		Test:    "TestParent",
	}
	child := APITestEvent{
		Time:    "2025-05-12T10:00:01Z",
		Action:  "pass",
		Package: "pkg",
		Test:    "TestChild",
		Elapsed: 0.05,
	}
	wrapped := struct {
		Parent APITestEvent   `json:"parent"`
		Children []APITestEvent `json:"children"`
	}{
		Parent: parent,
		Children: []APITestEvent{child},
	}
	b, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("marshal nested: %v", err)
	}
	var out struct {
		Parent APITestEvent   `json:"parent"`
		Children []APITestEvent `json:"children"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal nested: %v", err)
	}
	if len(out.Children) != 1 || out.Children[0].Test != "TestChild" {
		t.Errorf("nested serialization failed: %+v", out)
	}
}
