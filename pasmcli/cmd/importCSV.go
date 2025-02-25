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
	"bytes"
	"encoding/json"
	// external
	"github.com/spf13/cobra"
)

// importCSVCmd represents import-csv
var importCSVCmd = &cobra.Command{
	Use:	"import-csv",
	Short:	"Import CSV Secrets",
	Run: func(cmd *cobra.Command, args []string) {
	    flags := cmd.Flags()
	    params := map[string]interface{}{}

	    // csv file
	    csv_file, _ := flags.GetString("csv_file")
	    charCheck(len(csv_file))
	    params["csv_file"] = csv_file

	    // secret type
	    secret_type, _ := flags.GetString("secret_type")
	    charCheck(len(secret_type))
	    params["secret_type"] = secret_type
	    
	    // JSONify
	    jsonParams, err := json.Marshal(params)
	    if (err != nil) {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	    }

	    // now POST
	    endpoint := GetEndPoint("", "1.0", "ImportCSVSecrets")
	    ret, err := DoPostFormData(endpoint,
	                           GetCACertFile(),
				   AuthTokenKV(),
				   jsonParams,
				   "application/json")
	   if err != nil {
	       fmt.Printf("\nHTTP request failed: %s\n", err)
	       os.Exit(4)
	   } else {
	       // type assertion
	       retBytes := ret["data"].(*bytes.Buffer)
	       retStatus := ret["status"].(int)
	       retStr := retBytes.String()

	       if (retStr == "" && retStatus == 404) {
		       fmt.Println("\nAction denied\n")
		   os.Exit(5)
	       }

	       fmt.Println("\n" + retStr + "\n")

	       // make a decision on what to exit with
	       retMap := JsonStrToMap(retStr)
	       if _, present := retMap["error"]; present {
		   os.Exit(3)
	       } else {
		   fmt.Println("CSV file", csv_file, "accepted. Starting import...\n")
		   os.Exit(0)
	       }
           }
    },
}

func init() {
    rootCmd.AddCommand(importCSVCmd)
    importCSVCmd.Flags().StringP("csv_file", "c", "", 
    	"CSV file that the user wants to import")
    importCSVCmd.Flags().StringP("secret_type", "t", "", 
    	"Secret type corresponding to the secrets in the CSV which can be esxi, static, or SSH key endpoint")

    // mark mandatory fields as required
    importCSVCmd.MarkFlagRequired("csv_file")
    importCSVCmd.MarkFlagRequired("secret_type")
}
