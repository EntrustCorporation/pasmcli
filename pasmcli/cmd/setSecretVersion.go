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


// setSecretVersionCmd represents the set-secret-version command
var setSecretVersionCmd = &cobra.Command{
    Use:   "set-secret-version",
    Short: "Set a specific version of Secret to current",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        // set access token
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretId, _ := flags.GetString("secretid")
        params["secret_id"] = secretId

        // secret version
        version, _ := flags.GetInt("version")
        params["version"] = version

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "SetSecretVersion")
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
                fmt.Println("\nSecret not found\n")
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
    rootCmd.AddCommand(setSecretVersionCmd)
    setSecretVersionCmd.Flags().StringP("boxid", "b", "",
                                 "Id or name of the Box where the Secret is")
    setSecretVersionCmd.Flags().StringP("secretid", "s", "",
                                 "Id or name of the Secret")
    setSecretVersionCmd.Flags().IntP("version", "v", 0,
                              "Version of the Secret to be set as current")

    // mark mandatory fields as required
    setSecretVersionCmd.MarkFlagRequired("boxid")
    setSecretVersionCmd.MarkFlagRequired("secretid")
    setSecretVersionCmd.MarkFlagRequired("version")
}
