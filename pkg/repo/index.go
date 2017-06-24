package repo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/provenance"
	"k8s.io/helm/pkg/urlutil"

	"github.com/Azure/draft/pkg/draft/pack"
)

var indexPath = "index.yaml"

// APIVersionV1 is the v1 API version for index and repository files.
const APIVersionV1 = "v1"

var (
	// ErrNoAPIVersion indicates that an API version was not specified.
	ErrNoAPIVersion = errors.New("no API version specified")
	// ErrNoPackVersion indicates that a chart with the given version is not found.
	ErrNoPackVersion = errors.New("no pack version found")
	// ErrNoPackName indicates that a chart with the given name is not found.
	ErrNoPackName = errors.New("no pack name found")
)

// IndexFile represents the index file in a chart repository
type IndexFile struct {
	APIVersion string                  `json:"apiVersion"`
	Generated  time.Time               `json:"generated"`
	Entries    map[string]PackVersions `json:"entries"`
	PublicKeys []string                `json:"publicKeys,omitempty"`
}

// NewIndexFile initializes an index.
func NewIndexFile() *IndexFile {
	return &IndexFile{
		APIVersion: APIVersionV1,
		Generated:  time.Now(),
		Entries:    map[string]PackVersions{},
		PublicKeys: []string{},
	}
}

// LoadIndexFile takes a file at the given path and returns an IndexFile object
func LoadIndexFile(path string) (*IndexFile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadIndex(b)
}

// IndexDirectory reads a (flat) directory and generates an index.
//
// It indexes only charts that have been packaged (*.tgz).
//
// The index returned will be in an unsorted state
func IndexDirectory(dir, baseURL string) (*IndexFile, error) {
	archives, err := filepath.Glob(filepath.Join(dir, "*.tgz"))
	if err != nil {
		return nil, err
	}
	index := NewIndexFile()
	for _, arch := range archives {
		fname := filepath.Base(arch)
		// Load pack from archive
		p, err := pack.Load(arch)
		if err != nil {
			// Assume this is not a pack.
			continue
		}
		hash, err := provenance.DigestFile(arch)
		if err != nil {
			return index, err
		}
		index.Add(p.Metadata, fname, baseURL, hash)
	}
	return index, nil
}

// loadIndex loads an index file and does minimal validity checking.
//
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (*IndexFile, error) {
	i := &IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return i, err
	}
	// TODO i.SortEntries()
	if i.APIVersion == "" {
		return i, ErrNoAPIVersion
	}
	return i, nil
}

// Add adds a file to the index
// This can leave the index in an unsorted state
func (i IndexFile) Add(md *pack.Metadata, filename, baseURL, digest string) {
	u := filename
	if baseURL != "" {
		var err error
		_, file := filepath.Split(filename)
		u, err = urlutil.URLJoin(baseURL, file)
		if err != nil {
			u = filepath.Join(baseURL, file)
		}
	}
	cr := &PackVersion{
		URLs:     []string{u},
		Metadata: md,
		Digest:   digest,
		Created:  time.Now(),
	}
	if ee, ok := i.Entries[md.Name]; !ok {
		i.Entries[md.Name] = PackVersions{cr}
	} else {
		i.Entries[md.Name] = append(ee, cr)
	}
}

// SortEntries sorts the entries by version in descending order.
//
// In canonical form, the individual version records should be sorted so that
// the most recent release for every version is in the 0th slot in the
// Entries.PackVersion array. That way, tooling can predict the newest
// version without needing to parse SemVers.
func (i IndexFile) SortEntries() {
	for _, versions := range i.Entries {
		sort.Sort(sort.Reverse(versions))
	}
}

// WriteFile writes an index file to the given destination path.
//
// The mode on the file is set to 'mode'.
func (i IndexFile) WriteFile(dest string, mode os.FileMode) error {
	b, err := yaml.Marshal(i)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, b, mode)
}

// Get returns the PackVersion for the given name.
//
// If version is empty, this will return the chart with the highest version.
func (i IndexFile) Get(name, version string) (*PackVersion, error) {
	vs, ok := i.Entries[name]
	if !ok {
		return nil, ErrNoPackName
	}
	if len(vs) == 0 {
		return nil, ErrNoPackVersion
	}

	var constraint *semver.Constraints
	if len(version) == 0 {
		constraint, _ = semver.NewConstraint("*")
	} else {
		var err error
		constraint, err = semver.NewConstraint(version)
		if err != nil {
			return nil, err
		}
	}

	for _, ver := range vs {
		test, err := semver.NewVersion(ver.Version)
		if err != nil {
			continue
		}

		if constraint.Check(test) {
			return ver, nil
		}
	}
	return nil, fmt.Errorf("No chart version found for %s-%s", name, version)
}

// Has returns true if the index has an entry for a chart with the given name and exact version.
func (i IndexFile) Has(name, version string) bool {
	_, err := i.Get(name, version)
	return err == nil
}

// Merge merges the given index file into this index.
//
// This merges by name and version.
//
// If one of the entries in the given index does _not_ already exist, it is added.
// In all other cases, the existing record is preserved.
//
// This can leave the index in an unsorted state
func (i *IndexFile) Merge(f *IndexFile) {
	for _, cvs := range f.Entries {
		for _, cv := range cvs {
			if !i.Has(cv.Name, cv.Version) {
				e := i.Entries[cv.Name]
				i.Entries[cv.Name] = append(e, cv)
			}
		}
	}
}
