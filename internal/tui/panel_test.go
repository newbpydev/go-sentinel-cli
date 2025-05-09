package tui

import (
	"strings"
	"testing"
)

func TestPanelLeafBasic(t *testing.T) {
	panel := Panel{
		Content: []string{"Hello", "World"},
		Options: PanelOptions{
			Title:   "Title",
			Width:   12,
			Padding: 1,
			Border:  true,
		},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Title") {
		t.Errorf("Panel should contain title")
	}
	if !strings.Contains(joined, "Hello") || !strings.Contains(joined, "World") {
		t.Errorf("Panel should render content")
	}
}

func TestPanelFlexRow(t *testing.T) {
	child1 := &Panel{
		Content: []string{"A"},
		Options: PanelOptions{Width: 3, Border: true},
	}
	child2 := &Panel{
		Content: []string{"B"},
		Options: PanelOptions{Width: 3, Border: true},
	}
	panel := Panel{
		Options: PanelOptions{Flex: true, FlexDirection: "row", Gap: 1},
		Children: []*Panel{child1, child2},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "A") || !strings.Contains(joined, "B") {
		t.Errorf("Flex row should render both children")
	}
	if !strings.Contains(joined, " ") {
		t.Errorf("Flex row should include gap")
	}
}

func TestPanelFlexColumn(t *testing.T) {
	child1 := &Panel{
		Content: []string{"A"},
		Options: PanelOptions{Height: 1, Border: true},
	}
	child2 := &Panel{
		Content: []string{"B"},
		Options: PanelOptions{Height: 1, Border: true},
	}
	panel := Panel{
		Options: PanelOptions{Flex: true, FlexDirection: "column", Gap: 1},
		Children: []*Panel{child1, child2},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "A") || !strings.Contains(joined, "B") {
		t.Errorf("Flex column should render both children")
	}
	if !strings.Contains(joined, "\n\n") {
		t.Errorf("Flex column should include gap (empty line)")
	}
}

func TestPanelOrder(t *testing.T) {
	child1 := &Panel{
		Content: []string{"A"},
		Options: PanelOptions{Order: 2, Width: 3, Border: true},
	}
	child2 := &Panel{
		Content: []string{"B"},
		Options: PanelOptions{Order: 1, Width: 3, Border: true},
	}
	panel := Panel{
		Options: PanelOptions{Flex: true, FlexDirection: "row"},
		Children: []*Panel{child1, child2},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	firstIdx := strings.Index(joined, "B")
	secondIdx := strings.Index(joined, "A")
	if !(firstIdx > 0 && firstIdx < secondIdx) {
		t.Errorf("Order property should reorder children")
	}
}

func TestPanelGrow(t *testing.T) {
	child1 := &Panel{
		Content: []string{"A"},
		Options: PanelOptions{Grow: 2, Border: true},
	}
	child2 := &Panel{
		Content: []string{"B"},
		Options: PanelOptions{Grow: 1, Border: true},
	}
	panel := Panel{
		Options: PanelOptions{Flex: true, FlexDirection: "row", Width: 20},
		Children: []*Panel{child1, child2},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	countA := strings.Count(joined, "A")
	countB := strings.Count(joined, "B")
	if countA <= countB {
		t.Errorf("Grow property should allocate more space to child1: got A=%d B=%d\nRendered:\n%s", countA, countB, joined)
		for i, line := range lines {
			t.Logf("[%d] len=%d: %q", i, len(line), line)
		}
	}
}

func TestPanelAlignItemsCenter(t *testing.T) {
	child1 := &Panel{
		Content: []string{"A"},
		Options: PanelOptions{Height: 1, Border: true},
	}
	child2 := &Panel{
		Content: []string{"B", "B"},
		Options: PanelOptions{Height: 2, Border: true},
	}
	panel := Panel{
		Options: PanelOptions{Flex: true, FlexDirection: "column", AlignItems: "center", Height: 5},
		Children: []*Panel{child1, child2},
	}
	lines := panel.Render()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "A") || !strings.Contains(joined, "B") {
		t.Errorf("AlignItems center should render all children")
	}
	if !strings.HasPrefix(joined, "\n") {
		t.Errorf("AlignItems center should pad at the top\nRendered:\n%s", joined)
		for i, line := range lines {
			t.Logf("[%d] len=%d: %q", i, len(line), line)
		}
	}
}

// TODO: Add tests for JustifyContent, Shrink, Min/Max constraints, overflow, and advanced flex edge cases.
