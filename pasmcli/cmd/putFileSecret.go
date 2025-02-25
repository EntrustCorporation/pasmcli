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
    "strings"
    // external
    "github.com/spf13/cobra"
)

// putFileSecretCmd represents the create-file-secret command
// static secret
var putFileSecretCmd = &cobra.Command{
    Use:   "put-file-secret",
    Short: "Puts new file in pre-existing File Secret within the specified Box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        // file
        filename, _ := flags.GetString("filename")
        fi, err := os.Stat(filename)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        // 5 MB
        max_filesize_bytes := int64(1000 * 1000 * 5)
        if fi.Size() > max_filesize_bytes {
            fmt.Println("File size cannot be greater than 5 MB")
            os.Exit(1)
        }
        b64file, err := B64File(filename)
        if err != nil {
            fmt.Println(err)
            os.Exit(4)
        }

        if len(b64file) == 0 {
            fmt.Println("Error. Empty file provided: %s", filename)
            os.Exit(1)
        }

        params["secret_data"] = b64file

        params["secret_subtype_info"] = map[string]interface{}{
            "type": "file",
            "info": map[string]interface{} {
                "filename": filename,
            },
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "PutSecretValue")
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
            retStr := strings.ReplaceAll(retBytes.String(), "\\", "")

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
                os.Exit(0)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(putFileSecretCmd)
    putFileSecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    putFileSecretCmd.Flags().StringP("secretid", "s", "",
                                    "Id or name of the Secret")
    putFileSecretCmd.Flags().StringP("filename", "f", "",
                                    "File to store as Secret data")

    // mark mandatory fields as required
    putFileSecretCmd.MarkFlagRequired("boxid")
    putFileSecretCmd.MarkFlagRequired("secretid")
    putFileSecretCmd.MarkFlagRequired("filename")
}
