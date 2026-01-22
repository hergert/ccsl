package render

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/types"
)

var templateRe = regexp.MustCompile(`\{([-\w:.]+)(\?[^}]*)?\}`)
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// segmentPos tracks a segment's position in the rendered output
type segmentPos struct {
	id       string
	start    int // byte offset in result
	end      int // byte offset in result
	priority int
	rawText  string // original segment text (without ANSI)
}

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

	// Track positions during expansion
	var positions []segmentPos
	var result strings.Builder
	lastEnd := 0

	for _, match := range templateRe.FindAllStringSubmatchIndex(template, -1) {
		// Append literal text before this match
		result.WriteString(template[lastEnd:match[0]])

		// Extract key and query parts
		key := template[match[2]:match[3]]
		var query string
		if match[4] >= 0 && match[5] >= 0 {
			query = template[match[4]:match[5]]
		}

		seg, ok := segMap[key]
		if ok && seg.Text != "" {
			var prefix, suffix string
			if query != "" {
				q := strings.TrimPrefix(query, "?")
				vals, _ := url.ParseQuery(q)
				prefix = vals.Get("prefix")
				suffix = vals.Get("suffix")
			}

			// Track position before adding segment
			start := result.Len()
			result.WriteString(prefix)
			result.WriteString(pal.Apply(seg.Text, seg.Style))
			result.WriteString(suffix)

			positions = append(positions, segmentPos{
				id:       key,
				start:    start,
				end:      result.Len(),
				priority: seg.Priority,
				rawText:  seg.Text,
			})
		}

		lastEnd = match[1]
	}
	// Append remaining literal text
	result.WriteString(template[lastEnd:])

	out := result.String()
	if maxLen > 0 && visibleLen(out) > maxLen {
		out = truncateWithPositions(out, positions, maxLen)
	}

	return out
}

func truncateWithPositions(text string, positions []segmentPos, maxLen int) string {
	if maxLen <= 3 {
		return "..."[:maxLen]
	}

	// Guard: need at least one segment
	if len(positions) == 0 {
		return runesTruncate(text, maxLen-3) + "..."
	}

	// Find lowest priority segment to trim
	low := positions[0]
	for _, p := range positions[1:] {
		if p.priority < low.priority {
			low = p
		}
	}

	// Use tracked position for accurate replacement
	segText := text[low.start:low.end]
	budget := maxLen - (visibleLen(text) - visibleLen(segText))
	if budget > 3 {
		trimmed := runesTruncate(segText, budget-3) + "..."
		out := text[:low.start] + trimmed + text[low.end:]
		if visibleLen(out) <= maxLen {
			return out
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
