package render

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"ccsl/internal/palette"
	"ccsl/internal/types"
)

var templateRe = regexp.MustCompile(`\{([-\w:.]+)(\?[^}]*)?\}`)
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleLen returns the visible length excluding ANSI codes
func visibleLen(s string) int {
	return utf8.RuneCountInString(ansiRe.ReplaceAllString(s, ""))
}

// Line renders a full statusline based on a template and segments
func Line(template string, segments []types.Segment, pal *palette.Palette, maxLen int) string {
	segMap := make(map[string]types.Segment)
	for _, seg := range segments {
		segMap[seg.ID] = seg
	}

	result := templateRe.ReplaceAllStringFunc(template, func(match string) string {
		return processMatch(match, segMap, pal)
	})

	if maxLen > 0 && visibleLen(result) > maxLen {
		result = truncate(result, segments, maxLen)
	}

	return result
}

func processMatch(match string, segMap map[string]types.Segment, pal *palette.Palette) string {
	parts := templateRe.FindStringSubmatch(match)
	if len(parts) < 2 {
		return match
	}

	key := parts[1]
	seg, ok := segMap[key]
	if !ok || seg.Text == "" {
		return ""
	}

	var prefix, suffix string
	if len(parts) > 2 && parts[2] != "" {
		q := strings.TrimPrefix(parts[2], "?")
		vals, _ := url.ParseQuery(q)
		prefix = vals.Get("prefix")
		suffix = vals.Get("suffix")
	}

	return prefix + pal.Apply(seg.Text, seg.Style) + suffix
}

func truncate(text string, segments []types.Segment, maxLen int) string {
	if maxLen <= 3 {
		return "..."[:maxLen]
	}

	// Guard: need at least one segment
	if len(segments) == 0 {
		return runesTruncate(text, maxLen-3) + "..."
	}

	// Find lowest priority segment to trim
	low := segments[0]
	for _, s := range segments[1:] {
		if s.Priority < low.Priority {
			low = s
		}
	}

	if low.Text != "" {
		idx := strings.LastIndex(text, low.Text)
		if idx >= 0 {
			budget := maxLen - (visibleLen(text) - visibleLen(low.Text))
			if budget > 3 {
				trimmed := runesTruncate(low.Text, budget-3) + "..."
				out := text[:idx] + trimmed + text[idx+len(low.Text):]
				if visibleLen(out) <= maxLen {
					return out
				}
			}
		}
	}

	return runesTruncate(text, maxLen-3) + "..."
}

// runesTruncate truncates by visible runes (UTF-8 safe, ANSI-aware)
func runesTruncate(s string, n int) string {
	if n <= 0 {
		return ""
	}

	var result strings.Builder
	visible := 0
	i := 0

	for i < len(s) && visible < n {
		// Check for ANSI escape sequence
		if loc := ansiRe.FindStringIndex(s[i:]); loc != nil && loc[0] == 0 {
			result.WriteString(s[i : i+loc[1]])
			i += loc[1]
			continue
		}

		// Decode next rune
		r, size := utf8.DecodeRuneInString(s[i:])
		result.WriteRune(r)
		i += size
		visible++
	}

	return result.String()
}
