package render

import (
	"net/url"
	"regexp"
	"strings"

	"ccsl/internal/types"
	"ccsl/internal/palette"
)

var templateRegex = regexp.MustCompile(`\{(\w+)(\?[^}]*)?\}`)

// Line renders segments using a template string and applies styling
func Line(template string, segments []types.Segment, pal *palette.Palette, maxLength int) string {
	// Create segment lookup map
	segMap := make(map[string]types.Segment)
	for _, seg := range segments {
		segMap[seg.ID] = seg
	}

	// Process template substitutions
	result := templateRegex.ReplaceAllStringFunc(template, func(match string) string {
		return processTemplateMatch(match, segMap, pal)
	})

	// Apply final truncation if needed
	if len(result) > maxLength {
		result = intelligentTruncate(result, segments, maxLength)
	}

	return result
}

func processTemplateMatch(match string, segMap map[string]types.Segment, pal *palette.Palette) string {
	// Parse the match: {segment?prefix=...&suffix=...}
	parts := templateRegex.FindStringSubmatch(match)
	if len(parts) < 2 {
		return match
	}

	segmentID := parts[1]
	options := ""
	if len(parts) > 2 {
		options = parts[2]
	}

	segment, exists := segMap[segmentID]
	if !exists || segment.Text == "" {
		return "" // segment doesn't exist or is empty
	}

	// Parse options
	var prefix, suffix string
	if options != "" {
		q := strings.TrimPrefix(options, "?")
		vals, _ := url.ParseQuery(q)
		prefix = vals.Get("prefix")
		suffix = vals.Get("suffix")
	}

	// Apply styling
	styledText := pal.Apply(segment.Text, segment.Style)
	
	return prefix + styledText + suffix
}

func intelligentTruncate(text string, segments []types.Segment, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// 1) Find the lowest-priority segment (often 'prompt') and trim its contribution.
	// We do a best-effort pass: try to shorten the last occurrence of that segment's text.
	if maxLength > 3 {
		low := segments[0]
		for _, s := range segments {
			if s.Priority < low.Priority { low = s }
		}
		if low.Text != "" {
			idx := strings.LastIndex(text, low.Text)
			if idx >= 0 {
				// Available budget when replacing this segment with a shorter version
				keep := maxLength - (len(text) - len(low.Text))
				if keep > 3 {
					trimmed := wordTrim(low.Text, keep-3) + "..."
					out := text[:idx] + trimmed + text[idx+len(low.Text):]
					if len(out) <= maxLength {
						return out
					}
					text = out // fallthrough to hard cut
				}
			}
		}
	}
	// 2) Hard cut with ellipsis at the very end as last resort.
	if maxLength <= 3 { return strings.Repeat(".", maxLength) }
	return text[:maxLength-3] + "..."
}

// wordTrim tries to cut on a word boundary.
func wordTrim(s string, n int) string {
	if len(s) <= n { return s }
	cut := s[:n]
	if i := strings.LastIndexAny(cut, " \t"); i > n/2 {
		return strings.TrimSpace(cut[:i])
	}
	return strings.TrimSpace(cut)
}