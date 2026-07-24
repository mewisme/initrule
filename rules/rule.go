package rules

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mewisme/agentrule/agents"
)

type Options struct {
	Cwd      string
	Update   bool   // --update / -u
	Target   string // auto|all|csv; empty = auto
	Location string // local|global; empty = local
}

type WorkResult struct {
	Label  string // completed result label
	Detail string // optional secondary info
}

type WorkItem struct {
	Rule  string
	Key   string
	Label string // pending/running label
	Run   func(context.Context, Options) (WorkResult, error)
}

type Step struct {
	Key   string // preinstall | install | postinstall | ...
	Label string
	Run   func(ctx context.Context, opts Options) (WorkResult, error)
}

type Rule struct {
	Name  string
	Steps []Step
}

var (
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	titleStyle = lipgloss.NewStyle().Bold(true)

	goos = runtime.GOOS // injectable for tests
)

func writeRule(opts Options, name string) error {
	body, err := content(name)
	if err != nil {
		return err
	}
	loc, err := agents.ParseLocation(opts.Location)
	if err != nil {
		return err
	}
	targets, err := agents.Resolve(loc, opts.Cwd, opts.Target)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no agent targets selected")
	}
	var wrote []string
	for _, t := range targets {
		res, err := t.WriteRule(loc, opts.Cwd, name, string(body))
		if err != nil {
			return fmt.Errorf("%s: %w", t.ID(), err)
		}
		if res.Action == "skipped" {
			continue
		}
		wrote = append(wrote, fmt.Sprintf("%s (%s)", t.DisplayName(), res.Action))
	}
	if len(wrote) == 0 {
		return fmt.Errorf("no agent wrote the rule (all skipped for --location=%s)", loc)
	}
	return nil
}

// All is the registered rule list (append new rules here).
func All() []Rule {
	out := []Rule{
		{
			Name: "codegraph",
			Steps: []Step{
				{Key: "preinstall", Label: "Ensure codegraph binary", Run: ensureCodegraphBinary},
				{Key: "install", Label: "Run: codegraph install (MCP)", Run: installCodegraphAgents},
				{Key: "index", Label: "Run: codegraph init", Run: initCodegraph},
				{Key: "postinstall", Label: "Wrote rule to agent targets", Run: writeRuleWork("codegraph")},
			},
		},
		{
			Name: "ponytail",
			Steps: []Step{
				{Key: "install", Label: "Wrote rule to agent targets", Run: writeRuleWork("ponytail")},
			},
		},
		{
			Name: "i-have-adhd",
			Steps: []Step{
				{Key: "install", Label: "Wrote rule to agent targets", Run: writeRuleWork("i-have-adhd")},
			},
		},
	}
	if goos == "windows" {
		out = append(out, Rule{
			Name: "powershell",
			Steps: []Step{
				{Key: "install", Label: "Wrote rule to agent targets", Run: writeRuleWork("powershell")},
			},
		})
	}
	return out
}

func ByName(name string) (Rule, bool) {
	for _, r := range All() {
		if r.Name == name {
			return r, true
		}
	}
	return Rule{}, false
}

func Names() []string {
	all := All()
	out := make([]string, len(all))
	for i, r := range all {
		out[i] = r.Name
	}
	return out
}

// WorkItems builds a side-effect-free execution plan for the named rules
// in registry order. Unknown names error before any Run closures execute.
func WorkItems(names []string, opts Options) ([]WorkItem, error) {
	_ = opts // reserved for plan closures; construction must stay side-effect free
	want := map[string]bool{}
	for _, n := range names {
		if _, ok := ByName(n); !ok {
			return nil, fmt.Errorf("unknown rule %q", n)
		}
		want[n] = true
	}
	var items []WorkItem
	for _, r := range All() {
		if !want[r.Name] {
			continue
		}
		for _, s := range r.Steps {
			if s.Run == nil {
				continue
			}
			items = append(items, WorkItem{
				Rule:  r.Name,
				Key:   s.Key,
				Label: s.Label,
				Run:   s.Run,
			})
		}
	}
	return items, nil
}

func pipe() { fmt.Println(dimStyle.Render("│")) }

func runRules(ruleList []Rule, opts Options) error {
	names := make([]string, len(ruleList))
	for i, r := range ruleList {
		names[i] = r.Name
	}
	items, err := WorkItems(names, opts)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s  %s\n", dimStyle.Render("┌"), titleStyle.Render("Installing rules"))
	pipe()

	loc := opts.Location
	if loc == "" {
		loc = "local"
	}
	tgt := opts.Target
	if tgt == "" {
		tgt = "auto"
	}
	fmt.Printf("%s  Initialized in %s\n", okStyle.Render("◆"), opts.Cwd)
	fmt.Printf("%s  Agents: %s · location: %s\n", dimStyle.Render("│"), tgt, loc)
	pipe()

	var currentRule string
	for _, item := range items {
		if item.Rule != currentRule {
			if currentRule != "" {
				pipe()
			}
			currentRule = item.Rule
			fmt.Printf("%s  %s\n", okStyle.Render("◆"), item.Rule)
		}
		result, err := item.Run(context.Background(), opts)
		if err != nil {
			fmt.Printf("%s  %s %s\n", dimStyle.Render("│"), errStyle.Render("*"), item.Label+" failed")
			pipe()
			fmt.Println(dimStyle.Render("└  ") + "Failed")
			return fmt.Errorf("%s: %s: %w", item.Rule, item.Key, err)
		}
		label := result.Label
		if label == "" {
			label = item.Label
		}
		fmt.Printf("%s  %s %s\n", dimStyle.Render("│"), okStyle.Render("*"), label)
		if result.Detail != "" {
			fmt.Printf("%s  %s\n", dimStyle.Render("│"), dimStyle.Render(result.Detail))
		}
	}
	if currentRule != "" {
		pipe()
	}

	fmt.Println(dimStyle.Render("└  ") + "Done")
	return nil
}

// ruleWriteSummary lists which agents received the rule (best-effort for display).
func ruleWriteSummary(opts Options, name string) string {
	loc, err := agents.ParseLocation(opts.Location)
	if err != nil {
		return ""
	}
	targets, err := agents.Resolve(loc, opts.Cwd, opts.Target)
	if err != nil || len(targets) == 0 {
		return ""
	}
	var ids []string
	for _, t := range targets {
		if !t.SupportsLocation(loc) {
			continue
		}
		// Skip known no-op instruction surfaces in the summary.
		switch t.ID() {
		case "hermes", "antigravity":
			continue
		}
		ids = append(ids, t.ID())
	}
	if len(ids) == 0 {
		return ""
	}
	return "Wrote " + name + " → " + strings.Join(ids, ", ")
}

// Run executes one rule inside the shared install timeline.
func Run(r Rule, opts Options) error {
	return runRules([]Rule{r}, opts)
}

// RunNames resolves and runs rules in registry order (not arg order).
func RunNames(names []string, opts Options) error {
	want := map[string]bool{}
	for _, n := range names {
		if _, ok := ByName(n); !ok {
			return fmt.Errorf("unknown rule %q", n)
		}
		want[n] = true
	}
	var rules []Rule
	for _, r := range All() {
		if want[r.Name] {
			rules = append(rules, r)
		}
	}
	return runRules(rules, opts)
}

func RunAll(opts Options) error {
	return runRules(All(), opts)
}
