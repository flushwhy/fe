package o3de

import (
	"strings"
)

// SummarizeBuildFailure extracts actionable, distinct clues from CMake/compiler output.
// It intentionally preserves the original build output; these are extra pointers, not a replacement.
func SummarizeBuildFailure(output string) []string {
	rules := []struct {
		match  []string
		summary string
	}{
		{[]string{"cmake error", "cmake error at"}, "CMake configuration or build-system error: inspect the CMake Error lines above."},
		{[]string{"could not find", "could not locate", "not found"}, "A dependency or tool may be missing; check the named package/tool and CMake search paths."},
		{[]string{"fatal error:", "no such file or directory"}, "A required header or file is missing; check dependencies, include paths, and generated files."},
		{[]string{"undefined reference", "unresolved external symbol", "ld returned"}, "Linking failed; check missing libraries, target linkage, and architecture/configuration mismatches."},
		{[]string{"ninja: error", "make: ***", "no rule to make target"}, "The build system could not complete a target; reconfigure the build directory with CMake if it is stale."},
		{[]string{" error:", " error C", " error LNK"}, "The compiler reported an error; use the first compiler error above as the likely root cause."},
	}
	lower := strings.ToLower(output)
	var summaries []string
	seen := map[string]bool{}
	for _, rule := range rules {
		for _, token := range rule.match {
			if strings.Contains(lower, token) {
				if !seen[rule.summary] {
					summaries = append(summaries, rule.summary)
					seen[rule.summary] = true
				}
				break
			}
		}
	}
	if len(summaries) == 0 {
		summaries = append(summaries, "CMake build failed; review the final build output and rerun with --verbose for more detail.")
	}
	if strings.Contains(lower, "cmake --build") || strings.Contains(lower, "could not load cache") {
		summaries = append(summaries, "If this directory has not been configured yet, configure it with the project's documented O3DE workflow first.")
	}
	if len(summaries) > 5 {
		summaries = summaries[:5]
	}
	return summaries
}
