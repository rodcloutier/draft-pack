package repo

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/getter"
)

// PackRepository represents a pack repository
type PackRepository struct {
	Config     *Entry
	ChartPaths []string
	IndexFile  *IndexFile
	Client     getter.Getter
}

// Entry represents a collection of parameters for pack repository
type Entry struct {
	Name     string `json:"name"`
	Cache    string `json:"cache"`
	URL      string `json:"url"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	CAFile   string `json:"caFile"`
}

// NewChartRepository constructs PackRepository
func NewRepository(cfg *Entry, getters getter.Providers) (*PackRepository, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid pack URL format: %s", cfg.URL)
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

// FindPackInRepoURL finds pack in pack repository pointed by repoURL
// without adding repo to repostiories
func FindPackInRepoURL(repoURL, chartName, chartVersion, certFile, keyFile, caFile string, getters getter.Providers) (string, error) {

	// Download and write the index file to a temporary location
	tempIndexFile, err := ioutil.TempFile("", "tmp-repo-file")
	if err != nil {
		return "", fmt.Errorf("cannot write index file for repository requested")
	}
	defer os.Remove(tempIndexFile.Name())

	e := Entry{
		URL:      repoURL,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}
	r, err := NewRepository(&e, getters)
	if err != nil {
		return "", err
	}
	if err := r.DownloadIndexFile(tempIndexFile.Name()); err != nil {
		return "", fmt.Errorf("Looks like %q is not a valid pack repository or cannot be reached: %s", repoURL, err)
	}

	// Read the index file for the repository to get pack information and return pack URL
	repoIndex, err := LoadIndexFile(tempIndexFile.Name())
	if err != nil {
		return "", err
	}

	errMsg := fmt.Sprintf("pack %q", chartName)
	if chartVersion != "" {
		errMsg = fmt.Sprintf("%s version %q", errMsg, chartVersion)
	}
	cv, err := repoIndex.Get(chartName, chartVersion)
	if err != nil {
		return "", fmt.Errorf("%s not found in %s repository", errMsg, repoURL)
	}

	if len(cv.URLs) == 0 {
		return "", fmt.Errorf("%s has no downloadable URLs", errMsg)
	}

	return cv.URLs[0], nil
}
