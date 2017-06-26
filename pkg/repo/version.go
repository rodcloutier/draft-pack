package repo

import (
	"time"

	"github.com/Masterminds/semver"

	"github.com/Azure/draft/pkg/draft/pack"
)

// PackVersion represents a pack entry in the IndexFile
type PackVersion struct {
	*pack.Metadata
	URLs    []string  `json:"urls"`
	Created time.Time `json:"created,omitempty"`
	Removed bool      `json:"removed,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}

// PackVersions is a list of versioned pack references.
// Implements a sorter on Version.
type PackVersions []*PackVersion

// Len returns the length.
func (c PackVersions) Len() int { return len(c) }

// Swap swaps the position of two items in the versions slice.
func (c PackVersions) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (c PackVersions) Less(a, b int) bool {
	// Failed parse pushes to the back.
	i, err := semver.NewVersion(c[a].Version)
	if err != nil {
		return true
	}
	j, err := semver.NewVersion(c[b].Version)
	if err != nil {
		return false
	}
	return i.LessThan(j)
}
