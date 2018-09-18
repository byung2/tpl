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

func newExecCommand() *cobra.Command {
	var opts tpl.TmplOpts
	createCmd := &cobra.Command{
		Use:   "exec [OPTIONS] TMPL_FILE [TMPL_FILE...]",
		Short: "Execute Go templates",
		//Long:  `Execute Go templates"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RequiresMinArgs(cmd, args, 1)
			if err != nil {
				return err
			}
			opts.TmplFiles = args
			return exec(&opts)
		},
	}
	createCmd.Flags().StringVarP(&opts.DataFilesStr, "datafile", "d", "", "Colon separated files containing data objects")
	createCmd.Flags().BoolVarP(&opts.UseEnv, "env", "e", false, "Load the environment variables into the data objects")
	createCmd.Flags().StringVarP(&opts.UseEnvFromPrefix, "env-prefix", "p", "", `Key prefix to load environment variables.
If a template key has a dot chain of the given value as a prefix,
load the corresponding environment variable into the data objects`)
	createCmd.Flags().StringVarP(&opts.DataOutFile, "export-data", "x", "", `Output file to store the data. Omit to do not store data.
The data also contains the values obtained in interactive mode`)
	createCmd.Flags().BoolVarP(&opts.FoldContext, "fold-context", "c", false, `Folds the parent context of missing keys when searching.
Only meaningful if the template file is yaml|json format`)
	createCmd.Flags().StringVarP(&opts.DataFormat, "format", "f", "yaml", "Default format for input data file without extention")
	createCmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, `Search for missing keys and input values from the stdin.
(Do not support template files including 'Actions' or 'Fuctions')`)
	createCmd.Flags().StringVarP(&opts.MissingKey, "missingkey", "m", "error", "The missingkey gotemplate option")
	createCmd.Flags().StringVarP(&opts.Output, "out", "o", "", `Output file to store processed templates. Omit to use stdout,
but if 'outdir' flag is specified, output will not be stdout`)
	createCmd.Flags().StringVarP(&opts.OutDir, "outdir", "", "", `Directory to store the processed templates.
If multiple template files are given, name of each file will be used
instead of the 'out' flag ($outdir/$TMPL_FILE_WITHOUT_TMPL_EXT)"`)
	createCmd.Flags().BoolVarP(&opts.Overwrite, "overwrite", "", false, "Overwrite file if it exists")
	createCmd.Flags().BoolVarP(&opts.ShowProcessedFile, "show-file", "s", false, "Show processed file info")
	return createCmd
}

func exec(opts *tpl.TmplOpts) error {
	tmpl, err := opts.OptsToTmpl()
	if err != nil {
		return err
	}
	err = tmpl.ExecuteFiles()
	if err != nil {
		return fmt.Errorf("failed to execute templates: %v", err)
	}
	tmpl.FillDestPath("")
	err = tmpl.WriteProcessedTmpl()
	if err != nil {
		return fmt.Errorf("failed to write processed templates: %v", err)
	}
	return nil
}
