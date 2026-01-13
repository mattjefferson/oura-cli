package app

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/mattjefferson/oura-cli/internal/oura"
	"github.com/mattjefferson/oura-cli/internal/output"
)

func runResources(printer *output.Printer) int {
	resources := oura.Resources()
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Key < resources[j].Key
	})

	if printer.JSONCompact {
		entries := make([]map[string]any, 0, len(resources))
		for _, r := range resources {
			entries = append(entries, map[string]any{
				"name":  r.Key,
				"list":  r.SupportsList,
				"get":   r.SupportsGet,
				"query": queryLabel(r.Query),
				"path":  r.PathSegment,
			})
		}
		b, err := json.Marshal(entries)
		if err != nil {
			printer.Errorf("json encode failed: %v", err)
			return 1
		}
		if err := printer.PrintJSON(b); err != nil {
			printer.Errorf("output failed: %v", err)
			return 1
		}
		return 0
	}

	lines := make([]string, 0, len(resources))
	for _, r := range resources {
		line := r.Key + " (" + strings.TrimSpace(queryLabel(r.Query)) + ")"
		lines = append(lines, line)
	}
	printer.Write(strings.Join(lines, "\n") + "\n")
	return 0
}

func queryLabel(q oura.QueryKind) string {
	switch q {
	case oura.QueryNone:
		return "no filters"
	case oura.QueryNextTokenOnly:
		return "next_token"
	case oura.QueryDate:
		return "date"
	case oura.QueryDateTime:
		return "datetime"
	default:
		return "unknown"
	}
}
