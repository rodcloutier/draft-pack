package draftpath

import (
	"fmt"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

type Home struct {
	home draftpath.Home
}

func NewHome(h string) Home {
	return Home{
		home: draftpath.Home(h),
	}
}

func (h Home) String() string {
	return string(h.home)
}

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

func (h Home) Repository() string {
	return h.Path("repository")
}

// RepositoryFile returns the path to the repositories.yaml file.
func (h Home) RepositoryFile() string {
	return h.Path("repository", "repositories.yaml")
}

// Cache returns the path to the local cache.
func (h Home) Cache() string {
	return h.Path("repository", "cache")
}

// CacheIndex returns the path to an index for the given named repository.
func (h Home) CacheIndex(name string) string {
	target := fmt.Sprintf("%s-index.yaml", name)
	return h.Path("repository", "cache", target)
}
