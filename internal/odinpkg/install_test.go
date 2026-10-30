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
		{"https://github.com/owner/repo?archive=1", "", "", false},
		{"https://github.com/owner/repo#main", "", "", false},
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


func TestCleanArchivePath(t *testing.T) {
	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{"repo/main.odin", "repo/main.odin", true},
		{"repo/assets/", "repo/assets", true},
		{"../outside.odin", "", false},
		{"repo/../../outside.odin", "", false},
		{"/absolute.odin", "", false},
		{"repo\\\\outside.odin", "", false},
		{"repo//main.odin", "", false},
		{"repo/./main.odin", "", false},
	}
	for _, tt := range tests {
		got, err := cleanArchivePath(tt.name)
		if (err == nil) != tt.ok {
			t.Errorf("cleanArchivePath(%q) error = %v; want success %v", tt.name, err, tt.ok)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("cleanArchivePath(%q) = %q; want %q", tt.name, got, tt.want)
		}
	}
}
