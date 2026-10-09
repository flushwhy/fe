package odinpkg

import "testing"

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		input string
		owner string
		repo  string
		ok    bool
	}{
		{"https://github.com/flushwhy/ltdk.odin", "flushwhy", "ltdk.odin", true},
		{"https://github.com/owner/repo.git", "owner", "repo", true},
		{"http://github.com/owner/repo", "", "", false},
		{"https://example.com/owner/repo", "", "", false},
		{"https://github.com/owner/repo/tree/main", "", "", false},
		{"https://github.com/owner", "", "", false},
		{"", "", "", false},
	}
	for _, tt := range tests {
		owner, repo, err := parseGitHubURL(tt.input)
		if (err == nil) != tt.ok {
			t.Errorf("parseGitHubURL(%q) error = %v; want success %v", tt.input, err, tt.ok)
			continue
		}
		if err == nil && (owner != tt.owner || repo != tt.repo) {
			t.Errorf("parseGitHubURL(%q) = %q, %q; want %q, %q", tt.input, owner, repo, tt.owner, tt.repo)
		}
	}
}

func TestAlias(t *testing.T) {
	got, ok := aliases["ltdk"]
	if !ok || got != "https://github.com/flushwhy/ltdk.odin" {
		t.Fatalf("unexpected ltdk alias: %q, %v", got, ok)
	}
}
