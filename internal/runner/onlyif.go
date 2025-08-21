package runner

import (
	"fmt"
	"strings"
)

// Very small expression support:
// - "has(a.b.c)"
// - "eq(a.b, "value")", "ne(...)" (value may be unquoted numeric/bool)
// - "a.b.c" (truthy check)
func shouldRun(root map[string]any, expr string) bool {
	expr = strings.TrimSpace(expr)
	switch {
	case strings.HasPrefix(expr, "has(") && strings.HasSuffix(expr, ")"):
		path := strings.TrimSuffix(strings.TrimPrefix(expr, "has("), ")")
		_, ok := lookupPath(root, strings.TrimSpace(path))
		return ok
	case strings.HasPrefix(expr, "eq(") && strings.HasSuffix(expr, ")"):
		args := strings.TrimSuffix(strings.TrimPrefix(expr, "eq("), ")")
		p, v := split2(args, ",")
		val, ok := lookupPath(root, strings.TrimSpace(p))
		if !ok { return false }
		return stringify(val) == trimQuotes(strings.TrimSpace(v))
	case strings.HasPrefix(expr, "ne(") && strings.HasSuffix(expr, ")"):
		args := strings.TrimSuffix(strings.TrimPrefix(expr, "ne("), ")")
		p, v := split2(args, ",")
		val, ok := lookupPath(root, strings.TrimSpace(p))
		if !ok { return false }
		return stringify(val) != trimQuotes(strings.TrimSpace(v))
	default:
		// treat as path truthiness
		val, ok := lookupPath(root, expr)
		if !ok { return false }
		s := stringify(val)
		return s != "" && s != "0" && s != "false"
	}
}

func split2(s, sep string) (string, string) {
	i := strings.Index(s, sep)
	if i < 0 { return s, "" }
	return s[:i], s[i+1:]
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1:len(s)-1]
	}
	return s
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t { return "true" }
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func lookupPath(root map[string]any, path string) (any, bool) {
	cur := any(root)
	for _, k := range strings.Split(strings.TrimSpace(path), ".") {
		m, ok := cur.(map[string]any)
		if !ok { return nil, false }
		cur, ok = m[k]
		if !ok { return nil, false }
	}
	return cur, true
}