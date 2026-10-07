package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var readingID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Question is one multiple-choice item. Answer is the zero-based index into Options.
// The key ships to the browser: readings are self-study, not assessment.
type Question struct {
	Prompt  string   `json:"prompt" yaml:"q"`
	Options []string `json:"options" yaml:"options"`
	Answer  int      `json:"answer" yaml:"answer"`
	Explain string   `json:"explain,omitempty" yaml:"explain,omitempty"`
}

// Reading is a theory unit: a lesson with an optional self-check quiz and no lab.
// It lives in a directory holding reading.yaml and lesson.md.
type Reading struct {
	Version       int        `json:"version" yaml:"version"`
	ID            string     `json:"id" yaml:"id"`
	Title         string     `json:"title" yaml:"title"`
	Summary       string     `json:"summary" yaml:"summary"`
	EstimatedMins int        `json:"estimatedMinutes" yaml:"estimatedMinutes"`
	Prerequisites []string   `json:"prerequisites" yaml:"prerequisites"`
	Source        string     `json:"source,omitempty" yaml:"source,omitempty"`
	Quiz          []Question `json:"quiz,omitempty" yaml:"quiz,omitempty"`
	Lesson        string     `json:"lesson,omitempty" yaml:"-"`
}

type ReadingCatalog struct{ root string }

func NewReadingCatalog(root string) *ReadingCatalog { return &ReadingCatalog{root: root} }

// List returns reading metadata only: no lesson body and no quiz key.
func (c *ReadingCatalog) List() ([]Reading, error) {
	out := []Reading{}
	err := filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "reading.yaml" {
			return err
		}
		r, err := loadReading(path)
		if err != nil {
			return err
		}
		r.Quiz = nil
		out = append(out, *r)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

func (c *ReadingCatalog) Get(id string) (*Reading, error) {
	if !readingID.MatchString(id) {
		return nil, errors.New("invalid reading id")
	}
	var match string
	err := filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "reading.yaml" {
			return err
		}
		r, err := loadReading(path)
		if err != nil {
			return err
		}
		if r.ID == id {
			match = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if match == "" {
		return nil, fmt.Errorf("reading %q not found", id)
	}
	r, err := loadReading(match)
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(match), "lesson.md"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", match, err)
	}
	r.Lesson = string(body)
	return r, nil
}

func loadReading(path string) (*Reading, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Reading
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if r.Version != 1 || !readingID.MatchString(r.ID) || strings.TrimSpace(r.Title) == "" || len(r.Summary) < 10 || r.EstimatedMins < 1 {
		return nil, fmt.Errorf("%s: missing or invalid reading fields", path)
	}
	for i, q := range r.Quiz {
		if strings.TrimSpace(q.Prompt) == "" || len(q.Options) < 2 || q.Answer < 0 || q.Answer >= len(q.Options) {
			return nil, fmt.Errorf("%s: quiz question %d is invalid", path, i+1)
		}
	}
	return &r, nil
}
