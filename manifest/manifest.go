package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Entry struct {
	Path      string   `json:"path"`
	Templates []string `json:"templates"`
}

type Assemble struct {
	AssetsDir   string           `json:"assetsDir"`
	TemplateDir string           `json:"templateDir"`
	OutputDir   string           `json:"outputDir"`
	Targets     map[string]Entry `json:"targets"`
}

func LoadAssembleFile(filename string) (*Assemble, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil, errors.New("assemble.json does not exist")

	}
	defer file.Close()

	var v Assemble
	if err := json.NewDecoder(file).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (a *Assemble) AllTemplates() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	templates := make([]string, 0)
	dict := make(map[string]struct{})
	for _, v := range a.Targets {
		for _, target := range v.Templates {
			if _, ok := dict[target]; !ok {
				dict[target] = struct{}{}
				fullpath := filepath.Join(wd, a.TemplateDir, target)
				templates = append(templates, fullpath)
			}
		}
	}
	return templates, nil
}
