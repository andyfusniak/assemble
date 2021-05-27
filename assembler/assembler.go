package assembler

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

type Assembly struct {
	assetsDir   string
	templateDir string
	outputDir   string
	targets     []*Target
	cwd         string
	opts        *Opts
}

type Opts struct {
	DebugMode   bool
	VerboseMode bool
}

func NewAssembly(assetsDir, templateDir, outputDir string, opts *Opts) (*Assembly, error) {
	if opts == nil {
		opts = &Opts{}
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.New("failed to get current working directory")
	}

	if assetsDir == "" {
		assetsDir = "./public"
	}

	if templateDir == "" {
		templateDir = "./templates"
	}

	if outputDir == "" {
		outputDir = "./public"
	}

	if opts.DebugMode {
		log.Printf("assetsDir=%s", assetsDir)
		log.Printf("templateDir=%s", templateDir)
		log.Printf("outputDir=%s", outputDir)
	}

	return &Assembly{
		assetsDir:   assetsDir,
		templateDir: templateDir,
		outputDir:   outputDir,
		cwd:         wd,
		opts:        opts,
	}, nil
}

// NewTarget
func (a *Assembly) NewTarget(name, path string, templates []string) *Target {
	target := &Target{
		Name:      name,
		Path:      path,
		Templates: templates,
		assembly:  a,
	}
	a.targets = append(a.targets, target)
	return target
}

// Compile
func (a *Assembly) Compile() error {
	for _, target := range a.targets {
		if err := target.compile(); err != nil {
			return err
		}
	}
	return nil
}

// Targets
func (a *Assembly) Dump() {
	for _, t := range a.targets {
		fmt.Printf("name=%s path=%s templates=%v\n", t.Name, t.Path, t.tp)
	}
}

// WriteTargets writes all target files and copies the public directory to the outputDir static dir.
func (a *Assembly) WriteTargets() error {
	for _, target := range a.targets {
		if err := target.WriteToFile(); err != nil {
			return err
		}
	}
	if err := copy.Copy(a.assetsDir, filepath.Join(a.outputDir, "static")); err != nil {
		return err
	}
	return nil
}

func (a *Assembly) Routes() map[string]string {
	m := make(map[string]string)
	for _, target := range a.targets {
		m[target.Path] = target.Name
	}
	return m
}

type Target struct {
	// Name specifies the filename of the target output file relative to the outputDir.
	Name string

	// Path specifies the resource path for static hosting.
	Path string

	// Templates specifies the filenames of the templates required to compile the target.
	Templates []string

	assembly *Assembly
	tp       *template.Template
}

func (t *Target) compile() error {
	tmplResolvd := make([]string, 0, len(t.Templates))
	for _, tmplName := range t.Templates {
		tmplResolvd = append(tmplResolvd, filepath.Join(t.assembly.cwd, t.assembly.templateDir, tmplName))
	}
	t.tp = template.Must(template.New(t.Name).ParseFiles(tmplResolvd...))
	return nil
}

func (t *Target) ensureTargetDirExist() error {
	outputDir := filepath.Join(t.assembly.cwd, t.assembly.outputDir)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, 0700); err != nil {
			return fmt.Errorf("check permissions as failed to mkdir %q", outputDir)
		}
	}
	staticDir := filepath.Join(outputDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0700); err != nil {
			return fmt.Errorf("check permissions as failed to mkdir %q", staticDir)
		}
	}
	return nil
}

func (t *Target) WriteToFile() error {
	t.ensureTargetDirExist()
	if t.tp == nil {
		if err := t.compile(); err != nil {
			return err
		}
	}
	outfile := filepath.Join(t.assembly.cwd, t.assembly.outputDir, t.Name)
	out, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %+v", outfile, err)
	}
	if err := t.tp.ExecuteTemplate(out, "layout", nil); err != nil {
		return err
	}
	out.Close()
	return nil
}
