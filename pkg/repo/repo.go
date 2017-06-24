package repo

import (
	"os"
	"time"

	"github.com/facebookgo/atomicfile"
	"github.com/ghodss/yaml"
)

// RepoFile represents the repositories.yaml file in $HELM_HOME
type RepoFile struct {
	APIVersion   string    `json:"apiVersion"`
	Generated    time.Time `json:"generated"`
	Repositories []*Entry  `json:"repositories"`
}

// NewRepoFile generates an empty repositories file.
//
// Generated and APIVersion are automatically set.
func NewRepoFile() *RepoFile {
	return &RepoFile{
		APIVersion:   APIVersionV1,
		Generated:    time.Now(),
		Repositories: []*Entry{},
	}
}

// Add adds one or more repo entries to a repo file.
func (r *RepoFile) Add(re ...*Entry) {
	r.Repositories = append(r.Repositories, re...)
}

// Has returns true if the given name is already a repository name.
func (r *RepoFile) Has(name string) bool {
	for _, rf := range r.Repositories {
		if rf.Name == name {
			return true
		}
	}
	return false
}

// Update attempts to replace one or more repo entries in a repo file. If an
// entry with the same name doesn't exist in the repo file it will add it.
func (r *RepoFile) Update(re ...*Entry) {
	for _, target := range re {
		found := false
		for j, repo := range r.Repositories {
			if repo.Name == target.Name {
				r.Repositories[j] = target
				found = true
				break
			}
		}
		if !found {
			r.Add(target)
		}
	}
}

// WriteFile writes a repositories file to the given path.
func (r *RepoFile) WriteFile(path string, perm os.FileMode) error {
	f, err := atomicfile.New(path, perm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}

	_, err = f.File.Write(data)
	if err != nil {
		return err
	}

	return f.Close()
}
