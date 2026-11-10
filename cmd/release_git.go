package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type releaseGitState struct {
	IsRepo        bool
	Commit        string
	Branch        string
	Dirty         bool
	DirtyFiles    []string
	Unpushed      int
	UnpushedKnown bool
}

func inspectReleaseGit() (releaseGitState, error) {
	run := func(args ...string) ([]byte, error) {
		return exec.Command("git", args...).Output()
	}

	if _, err := run("rev-parse", "--show-toplevel"); err != nil {
		return releaseGitState{}, nil
	}

	state := releaseGitState{IsRepo: true}
	head, err := run("rev-parse", "HEAD")
	if err != nil {
		return state, fmt.Errorf("read Git HEAD: %w", err)
	}
	state.Commit = strings.TrimSpace(string(head))

	if branch, err := run("symbolic-ref", "--quiet", "--short", "HEAD"); err == nil {
		state.Branch = strings.TrimSpace(string(branch))
	} else {
		state.Branch = "(detached HEAD)"
	}

	if expected := strings.TrimSpace(os.Getenv("GITHUB_SHA")); expected != "" && !strings.EqualFold(expected, state.Commit) {
		return state, fmt.Errorf("checked-out commit %s does not match GITHUB_SHA %s", state.Commit, expected)
	}

	status, err := run("status", "--porcelain=v1", "--untracked-files=all", "-z")
	if err != nil {
		return state, fmt.Errorf("inspect Git working tree: %w", err)
	}
	state.DirtyFiles = parseReleaseGitStatus(status)
	state.Dirty = len(state.DirtyFiles) > 0

	upstream, err := run("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err == nil {
		count, countErr := run("rev-list", "--count", strings.TrimSpace(string(upstream))+"..HEAD")
		if countErr != nil {
			return state, fmt.Errorf("compare HEAD with its upstream: %w", countErr)
		}
		state.Unpushed, err = strconv.Atoi(strings.TrimSpace(string(count)))
		if err != nil {
			return state, fmt.Errorf("parse unpushed commit count: %w", err)
		}
		state.UnpushedKnown = true
		return state, nil
	}

	// Detached CI checkouts are common. A matching GITHUB_SHA identifies the
	// exact server-selected commit, so an upstream branch is not required.
	if expected := strings.TrimSpace(os.Getenv("GITHUB_SHA")); expected != "" && strings.EqualFold(expected, state.Commit) {
		state.UnpushedKnown = true
		return state, nil
	}

	// A local branch without an upstream can still be verified if a fetched
	// remote branch already contains its HEAD.
	remoteBranches, remoteErr := run("branch", "--remotes", "--contains", "HEAD")
	if remoteErr == nil && strings.TrimSpace(string(remoteBranches)) != "" {
		state.UnpushedKnown = true
		return state, nil
	}
	return state, nil
}

// parseReleaseGitStatus preserves Git's status codes and handles filenames containing spaces.
func parseReleaseGitStatus(status []byte) []string {
	var files []string
	for _, entry := range strings.Split(string(status), "\\x00") {
		if entry == "" {
			continue
		}
		if len(entry) >= 3 {
			files = append(files, entry[:2]+" "+entry[3:])
		} else {
			files = append(files, entry)
		}
	}
	return files
}

func validateReleaseGit(state releaseGitState, allowDirty, allowUnpushed bool) error {
	if !state.IsRepo {
		return nil
	}
	if state.Dirty && !allowDirty {
		var details strings.Builder
		fmt.Fprintln(&details, "release blocked: Git working tree has uncommitted or untracked changes:")
		for _, file := range state.DirtyFiles {
			fmt.Fprintf(&details, "  %s\\n", file)
		}
		fmt.Fprint(&details, "Commit or stash these changes, or explicitly pass --allow-dirty")
		return fmt.Errorf("%s", details.String())
	}
	if !state.UnpushedKnown && !allowUnpushed {
		return fmt.Errorf("release blocked: cannot verify that HEAD is present on a remote; configure an upstream or explicitly pass --allow-unpushed")
	}
	if state.Unpushed > 0 && !allowUnpushed {
		return fmt.Errorf("release blocked: HEAD is %d commit(s) ahead of its upstream; push the commits or explicitly pass --allow-unpushed", state.Unpushed)
	}
	return nil
}

func printReleaseGitState(state releaseGitState) {
	if !state.IsRepo {
		fmt.Println("Git: not a repository (Git checks skipped)")
		return
	}
	fmt.Printf("Git: %s (%s)\n", state.Commit, state.Branch)
	if state.Dirty {
		fmt.Println("  working tree: dirty")
		for _, file := range state.DirtyFiles {
			fmt.Printf("    %s\\n", file)
		}
	} else {
		fmt.Println("  working tree: clean")
	}
	if !state.UnpushedKnown {
		fmt.Println("  remote status: unverified")
	} else if state.Unpushed > 0 {
		fmt.Printf("  remote status: %d unpushed commit(s)\n", state.Unpushed)
	} else {
		fmt.Println("  remote status: commit verified")
	}
}
