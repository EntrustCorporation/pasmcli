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

// createFileSecretCmd represents the create-file-secret command
// static secret
var createFileSecretCmd = &cobra.Command{
    Use:   "create-file-secret",
    Short: "Create a Vault File Secret within the specified box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret name
        name, _ := flags.GetString("name")
        params["name"] = name

        // description
        if flags.Changed("description") {
            description, _ := flags.GetString("description")
            params["desc"] = description
        }

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

        // tags
        if ((flags.Changed("tagkey") && !flags.Changed("tagvalue")) ||
            (!flags.Changed("tagkey") && flags.Changed("tagvalue"))) {
                fmt.Println("Please provide both tag key & values")
                os.Exit(1)
        }

        if (flags.Changed("tagkey") && flags.Changed("tagvalue")) {
            tagkeyArray, _ := flags.GetStringArray("tagkey")
            tagvalueArray, _ := flags.GetStringArray("tagvalue")
            if (len(tagkeyArray) != len(tagvalueArray)) {
                fmt.Println("Please provide equal number of tag keys & values")
                os.Exit(1)
            }

            tagParams := map[string]interface{}{}
            for i := 0; i < len(tagvalueArray); i +=1 {
                if (IsJSON(tagvalueArray[i])) {
                    tagParams[tagkeyArray[i]] = JsonStrToMap(tagvalueArray[i])
                } else {
                    tagParams[tagkeyArray[i]] = tagvalueArray[i]
                }
            }
            params["tags"] = tagParams
        }

        if flags.Changed("expires_at") {
            expiresAt, _ := flags.GetString("expires_at")
            params["expires_at"] = expiresAt
        }

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
        endpoint := GetEndPoint("", "1.0", "CreateSecret")
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
    rootCmd.AddCommand(createFileSecretCmd)
    createFileSecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    createFileSecretCmd.Flags().StringP("name", "n", "",
                                    "Name of the Secret")
    createFileSecretCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Secret")
    createFileSecretCmd.Flags().StringP("filename", "f", "",
                                    "File to store as Secret data")

    createFileSecretCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Secret. " +
                                         "This option is repeatable.")
    createFileSecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Secret. " +
                                         "This option is repeatable.")
    createFileSecretCmd.Flags().StringP("expires_at", "e", "",
                                    "Expiration time in RFC 3339 format.")

    // mark mandatory fields as required
    createFileSecretCmd.MarkFlagRequired("boxid")
    createFileSecretCmd.MarkFlagRequired("name")
    createFileSecretCmd.MarkFlagRequired("filename")
}