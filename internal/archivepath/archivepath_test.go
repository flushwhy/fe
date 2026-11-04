package archivepath

import "testing"

func TestValidatePortableArchivePaths(t *testing.T) {
	for _, name := range []string{"main.go", "assets/player.png", "legacy_addon/__init__.py"} {
		if err := Validate(name); err != nil {
			t.Errorf("Validate(%q): %v", name, err)
		}
	}
}

func TestValidateRejectsUnsafeArchivePaths(t *testing.T) {
	for _, name := range []string{"", ".", "../escape", "a/../../escape", "/absolute", "C:/escape", `dir\file.txt`, "bad:name", "trailing.", "trailing ", "CON.txt", "nested/NUL"} {
		if err := Validate(name); err == nil {
			t.Errorf("Validate(%q) accepted unsafe path", name)
		}
	}
}
