// Package scaffold creates new projects from built-in or remote templates.
package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/flushwhy/fe/internal/ui"
)

//go:embed templates
var embeddedTemplates embed.FS

type Lang string

const (
	LangOdin Lang = "odin"
	LangC    Lang = "c"
	LangCpp  Lang = "cpp"
	LangZig  Lang = "zig"
	LangGo   Lang = "go"
)

var SupportedLangs = []Lang{LangOdin, LangC, LangCpp, LangZig, LangGo}

type Options struct {
	Lang      Lang
	Name      string // project directory name
	Dir       string // base directory to create project in (default: cwd)
	Module    string // Go module path
	RemoteURL string
	GoVersion string
}

type TemplateData struct {
	Name      string
	Lang      string
	Module    string
	GoVersion string
}

// projectRoot returns the absolute path where the project will be created.
func (o Options) projectRoot() (string, error) {
	base := o.Dir
	if base == "" {
		var err error
		base, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get working directory: %w", err)
		}
	}
	return filepath.Join(base, o.Name), nil
}

func Init(opts Options) error {
	if err := validateLang(opts.Lang); err != nil {
		return err
	}
	if opts.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if opts.GoVersion == "" {
		opts.GoVersion = goVersion()
	}
	if opts.Module == "" {
		opts.Module = opts.Name
	}

	root, err := opts.projectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		return fmt.Errorf("directory %q already exists", opts.Name)
	}

	data := TemplateData{
		Name:      opts.Name,
		Lang:      string(opts.Lang),
		Module:    opts.Module,
		GoVersion: opts.GoVersion,
	}

	ui.Title("%s  Initialising %s project: %s", ui.IconRocket, opts.Lang, opts.Name)

	var src fs.FS

	err = ui.RunSpinner("Resolving templates", func() error {
		if opts.RemoteURL != "" {
			var fetchErr error
			src, fetchErr = fetchRemote(opts.RemoteURL, string(opts.Lang))
			if fetchErr != nil {
				ui.Warn("Remote fetch failed (%v) — using built-in templates", fetchErr)
				src = nil
			}
		}
		if src == nil { // Filepath breaks on windows so use path.Join instead
			sub, err := fs.Sub(embeddedTemplates, path.Join("templates", string(opts.Lang)))
			if err != nil {
				return fmt.Errorf("no built-in template for %s: %w", opts.Lang, err)
			}
			src = sub
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = ui.RunSpinner("Writing project files", func() error {
		return writeTemplate(src, root, data)
	})
	if err != nil {
		return err
	}

	err = ui.RunSpinner("Finalising project", func() error {
		return postInit(opts, root)
	})
	if err != nil {
		return err
	}

	fmt.Println()
	ui.OK("Project %q ready", opts.Name)
	printNextSteps(opts.Lang, opts.Name)
	return nil
}

func writeTemplate(src fs.FS, destRoot string, data TemplateData) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(destRoot, path)
		if d.IsDir() {
			return os.MkdirAll(dest, os.ModePerm)
		}
		content, err := fs.ReadFile(src, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}
		destPath := strings.TrimSuffix(dest, ".tmpl")
		if filepath.Base(destPath) == "gitignore" {
			destPath = filepath.Join(filepath.Dir(destPath), ".gitignore")
		}
		rendered, err := renderTemplate(string(content), data)
		if err != nil {
			return fmt.Errorf("rendering template %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}
		return os.WriteFile(destPath, []byte(rendered), 0644)
	})
}

func renderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func postInit(opts Options, root string) error {
	switch opts.Lang {
	case LangC, LangCpp:
		stub := filepath.Join(root, "compile_commands.json")
		if err := os.WriteFile(stub, []byte("[]\n"), 0644); err != nil {
			return fmt.Errorf("could not write compile_commands.json stub: %w", err)
		}
	}
	return nil
}

func printNextSteps(lang Lang, name string) {
	switch lang {
	case LangOdin:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add ols            — wire up OLS for Neovim",
			"odin run . -file      — run the project",
		)
	case LangC:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add clangd         — wire up clangd for Neovim",
			"bear -- make          — generate compile_commands.json",
		)
	case LangCpp:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add clangd         — wire up clangd for Neovim",
			"cmake -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON",
		)
	case LangZig:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add zls            — wire up ZLS for Neovim",
			"zig build run",
		)
	case LangGo:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add gopls          — wire up gopls for Neovim",
			"go run .",
		)
	}
}

func fetchRemote(baseURL, lang string) (fs.FS, error) {
	url := strings.TrimRight(baseURL, "/") + "/" + lang + ".tar.gz"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote returned %d for %s", resp.StatusCode, url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return extractTarGz(data)
}

func validateLang(lang Lang) error {
	for _, l := range SupportedLangs {
		if l == lang {
			return nil
		}
	}
	names := make([]string, len(SupportedLangs))
	for i, l := range SupportedLangs {
		names[i] = string(l)
	}
	return fmt.Errorf("unsupported language %q — supported: %s", lang, strings.Join(names, ", "))
}

func goVersion() string {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}
