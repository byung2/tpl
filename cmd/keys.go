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

func newKeysCommand() *cobra.Command {
	var opts tpl.TmplOpts
	createCmd := &cobra.Command{
		Use:   "keys [OPTIONS] TMPL_FILE [TMPL_FILE...]",
		Short: "Show all missing keys and processed key:value pairs",
		Long: `Show all missing keys and processed key:value pairs.
(Do not support template files including 'Actions' or 'Fuctions')`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RequiresMinArgs(cmd, args, 1)
			if err != nil {
				return err
			}
			opts.TmplFiles = args
			return keys(&opts)
		},
	}
	createCmd.Flags().StringVarP(&opts.DataFilesStr, "datafile", "d", "", `Colon separated files containing data objects
to execute templates to retrieve processed key:value pairs.
Omit to get only the keys of unprocessed TMPL FILES`)
	createCmd.Flags().BoolVarP(&opts.UseEnv, "env", "e", false, "Load the environment variables into the data objects")
	createCmd.Flags().StringVarP(&opts.UseEnvFromPrefix, "env-prefix", "p", "", `Key prefix to load environment variables.
If a template key has a dot chain of the given value as a prefix,
load the corresponding environment variable into the data objects`)
	createCmd.Flags().StringVarP(&opts.DataFormat, "format", "f", "yaml", "Default format for input data file without extention")
	createCmd.Flags().StringVarP(&opts.DataOutFormat, "output-format", "t", "yaml", "Output format for data object")
	createCmd.Flags().BoolVarP(&opts.ShowOnlyMissingKey, "missing", "m", false, `Show only missing keys of processed template.
Only used for --datafile is specified`)
	createCmd.Flags().StringVarP(&opts.DataOutFile, "out", "o", "", "Output file to store the generated data. Omit to use stdout")
	return createCmd
}

func keys(opts *tpl.TmplOpts) error {
	tmpl, err := opts.OptsToTmpl()
	if err != nil {
		return err
	}
	keys, err := tmpl.ExtractKeys()
	if err != nil {
		return fmt.Errorf("failed to export keys: %v", err)
	}
	tmpl.WriteDataObject(keys)
	return nil
}
