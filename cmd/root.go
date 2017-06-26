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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-packs/cmd/repo"
)

const (
	homeEnvVar = "DRAFT_HOME"
)

var packDraft = `
This command consist of multiple subcommands to interact with packs

It can be used to list, create and package packs.
Example usage:
    $ draft packs list
`

func homePath() string {
	return os.ExpandEnv(homeEnvVar)
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "draft packs <cmd>",
	Short: "list, create, package packs",
	Long:  packDraft,
}

func init() {
	RootCmd.AddCommand(repo.RootCmd)
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
