package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/mewisme/agentrule/agents"
)

type selectItem struct {
	id      string
	title   string
	desc    string
	checked bool
	multi   bool
}

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.desc }
func (i selectItem) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(selectItem)
	if !ok {
		return
	}

	on := index == m.Index()
	var box string
	if i.multi {
		// Real selection, or fake-checked while the cursor is on the row.
		if i.checked || on {
			box = "[x]"
		} else {
			box = "[ ]"
		}
	} else if on {
		box = "[•]"
	} else {
		box = "[ ]"
	}

	body := fmt.Sprintf("%s %s", box, i.title)
	if i.checked || on {
		body = selectedStyle.Render(body)
	}
	if i.desc != "" {
		body += dimStyle.Render("  " + i.desc)
	}
	fmt.Fprint(w, body)
}

func newSelectList(items []list.Item, title string, _ bool) list.Model {
	l := list.New(items, itemDelegate{}, 40, 10)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.Styles.Title = titleStyle
	return l
}

func ruleItems(names []string, selected map[string]bool) []list.Item {
	items := make([]list.Item, len(names))
	for i, n := range names {
		items[i] = selectItem{id: n, title: n, checked: selected[n], multi: true}
	}
	return items
}

func locationItems(locs []agents.Location) []list.Item {
	items := make([]list.Item, len(locs))
	for i, loc := range locs {
		hint := "this project"
		if loc == agents.Global {
			hint = "all projects"
		}
		items[i] = selectItem{id: string(loc), title: string(loc), desc: hint, multi: false}
	}
	return items
}

func agentListItems(infos []agentInfo, selected map[string]bool) []list.Item {
	items := make([]list.Item, len(infos))
	for i, a := range infos {
		items[i] = selectItem{id: a.id, title: a.name, checked: selected[a.id], multi: true}
	}
	return items
}

func (m *model) toggleCurrent() tea.Cmd {
	item, ok := m.list.SelectedItem().(selectItem)
	if !ok || !item.multi {
		return nil
	}
	item.checked = !item.checked
	return m.list.SetItem(m.list.Index(), item)
}

func (m *model) toggleAll() tea.Cmd {
	items := m.list.Items()
	if len(items) == 0 {
		return nil
	}
	allOn := true
	for _, it := range items {
		si, ok := it.(selectItem)
		if !ok || !si.multi {
			continue
		}
		if !si.checked {
			allOn = false
			break
		}
	}
	newItems := make([]list.Item, len(items))
	for i, it := range items {
		si, ok := it.(selectItem)
		if !ok {
			newItems[i] = it
			continue
		}
		if si.multi {
			si.checked = !allOn
		}
		newItems[i] = si
	}
	return m.list.SetItems(newItems)
}

func (m model) selectedIDs() []string {
	var out []string
	for _, it := range m.list.Items() {
		si, ok := it.(selectItem)
		if ok && si.checked {
			out = append(out, si.id)
		}
	}
	return out
}

func (m model) selectedIDSet() map[string]bool {
	out := map[string]bool{}
	for _, id := range m.selectedIDs() {
		out[id] = true
	}
	return out
}

func (m *model) applyListSize() {
	w, h := m.listSize()
	m.list.SetSize(w, h)
}

func (m model) listSize() (int, int) {
	w := m.width - 4
	if w < 1 {
		w = 1
	}
	h := m.height - BannerHeight() - 4 - 3 // banner, title gap, footer
	if h < 1 {
		h = 1
	}
	return w, h
}

func hasChecked(items []list.Item) bool {
	for _, it := range items {
		if si, ok := it.(selectItem); ok && si.checked {
			return true
		}
	}
	return false
}

func clampIndex(i, n int) int {
	if n <= 0 {
		return 0
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}
