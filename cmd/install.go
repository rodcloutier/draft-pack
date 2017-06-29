// Copyright © 2017 Rodrigue Cloutier <rodcloutier@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/getter"

	"github.com/rodcloutier/draft-packs/pkg/downloader"
	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	. "github.com/rodcloutier/draft-packs/pkg/getter"
	"github.com/rodcloutier/draft-packs/pkg/repo"
)

var installDesc = ""

type installCmd struct {
	home     draftpath.Home
	version  string
	repoURL  string
	name     string
	verify   bool
	keyring  string
	certFile string
	keyFile  string
	caFile   string
	packPath string
}

func init() {
	ic := &installCmd{
		home: draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME")),
	}

	cmd := &cobra.Command{
		Use:   "install [PACK] [flags]",
		Short: "install a pack for usage",
		Long:  installDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing expected argument PACK name")
			}
			ic.name = args[0]

			ic.home = draftpath.NewHome(homePath())
			packPath, err := locatePackPath(ic.home, ic.repoURL, ic.name, ic.version, ic.verify, ic.keyring, ic.certFile, ic.keyFile, ic.caFile)
			if err != nil {
				return err
			}
			ic.packPath = packPath

			return ic.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&ic.version, "version", "v", "", "specify the exact pack version to install. If this is not specified, the latest version is installed")
	f.StringVar(&ic.repoURL, "repo", "", "chart repository url where to locate the requested pack")
	f.BoolVar(&ic.verify, "verify", false, "verify the package before icalling it")
	f.StringVar(&ic.keyring, "keyring", defaultKeyring(), "location of public keys used for verification")
	f.StringVar(&ic.certFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&ic.keyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&ic.caFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")

	RootCmd.AddCommand(cmd)
}

func (ic *installCmd) run() error {

	packsDir := ic.home.Packs()

	installed, err := isAlreadyInstalled(packsDir, ic.packPath)
	if err != nil {
		return err
	}
	if installed {
		fmt.Println("pack already installed")
		return nil
	}

	p, err := pack.Load(ic.packPath)
	if err != nil {
		return err
	}

	// Names will be converted to repo-pack to respect the
	// single level of directories in the pack directory
	// TODO (rod) since the packs are reade in an alphabebitcal order
	// maybe we would like to have a mechanism to order the installed
	// packs or any pack.
	name := strings.Replace(ic.name, "/", "-", -1)

	// Does it exists?
	exists, err := packExists(name, packsDir)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("pack with same name already exists")
	}

	packPath := filepath.Join(ic.home.Packs(), name)
	if _, err := os.Stat(packPath); os.IsNotExist(err) {
		os.MkdirAll(packPath, 0744)
	}

	err = p.SaveDir(packPath, true)
	if err != nil {
		return err
	}

	fmt.Printf("Pack %s installed. Happy drafting!\n", ic.name)

	return nil
}

func isAlreadyInstalled(packs, pack string) (bool, error) {

	packsDir, err := os.Stat(packs)
	if err != nil {
		return false, err
	}
	packDir, err := os.Stat(filepath.Dir(pack))
	if err != nil {
		return false, err
	}

	// Not rooted at the same place
	if !os.SameFile(packsDir, packDir) {
		return false, nil
	}

	name := filepath.Base(pack)

	return packExists(name, packs)
}

func packExists(name, packs string) (bool, error) {

	// Rooted at the same place, does it contain the actual directory
	files, err := ioutil.ReadDir(packs)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.Mode().IsDir() {
			if f.Name() == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	return os.ExpandEnv("$HOME/.gnupg/pubring.gpg")
}

func locatePackPath(home draftpath.Home, repoURL, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {

	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyFile(abs, keyring); err != nil {
				return "", err
			}
		}
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	// find the pack directory
	packRepo := filepath.Join(home.Packs(), name)
	if _, err := os.Stat(packRepo); err == nil {
		return filepath.Abs(packRepo)
	}

	providers := getter.Providers{
		{
			Schemes: []string{"http", "https"},
			New:     NewHTTPGetter,
		},
	}

	dl := downloader.Downloader{
		Home:    home,
		Out:     os.Stdout,
		Keyring: keyring,
		Getters: providers,
	}

	if verify {
		dl.Verify = downloader.VerifyAlways
	}
	if repoURL != "" {
		lname, err := repo.FindPackInRepoURL(repoURL, name, version,
			certFile, keyFile, caFile, providers)
		if err != nil {
			return "", err
		}
		name = lname
	}

	if _, err := os.Stat(home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(home.Archive(), 0744)
	}

	filename, _, err := dl.DownloadTo(name, version, home.Archive())
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, err
		}
		// debug("Fetched %s to %s\n", name, filename)
		return lname, nil
	}

	return filename, fmt.Errorf("file %q not found", name)
}
