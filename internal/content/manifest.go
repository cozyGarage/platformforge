package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Limits struct {
	CPUs       string `json:"cpus" yaml:"cpus"`
	Memory     string `json:"memory" yaml:"memory"`
	PIDs       int    `json:"pids" yaml:"pids"`
	Timeout    int    `json:"timeout" yaml:"timeout"`
	Network    bool   `json:"network" yaml:"network"`
	Privileged bool   `json:"privileged" yaml:"privileged"`
}

type Check struct {
	Type    string `json:"type" yaml:"type"`
	Name    string `json:"name" yaml:"name"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
	Port    int    `json:"port,omitempty" yaml:"port,omitempty"`
}

type Task struct {
	ID          string   `json:"id" yaml:"id"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description" yaml:"description"`
	Hints       []string `json:"hints" yaml:"hints"`
	Checks      []Check  `json:"checks" yaml:"checks"`
}

type Runtime struct {
	Type string `json:"type" yaml:"type"`
}

type Manifest struct {
	Version       int      `json:"version" yaml:"version"`
	ID            string   `json:"id" yaml:"id"`
	Title         string   `json:"title" yaml:"title"`
	Summary       string   `json:"summary" yaml:"summary"`
	Difficulty    string   `json:"difficulty" yaml:"difficulty"`
	EstimatedMins int      `json:"estimatedMinutes" yaml:"estimatedMinutes"`
	Prerequisites []string `json:"prerequisites" yaml:"prerequisites"`
	Image         string   `json:"image" yaml:"image"`
	Shell         string   `json:"shell" yaml:"shell"`
	Runtime       Runtime  `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Setup         []string `json:"setup" yaml:"setup"`
	Tasks         []Task   `json:"tasks" yaml:"tasks"`
	Limits        Limits   `json:"limits" yaml:"limits"`
	Attribution   string   `json:"attribution,omitempty" yaml:"attribution,omitempty"`
	Lesson        string   `json:"lesson,omitempty" yaml:"-"`
}

type Catalog struct{ root string }

func NewCatalog(root string) *Catalog { return &Catalog{root: root} }

func (c *Catalog) List() ([]Manifest, error) {
	var labs []Manifest
	err := filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "lab.yaml" {
			m, err := load(path, schemaPath(c.root))
			if err != nil {
				return err
			}
			labs = append(labs, *m)
		}
		return nil
	})
	sort.Slice(labs, func(i, j int) bool { return labs[i].ID < labs[j].ID })
	return labs, err
}

func (c *Catalog) Get(id string) (*Manifest, error) {
	if id == "" || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return nil, errors.New("invalid lab id")
	}
	var match string
	err := filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != "lab.yaml" {
			return nil
		}
		m, err := load(path, schemaPath(c.root))
		if err != nil {
			return err
		}
		if m.ID == id {
			if match != "" {
				return fmt.Errorf("duplicate lab id %q", id)
			}
			match = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if match == "" {
		return nil, fmt.Errorf("lab %q not found", id)
	}
	m, err := load(match, schemaPath(c.root))
	if err != nil {
		return nil, err
	}
	lesson, err := os.ReadFile(filepath.Join(filepath.Dir(match), "lesson.md"))
	if err == nil {
		m.Lesson = string(lesson)
	}
	return m, nil
}

func schemaPath(contentRoot string) string {
	return filepath.Join(contentRoot, "..", "schemas", "lab.schema.json")
}

func load(path, schemaPath string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if schemaPath != "" {
		if err := validate(data, schemaPath); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if m.Version != 1 || m.ID == "" || m.Title == "" || m.Image == "" || len(m.Tasks) == 0 {
		return nil, fmt.Errorf("%s: missing required manifest fields", path)
	}
	if m.Shell == "" {
		m.Shell = "/bin/sh"
	}
	if m.Limits.Memory == "" {
		m.Limits.Memory = "256m"
	}
	if m.Limits.CPUs == "" {
		m.Limits.CPUs = "0.5"
	}
	if m.Limits.PIDs == 0 {
		m.Limits.PIDs = 128
	}
	if m.Limits.Timeout == 0 {
		m.Limits.Timeout = 3600
	}
	if m.Runtime.Type == "" {
		m.Runtime.Type = "container"
	}
	return &m, nil
}

func validate(data []byte, schemaPath string) error {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("lab.schema.json", bytes.NewReader(schemaData)); err != nil {
		return err
	}
	schema, err := compiler.Compile("lab.schema.json")
	if err != nil {
		return err
	}
	var document any
	if err := yaml.Unmarshal(data, &document); err != nil {
		return err
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		return err
	}
	return schema.Validate(document)
}
