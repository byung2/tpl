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
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cliVersion = "0.1.0"

var cfgFile string
var version bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "tpl",
	//Short:         "Execute Go templates",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Printf("tpl version %s\n", cliVersion)
			return
		}
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tpl)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "version")
	rootCmd.AddCommand(newExecCommand())
	rootCmd.AddCommand(newEnsureCommand())
	rootCmd.AddCommand(newKeysCommand())
	rootCmd.AddCommand(newCompletionCommand())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tpl" (without extension).
		viper.SetConfigName(".tpl")
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		//fmt.Printf("failed to read viper config file: %v\n", err)
	}
}

// RequiresMinArgs returns an error if there is not at least min args
func RequiresMinArgs(cmd *cobra.Command, args []string, min int) error {
	if len(args) >= min {
		return nil
	}
	return fmt.Errorf(
		"\"%s\" requires at least %d argument(s).\nSee '%s --help'.\n\nUsage:  %s\n\n%s",
		cmd.CommandPath(),
		min,
		cmd.CommandPath(),
		cmd.UseLine(),
		cmd.Short,
	)
}
