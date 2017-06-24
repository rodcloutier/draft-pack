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

	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/spf13/cobra"
)

type packListCmd struct {
}

func init() {
	list := &packListCmd{}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return list.run()
		},
	}

	RootCmd.AddCommand(cmd)
}

func (p *packListCmd) run() error {

	builtins, err := pack.Builtins()
	if err != nil {
		return err
	}

	for pack := range builtins {
		fmt.Println("builtin/" + pack)
	}

	return nil
}
