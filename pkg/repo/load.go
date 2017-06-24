package repo

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// ErrRepoOutOfDate indicates that the repository file is out of date, but
// is fixable.
var ErrRepoOutOfDate = errors.New("repository file is out of date")

// LoadRepositoriesFile takes a file at the given path and returns a RepoFile object
//
// If this returns ErrRepoOutOfDate, it also returns a recovered RepoFile that
// can be saved as a replacement to the out of date file.
func LoadRepositoriesFile(path string) (*RepoFile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := &RepoFile{}
	err = yaml.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}

	// File is either corrupt, or is from before v2.0.0-Alpha.5
	if r.APIVersion == "" {
		m := map[string]string{}
		if err = yaml.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		r := NewRepoFile()
		for k, v := range m {
			r.Add(&Entry{
				Name:  k,
				URL:   v,
				Cache: fmt.Sprintf("%s-index.yaml", k),
			})
		}
		return r, ErrRepoOutOfDate
	}

	return r, nil
}
