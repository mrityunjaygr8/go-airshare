/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	// "github.com/mrityunjaygr8/go-airshare/functions"
	"github.com/mrityunjaygr8/go-airshare/utils"
)

var cfgFile string
var port int
var text string
var clipSend bool


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-airshare [flags] code [files] ",
	Short: "Golang port of airshare",
	Long: `A golang port of airshare, a python library for airdrop-like functionality.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if clipSend {
			textContent, err := utils.CopyClipBoard()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			text = string(textContent)
		}
		if text != "" {
			utils.CreateTextService(args[0], text, port)
		}
		if len(args) == 2 {
			utils.CreateFileService(args[0], args[1:], port)
		}
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-airshare.yaml)")

	rootCmd.Flags().IntVarP(&port ,"port", "p", utils.Default_Port, "the port where the webserver will be run")
	rootCmd.Flags().StringVarP(&text, "text", "t", "", "the text to be sent via the webserver")
	rootCmd.Flags().BoolVarP(&clipSend, "clip-send", "c", false, "send (serve) the clipboard content")
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

		// Search config in home directory with name ".go-airshare" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".go-airshare")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
