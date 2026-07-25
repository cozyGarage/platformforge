package content

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type UnlockGate struct {
	CompletedFromModule string `json:"completedFromModule,omitempty" yaml:"completedFromModule,omitempty"`
	Count               int    `json:"count,omitempty" yaml:"count,omitempty"`
}

type PathModule struct {
	ID         string      `json:"id" yaml:"id"`
	Title      string      `json:"title" yaml:"title"`
	Summary    string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Labs       []string    `json:"labs" yaml:"labs"`
	ComingSoon []string    `json:"comingSoon,omitempty" yaml:"comingSoon,omitempty"`
	Source     string      `json:"source,omitempty" yaml:"source,omitempty"`
	Unlock     *UnlockGate `json:"unlock,omitempty" yaml:"unlock,omitempty"`
}

type PathPhase struct {
	ID      string       `json:"id" yaml:"id"`
	Title   string       `json:"title" yaml:"title"`
	Summary string       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Modules []PathModule `json:"modules" yaml:"modules"`
}

type LearningPath struct {
	Version int         `json:"version" yaml:"version"`
	ID      string      `json:"id" yaml:"id"`
	Title   string      `json:"title" yaml:"title"`
	Summary string      `json:"summary" yaml:"summary"`
	Source  string      `json:"source,omitempty" yaml:"source,omitempty"`
	Phases  []PathPhase `json:"phases" yaml:"phases"`
}

type PathCatalog struct{ root string }

func NewPathCatalog(contentRoot string) *PathCatalog {
	return &PathCatalog{root: filepath.Join(contentRoot, "paths")}
}

func (p *PathCatalog) List() ([]LearningPath, error) {
	entries, err := os.ReadDir(p.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []LearningPath
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		item, err := p.load(filepath.Join(p.root, entry.Name()))
		if err != nil {
			return nil, err
		}
		paths = append(paths, *item)
	}
	return paths, nil
}

func (p *PathCatalog) Get(id string) (*LearningPath, error) {
	path := filepath.Join(p.root, id+".yaml")
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("path %q not found", id)
	}
	return p.load(path)
}

func (p *PathCatalog) load(path string) (*LearningPath, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lp LearningPath
	if err := yaml.Unmarshal(data, &lp); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if lp.ID == "" || lp.Title == "" || len(lp.Phases) == 0 {
		return nil, fmt.Errorf("%s: invalid learning path", path)
	}
	return &lp, nil
}
