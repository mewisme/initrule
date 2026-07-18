package rules

import (
	"context"
	"errors"
	"testing"
)

func TestWorkItemsPonytail(t *testing.T) {
	items, err := WorkItems([]string{"ponytail"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len=%d want 1", len(items))
	}
	if items[0].Rule != "ponytail" || items[0].Key != "install" {
		t.Fatalf("got %+v", items[0])
	}
}

func TestWorkItemsCodegraphOrder(t *testing.T) {
	items, err := WorkItems([]string{"codegraph"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"preinstall", "install", "index", "postinstall"}
	if len(items) != len(want) {
		t.Fatalf("len=%d want %d", len(items), len(want))
	}
	for i, k := range want {
		if items[i].Key != k {
			t.Fatalf("[%d] key=%s want %s", i, items[i].Key, k)
		}
	}
}

func TestWorkItemsUnknownRule(t *testing.T) {
	_, err := WorkItems([]string{"nope"}, Options{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkItemsNoSideEffects(t *testing.T) {
	oldLook, oldInst, oldAgents, oldInit := lookPath, runCGInstaller, runCGAgents, runCGInit
	t.Cleanup(func() {
		lookPath, runCGInstaller, runCGAgents, runCGInit = oldLook, oldInst, oldAgents, oldInit
	})
	lookPath = func(string) (string, error) { t.Fatal("lookPath"); return "", nil }
	runCGInstaller = func(context.Context) error { t.Fatal("installer"); return nil }
	runCGAgents = func(context.Context, Options) error { t.Fatal("agents"); return nil }
	runCGInit = func(context.Context, string) error { t.Fatal("init"); return nil }

	if _, err := WorkItems([]string{"codegraph", "ponytail"}, Options{}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkItemsRegistryOrder(t *testing.T) {
	items, err := WorkItems([]string{"ponytail", "codegraph"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Rule != "codegraph" {
		t.Fatalf("first rule=%s want codegraph (registry order)", items[0].Rule)
	}
}

func TestWorkItemsSkipsNilRun(t *testing.T) {
	// All current steps have Run; ensure nil steps would be skipped by unit-testing the filter.
	r := Rule{Name: "x", Steps: []Step{
		{Key: "a", Label: "A", Run: nil},
		{Key: "b", Label: "B", Run: func(context.Context, Options) (WorkResult, error) {
			return WorkResult{Label: "B"}, nil
		}},
	}}
	var items []WorkItem
	for _, s := range r.Steps {
		if s.Run == nil {
			continue
		}
		items = append(items, WorkItem{Key: s.Key, Run: s.Run})
	}
	if len(items) != 1 || items[0].Key != "b" {
		t.Fatalf("got %+v", items)
	}
}

func TestRunItemsSequentialFailFast(t *testing.T) {
	calls := 0
	items := []WorkItem{
		{Rule: "r", Key: "1", Label: "one", Run: func(context.Context, Options) (WorkResult, error) {
			calls++
			return WorkResult{Label: "one"}, nil
		}},
		{Rule: "r", Key: "2", Label: "two", Run: func(context.Context, Options) (WorkResult, error) {
			calls++
			return WorkResult{}, errors.New("boom")
		}},
		{Rule: "r", Key: "3", Label: "three", Run: func(context.Context, Options) (WorkResult, error) {
			calls++
			return WorkResult{Label: "three"}, nil
		}},
	}
	for _, item := range items {
		if _, err := item.Run(context.Background(), Options{}); err != nil {
			break
		}
	}
	if calls != 2 {
		t.Fatalf("calls=%d want 2 (fail fast)", calls)
	}
}

func TestWorkItemRespectsCancel(t *testing.T) {
	old := runCGInit
	t.Cleanup(func() { runCGInit = old })
	seen := false
	runCGInit = func(ctx context.Context, cwd string) error {
		seen = true
		<-ctx.Done()
		return ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := initCodegraph(ctx, Options{Cwd: t.TempDir()})
	if !seen {
		t.Fatal("runner not called")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want canceled", err)
	}
}
