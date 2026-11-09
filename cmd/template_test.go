package cmd

import "testing"

func TestTemplateListRejectsArguments(t *testing.T) {
	if err := templateListCmd.Args(templateListCmd, []string{"unexpected"}); err == nil {
		t.Fatal("template list accepted an unexpected argument")
	}
	if err := templateListCmd.Args(templateListCmd, nil); err != nil {
		t.Fatalf("template list rejected an empty argument list: %v", err)
	}
}

func TestTemplateUpdateRequiresOneTemplateID(t *testing.T) {
	for _, args := range [][]string{nil, {"one", "two"}} {
		if err := templateUpdateCmd.Args(templateUpdateCmd, args); err == nil {
			t.Fatalf("template update accepted arguments %q", args)
		}
	}
	if err := templateUpdateCmd.Args(templateUpdateCmd, []string{"go-basic"}); err != nil {
		t.Fatalf("template update rejected one template ID: %v", err)
	}
}
