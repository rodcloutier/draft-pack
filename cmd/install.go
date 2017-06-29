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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
)

var installDesc = ""

type installCmd struct {
	home    draftpath.Home
	version string
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
			return ic.run(args[0])
		},
	}

	f := cmd.Flags()
	f.StringVarP(&ic.version, "version", "v", "", "specify the exact pack version to install. If this is not specified, the latest version is installed")

	RootCmd.AddCommand(cmd)
}

func (ic *installCmd) run(name string) error {

	// Find the requested pack
	// Find out if it doesn't exist
	// if it does, make sure we can continue
	// Download it
	// copy it to the packs directory
	// Questions:
	// - should we be able to read from archive?
	// - how to address where is comes from, subdirectories?

	return nil
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
			// if _, err := downloader.VerifyPack(abs, keyring); err != nil {
			//     return "", err
			// }
		}
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	// find the pack directory
	packRepo := filepath.Join(home.Packs(), name)
	if _, err := os.Stat(packRepo); err == nil {
		return filepath.Abs(packRepo), nil
	}

	return "", nil
}
