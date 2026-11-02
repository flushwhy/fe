package changelog

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

var conventional = regexp.MustCompile(`^([a-z]+)(!?)(?:\(([^)]+)\))?:\s*(.+)$`)

type Entry struct {
	Type    string
	Subject string
	Breaking bool
	Details string
}

var headings = map[string]string{
	"feat": "Features",
	"fix": "Bug Fixes",
	"perf": "Performance",
	"refactor": "Refactoring",
	"docs": "Documentation",
	"test": "Tests",
	"build": "Build",
	"ci": "CI",
	"chore": "Maintenance",
}

func Generate(root, from, to string) (string, error) {
	if to == "" {
		to = "HEAD"
	}
	rangeSpec := to
	if from != "" {
		rangeSpec = from + ".." + to
	}
	cmd := exec.Command("git", "log", "--no-merges", "--format=%s%x1f%b%x1e", rangeSpec)
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read Git history (%s): %w", rangeSpec, err)
	}

	entries := parse(output)
	if len(entries) == 0 {
		return "", fmt.Errorf("no commits found in %s", rangeSpec)
	}
	return Format(entries, from, to), nil
}

func parse(output []byte) []Entry {
	var entries []Entry
	for _, record := range bytes.Split(output, []byte{0x1e}) {
		fields := bytes.SplitN(bytes.TrimSpace(record), []byte{0x1f}, 2)
		if len(fields) == 0 || len(fields[0]) == 0 {
			continue
		}
		subject := strings.TrimSpace(string(fields[0]))
		body := ""
		if len(fields) == 2 {
			body = strings.TrimSpace(string(fields[1]))
		}
		match := conventional.FindStringSubmatch(subject)
		entry := Entry{Type: "other", Subject: subject, Details: body}
		if match != nil {
			entry.Type = match[1]
			entry.Breaking = match[2] == "!" || strings.Contains(body, "BREAKING CHANGE:")
			entry.Subject = match[4]
			if match[3] != "" {
				entry.Subject = match[3] + ": " + entry.Subject
			}
		}
		if strings.Contains(body, "BREAKING CHANGE:") {
			entry.Breaking = true
		}
		entries = append(entries, entry)
	}
	return entries
}

func Format(entries []Entry, from, to string) string {
	groups := make(map[string][]Entry)
	var types []string
	for _, entry := range entries {
		if _, ok := groups[entry.Type]; !ok {
			types = append(types, entry.Type)
		}
		groups[entry.Type] = append(groups[entry.Type], entry)
	}
	sort.SliceStable(types, func(i, j int) bool {
		order := []string{"feat", "fix", "perf", "refactor", "docs", "test", "build", "ci", "chore", "other"}
		rank := func(value string) int {
			for i, item := range order {
				if item == value { return i }
			}
			return len(order)
		}
		return rank(types[i]) < rank(types[j])
	})

	var out strings.Builder
	title := "## Changes"
	if from != "" || to != "" {
		title += " (" + from + "..." + to + ")"
	}
	out.WriteString(title + "\n")
	for _, kind := range types {
		heading := headings[kind]
		if heading == "" { heading = "Other Changes" }
		out.WriteString("\n### " + heading + "\n\n")
		for _, entry := range groups[kind] {
			prefix := ""
			if entry.Breaking { prefix = "**BREAKING:** " }
			fmt.Fprintf(&out, "- %s%s\n", prefix, entry.Subject)
			if entry.Breaking && entry.Details != "" {
				for _, line := range strings.Split(entry.Details, "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "BREAKING CHANGE:") {
						fmt.Fprintf(&out, "  - %s\n", strings.TrimPrefix(line, "BREAKING CHANGE:"))
					}
				}
			}
		}
	}
	return strings.TrimSpace(out.String()) + "\n"
}
