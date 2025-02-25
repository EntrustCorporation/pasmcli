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


// listLeasesBySecretCmd represents the list-lease-by-secret command
var listLeasesBySecretCmd = &cobra.Command{
    Use:   "list-leases-by-secret",
    Short: "List all Lease pertaining to a given Secret",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        // filters
        if flags.Changed("filters") {
            filters, _ := flags.GetString("filters")
            params["filters"] = filters
        }
        // max items
        if flags.Changed("max-items") {
            maxItems, _ := flags.GetInt("max-items")
            params["max_items"] = maxItems
        }
        // fields
        if flags.Changed("field") {
            fieldsArray, _ := flags.GetStringArray("field")
            params["fields"] = fieldsArray
        }
        // next token
        if flags.Changed("next-token") {
            nextToken, _ := flags.GetString("next-token")
            params["next_token"] = nextToken
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "ListLeasesBySecret")
        ret, err := DoPost(endpoint,
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
                fmt.Println("\nLeases not found\n")
                os.Exit(5)
            }

            fmt.Println("\n" + retStr + "\n")

            // make a decision on what to exit with
            retMap := JsonStrToMap(retStr)
            if _, present := retMap["error"]; present {
                os.Exit(3)
            } else {
                os.Exit(0)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(listLeasesBySecretCmd)
    listLeasesBySecretCmd.Flags().StringP("boxid", "b", "",
                                         "Id of the Box under which " +
                                         "the Secret is")
    listLeasesBySecretCmd.Flags().StringP("secretid", "s", "",
                                         "Id of the Secret, whose " +
                                         "Lease we want to list")
    listLeasesBySecretCmd.Flags().StringP("filters", "l", "",
                               "Conditional expression to list filtered " +
                               "leases")
    listLeasesBySecretCmd.Flags().IntP("max-items", "m", 0,
                            "Maximum number of items to include in " +
                            "response")
    listLeasesBySecretCmd.Flags().StringArrayP("field", "f", []string{},
                                    "Lease field to include in the response")
    listLeasesBySecretCmd.Flags().StringP("next-token", "n", "",
                               "Token from which subsequent Leases would " +
                               "be processed")

    // mark mandatory fields as required
    listLeasesBySecretCmd.MarkFlagRequired("boxid")
    listLeasesBySecretCmd.MarkFlagRequired("secretid")
}
