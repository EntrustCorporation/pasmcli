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


// rotateSecretCmd represents the rotate-secret command
var rotateSecretCmd = &cobra.Command{
    Use:   "rotate-secret",
    Short: "Manually rotate a managed Secret. Providing --version, -v is optional. " +
           "If provided, should match current Secret version.",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        // version
        if flags.Changed("version") {
            version, _ := flags.GetInt("version")
            params["version"] = version
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "RotateSecret")
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
            retMap := JsonStrToMap(retStr)

            if (retStr == "" && retStatus == 404) {
                fmt.Println("\nSecret not found\n")
                os.Exit(5)
            }

            if retStatus == 200 {
                fmt.Println("\nSecret rotated successfully\n")
                fmt.Println("Secret details\n--------------\n")

                // secret details returned 
                fmt.Printf("id:              %v\n", retMap["secret_id"])
                fmt.Printf("name:            %v\n", retMap["name"])
                fmt.Printf("revision:        %v\n", retMap["revision"])
                fmt.Printf("current version: %v\n\n", retMap["current_version"])
                os.Exit(0)
            } else {
                // make a decision on what to exit with
                if _, present := retMap["error"]; present {
                    fmt.Printf("\nSecret rotation failure: %v\n\n", retMap["error"])
                    os.Exit(3)
                } else {
                    fmt.Println("\n" + retStr + "\n")
                    fmt.Println("\nUnknown error\n")
                    os.Exit(100)
                }
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(rotateSecretCmd)
    rotateSecretCmd.Flags().StringP("boxid", "b", "",
                                "Id or name of the Box under which the " +
                                "Secret is")
    rotateSecretCmd.Flags().StringP("secretid", "s", "",
                               "Id or name of the Secret to rotate")
    rotateSecretCmd.Flags().IntP("version", "v", 0,
                           "(optional) current version of the Secret")

    // mark mandatory fields as required
    rotateSecretCmd.MarkFlagRequired("boxid")
    rotateSecretCmd.MarkFlagRequired("secretid")
}
