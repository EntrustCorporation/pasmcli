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

	"github.com/spf13/cobra"
)

var sampleCsvHeader = "host,port,user,password\n"
var sampleCsvRow = "10.1.2.3,22,test_user,P4$$w0rd\n"

var downloadSampleSetupSshCsvCmd = &cobra.Command{
	Use:   "download-sample-setup-ssh-csv",
	Short: "Download a sample CSV file for setting up SSH Proxy on a bulk of servers",
	Run:   downloadSampleSetupSshCsv,
}

func downloadSampleSetupSshCsv(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()

	if flags.Changed("output-file") {
		output_file, _ := flags.GetString("output-file")
		writeCsvToFile(output_file)
		return
	}
	fmt.Print(sampleCsvHeader)
	fmt.Print(sampleCsvRow)
}

func writeCsvToFile(output_file string) {
	file, err := os.Create(output_file)
	if err != nil {
		fmt.Printf("Error occurred while trying to create %s. %s\n", output_file, err.Error())
		os.Exit(1)
	}
	writeLineToFile(file, sampleCsvHeader)
	writeLineToFile(file, sampleCsvRow)
}

func writeLineToFile(file *os.File, line string) {
	_, err := file.WriteString(line)
	if err != nil {
		fmt.Printf("Error occurred while trying to write to %s.\n", file.Name())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(downloadSampleSetupSshCsvCmd)
	downloadSampleSetupSshCsvCmd.Flags().StringP("output-file", "o", "",
		"Name of the output file in which sample csv needs to be saved")
}
