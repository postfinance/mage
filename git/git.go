package git

import (
	"bytes"
	"html/template"
	"time"

	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

const (
	dfltTemplate   = `{{if ne .Tag.Name ""}}{{.Tag.Name}}-{{.Tag.Count}}-{{end}}{{.Commit.ShortHash}}{{if .Dirty}}-dirty{{end}}`
	pkgTemplate    = `{{if ne .Tag.Name ""}}{{.Tag.Name}}-{{.Tag.Count}}{{else}}{{.Tag.Count}}{{end}}{{if .Dirty}}-dirty{{end}}`
	semVerTemplate = `{{if ne .Tag.Name ""}}{{.Tag.Name}}.{{.Tag.Count}}{{else}}0.0.{{.Tag.Count}}{{end}}`
)

// Info holds relevant git information for HEAD.
type Info struct {
	Commit Commit
	Dirty  bool
	Tag    Tag
	tmpl   string
	s      string
	dirty  bool
}

// Commit contains all revision information.
type Commit struct {
	Hash      string
	ShortHash string
	Timestamp time.Time
	Author    author
}

type author struct {
	Name  string
	Email string
}

// Tag information about the newest tag. Count represents
// the number of commits since last tag. If there is no tag,
// Count is the total number of commits.
type Tag struct {
	Name  string
	Count int
}

// String renders the template. The default template creates the same output as
//     git describe --always --long --tags --dirty.
//
// This behaviour can be changed by using a different template.
func (i Info) String() string {
	return i.s
}

// New gets HEAD Info from git repo in path.
func New(path string, opts ...func(*Info)) (*Info, error) {
	i := Info{
		tmpl: dfltTemplate,
	}
	for _, opt := range opts {
		opt(&i)
	}
	r, err := gogit.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}
	cIter, err := r.Log(&gogit.LogOptions{
		From:  ref.Hash(),
		Order: gogit.LogOrderCommitterTime,
	})

	m, err := getTagMap(r)
	if err != nil {
		return nil, err
	}

	// Search the tag
	var tag string
	var count int
	var auth *object.Signature
	err = cIter.ForEach(func(c *object.Commit) error {
		if auth == nil {
			auth = &c.Author
		}
		if t, ok := m[c.Hash.String()]; ok {
			tag = t
		}
		if len(tag) != 0 {
			return storer.ErrStop
		}
		count++
		return nil
	})

	// perform dirty check
	d := false
	if i.dirty {
		d, err = dirty(r)
		if err != nil {
			return nil, err
		}
	}
	i.Dirty = d
	i.Commit.Hash = ref.Hash().String()
	i.Commit.ShortHash = i.Commit.Hash[:8]
	i.Commit.Author.Name = auth.Name
	i.Commit.Author.Email = auth.Email
	i.Commit.Timestamp = auth.When
	i.Tag.Name = tag
	i.Tag.Count = count
	t, err := template.New("string").Parse(i.tmpl)
	if err != nil {
		return nil, err
	}
	var val bytes.Buffer
	if err := t.Execute(&val, i); err != nil {
		return nil, err
	}
	i.s = string(val.Bytes())
	return &i, nil
}

// WithPackageTemplate is a template that can be used for packages.
// It creates the following:
//     <tagname>-<count>
//
// If there is no tag name, the following string is created:
//     <count>
func WithPackageTemplate() func(*Info) {
	return func(i *Info) {
		i.tmpl = pkgTemplate
	}
}

// WithSemverTemplate creates a semver string.
// It creates the following:
//     <tagname>.<count>
//
// If there is no tag name, the following string is created:
//     0.0.<count>
func WithSemverTemplate() func(*Info) {
	return func(i *Info) {
		i.tmpl = semVerTemplate
	}
}

// WithTemplate is an option to use a custom template.
func WithTemplate(tmpl string) func(*Info) {
	return func(i *Info) {
		i.tmpl = tmpl
	}
}

// WithDirtyCheck checks if repository is dirty.
func WithDirtyCheck() func(*Info) {
	return func(i *Info) {
		i.dirty = true
	}
}

func getTagMap(r *gogit.Repository) (map[string]string, error) {
	m := make(map[string]string)
	// annotated tags
	tagsObj, err := r.TagObjects()
	if err != nil {
		return nil, err
	}

	err = tagsObj.ForEach(func(t *object.Tag) error {
		m[t.Target.String()] = t.Name
		return nil
	})
	if err != nil {
		return nil, err
	}
	// unannotated tags
	tags, err := r.Tags()
	if err != nil {
		return nil, err
	}
	err = tags.ForEach(func(t *plumbing.Reference) error {
		m[t.Hash().String()] = t.Name().Short()
		return nil
	})

	return m, nil
}

func dirty(r *gogit.Repository) (bool, error) {
	w, err := r.Worktree()
	if err != nil {
		return false, err
	}
	s, err := w.Status()
	if err != nil {
		return false, err
	}

	for _, status := range s {
		if status.Staging == gogit.Untracked {
			continue
		}
		if status.Staging != gogit.Unmodified {
			return true, nil
		}
		// if status.Worktree != gogit.Unmodified && status.Worktree != gogit.Untracked {
		if status.Worktree != gogit.Unmodified {
			return true, nil
		}
	}
	return false, nil
}
