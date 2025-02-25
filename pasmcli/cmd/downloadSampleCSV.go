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
    // standard
    "os"
    "fmt"
    // external
    "github.com/spf13/cobra"
)


// downloadSampleCSVCmd represents the download-sample-csv command
var downloadSampleCSVCmd = &cobra.Command{
    Use:   "download-sample-csv",
    Short: "Download a sample CSV file for a specific secret type",
    Run: func(cmd *cobra.Command, args []string) {
        
        flags := cmd.Flags()
        

        // secret type
        secret_type, _ := flags.GetString("secret_type")
        if secret_type == "SSH key endpoint" {
            secret_type = "SSH%20key%20endpoint"
        }
        charCheck(len(secret_type))

        // now GET
        endpoint := GetEndPoint2("", "1.0", "GetSampleCSV/?secret_type=" + secret_type)
        fname, err := DoGetDownload(endpoint, GetCACertFile(),
                      AuthTokenKV())
        if err != nil {
            fmt.Printf("\nHTTP request failed: %s\n", err)
            os.Exit(4)
        } else {
            fmt.Println("\nSuccessfully downloaded sample CSV file " +
              "as - " + fname + "\n")
        }
    },
}

func init() {
    rootCmd.AddCommand(downloadSampleCSVCmd)

    downloadSampleCSVCmd.Flags().StringP("secret_type", "t", "", 
        "Secret type corresponding to the sample CSV that the user wants " +
        "to download, which can be esxi, static, or SSH key endpoint")
        // mark mandatory fields as required
    downloadSampleCSVCmd.MarkFlagRequired("secret_type")
}
