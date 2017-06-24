package repo

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/getter"
)

// PackRepository represents a chart repository
type PackRepository struct {
	Config     *Entry
	ChartPaths []string
	IndexFile  *IndexFile
	Client     getter.Getter
}

// Entry represents a collection of parameters for chart repository
type Entry struct {
	Name     string `json:"name"`
	Cache    string `json:"cache"`
	URL      string `json:"url"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	CAFile   string `json:"caFile"`
}

// NewChartRepository constructs PackRepository
func NewPackRepository(cfg *Entry, getters getter.Providers) (*PackRepository, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid chart URL format: %s", cfg.URL)
	}

	getterConstructor, err := getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("Could not find protocol handler for: %s", u.Scheme)
	}
	client, _ := getterConstructor(cfg.URL, cfg.CertFile, cfg.KeyFile, cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("Could not construct protocol handler for: %s", u.Scheme)
	}

	return &PackRepository{
		Config:    cfg,
		IndexFile: NewIndexFile(),
		Client:    client,
	}, nil
}

// DownloadIndexFile fetches the index from a repository.
func (r *PackRepository) DownloadIndexFile(cachePath string) error {
	var indexURL string

	indexURL = strings.TrimSuffix(r.Config.URL, "/") + "/index.yaml"
	resp, err := r.Client.Get(indexURL)
	if err != nil {
		return err
	}

	index, err := ioutil.ReadAll(resp)
	if err != nil {
		return err
	}

	if _, err := loadIndex(index); err != nil {
		return err
	}

	// In Helm 2.2.0 the config.cache was accidentally switched to an absolute
	// path, which broke backward compatibility. This fixes it by prepending a
	// global cache path to relative paths.
	//
	// It is changed on DownloadIndexFile because that was the method that
	// originally carried the cache path.
	cp := r.Config.Cache
	if !filepath.IsAbs(cp) {
		cp = filepath.Join(cachePath, cp)
	}

	return ioutil.WriteFile(cp, index, 0644)
}
