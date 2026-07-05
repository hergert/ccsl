package effort

import "testing"

func TestParse(t *testing.T) {
	e, ok := Parse(map[string]any{"effort": map[string]any{"level": "max"}})
	if !ok || e.Level != "max" {
		t.Errorf("Parse = %+v ok=%v, want Level=max ok=true", e, ok)
	}
	if _, ok := Parse(map[string]any{}); ok {
		t.Error("expected ok=false without effort")
	}
	if _, ok := Parse(map[string]any{"effort": map[string]any{"level": ""}}); ok {
		t.Error("expected ok=false for empty level")
	}
}

func TestRender(t *testing.T) {
	seg := Effort{Level: "xhigh"}.Render()
	if seg.Text != "xhigh" || seg.Style != "dim" {
		t.Errorf("Render = %+v, want Text=xhigh Style=dim", seg)
	}
	if seg.Priority != 32 {
		t.Errorf("Priority = %d, want 32", seg.Priority)
	}
}
