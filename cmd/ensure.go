// Copyright © 2018 byung2
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

	"github.com/byung2/tpl"
	"github.com/spf13/cobra"
)

func newEnsureCommand() *cobra.Command {
	var opts tpl.TmplOpts
	createCmd := &cobra.Command{
		Use:   "ensure [OPTIONS] TMPL_FILE [TMPL_FILE...]",
		Short: "Check for missing keys",
		//Long:  `Check for missing keys`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RequiresMinArgs(cmd, args, 1)
			if err != nil {
				return err
			}
			opts.TmplFiles = args
			return ensure(&opts)
		},
	}
	createCmd.Flags().StringVarP(&opts.DataFilesStr, "datafile", "d", "", "Colon separated files containing data objects")
	createCmd.Flags().BoolVarP(&opts.UseEnv, "env", "e", false, "Load the environment variables into the data objects")
	createCmd.Flags().StringVarP(&opts.UseEnvFromPrefix, "env-prefix", "p", "", `Key prefix to load environment variables.
If a template key has a dot chain of the given value as a prefix,
load the corresponding environment variable into the data objects`)
	createCmd.Flags().StringVarP(&opts.DataFormat, "format", "f", "yaml", "Default format for input data file without extention")
	return createCmd
}

func ensure(opts *tpl.TmplOpts) error {
	tmpl, err := opts.OptsToTmpl()
	if err != nil {
		return err
	}
	err = tmpl.EnsureFiles()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", "there is no missing key")
	return nil
}
