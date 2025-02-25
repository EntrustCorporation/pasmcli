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


// untagSecretCmd represents the untag-secret command
var untagSecretCmd = &cobra.Command{
    Use:   "untag-secret",
    Short: "Untag Secret of specific tags",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        // revision
        revision, _ := flags.GetInt("revision")
        params["revision"] = revision

        // tags
        tagsArray, _ := flags.GetStringArray("tag")
        params["tags"] = tagsArray

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "UntagSecret")
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

            if retStatus == 204 {
                fmt.Println("\nUntag successful\n")
                os.Exit(0)
            } else {
                fmt.Println("\n" + retStr + "\n")
                // make a decision on what to exit with
                retMap := JsonStrToMap(retStr)
                if _, present := retMap["error"]; present {
                    os.Exit(3)
                } else {
                    fmt.Println("\nUnknown error\n")
                    os.Exit(100)
                }
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(untagSecretCmd)
    untagSecretCmd.Flags().StringP("boxid", "b", "",
                                "Id or name of the Box under which the " +
                                "Secret is")
    untagSecretCmd.Flags().StringP("secretid", "s", "",
                               "Id or name of the Secret to untag")
    untagSecretCmd.Flags().IntP("revision", "R", 0,
                             "Revision number of the Secret")
    untagSecretCmd.Flags().StringArrayP("tag", "t", []string{},
                                     "Tag to disassociate with the Secret. " +
                                     "This option is repeatable.")

    // mark mandatory fields as required
    untagSecretCmd.MarkFlagRequired("boxid")
    untagSecretCmd.MarkFlagRequired("secretid")
    untagSecretCmd.MarkFlagRequired("revision")
    untagSecretCmd.MarkFlagRequired("tag")
}
