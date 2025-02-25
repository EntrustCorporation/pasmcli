/*
 Copyright 2020-2025 Entrust Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var gAccessTokenFile string

const defaultCfgFileName = "pasmcli.cfg"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pasmcli",
	Short: "Entrust PASM Vault CLI",
	Long: `Perform PASM Vault operations

Create and manage Secrets.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initAccessToken)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is "+defaultCfgFileName+" on your home/profile directory)")
	// Access Token File (optional)
	rootCmd.PersistentFlags().StringVar(&gAccessTokenFile, loginOptionTokenFile, "",
		"Name of the File (with full path) for saving and reusing Access Token "+
			"and Server details. Login command creates this file while other commands "+
			"use this file. If a token file is not specified, default file vault_token.txt "+
			"is created in pasmcli.data/ under your home or profile directory.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(defaultCfgFileName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initAccessToken() {

	if len(os.Args) < 2 {
		return
	}

	// some commands don't need access token
	excludedCommands := [...]string{"login", "version", "help", "download-sample-setup-ssh-csv"}
	for _, cmd := range excludedCommands {
		if os.Args[1] == cmd {
			return
		}
	}

	for i := 2; i < len(os.Args); i++ {
		var param = os.Args[i]
		if strings.HasPrefix(param, "-") && len(param) > 2 && param[1] != '-' {
			fmt.Printf("\nInvalid parameter %q. Parameter names must be prefixed with --\nE.g. -%s\n\n",
				param, param)
			os.Exit(1)
		}
	}

	// load access token file here
	tokenFile, err := LoadAccessToken(gAccessTokenFile)
	if err != nil {
		fmt.Printf("\nError getting Server information from %s. %v.\nIf you are not logged in yet, log into Vault by running login command.\n\n",
			tokenFile, err)
		os.Exit(1)
	}
}
