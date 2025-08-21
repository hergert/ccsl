package render

import (
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
		// Simple parsing for prefix= and suffix=
		if strings.Contains(options, "prefix=") {
			prefixStart := strings.Index(options, "prefix=") + 7
			prefixEnd := strings.Index(options[prefixStart:], "&")
			if prefixEnd == -1 {
				prefixEnd = strings.Index(options[prefixStart:], "}")
				if prefixEnd == -1 {
					prefixEnd = len(options[prefixStart:])
				}
			}
			prefix = options[prefixStart : prefixStart+prefixEnd]
		}
		
		if strings.Contains(options, "suffix=") {
			suffixStart := strings.Index(options, "suffix=") + 7
			suffixEnd := strings.Index(options[suffixStart:], "&")
			if suffixEnd == -1 {
				suffixEnd = strings.Index(options[suffixStart:], "}")
				if suffixEnd == -1 {
					suffixEnd = len(options[suffixStart:])
				}
			}
			suffix = options[suffixStart : suffixStart+suffixEnd]
		}
	}

	// Apply styling
	styledText := pal.Apply(segment.Text, segment.Style)
	
	return prefix + styledText + suffix
}

func intelligentTruncate(text string, segments []types.Segment, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// For now, simple truncation with ellipsis
	// TODO: Implement priority-based truncation
	if maxLength <= 3 {
		return strings.Repeat(".", maxLength)
	}

	return text[:maxLength-3] + "..."
}