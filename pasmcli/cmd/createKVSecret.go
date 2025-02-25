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

// createKVSecretCmd represents the create-kv-secret command
// static secret
var createKVSecretCmd = &cobra.Command{
    Use:   "create-kv-secret",
    Short: "Create a Vault Key-Value Secret within the specified box",
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

        // data
        datakeyProvided := flags.Changed("datakey")
        datavalueProvided := flags.Changed("datavalue")
        datakvProvided := datakeyProvided || datavalueProvided

        if (!datakvProvided) {
            fmt.Println("Please specify --datakey & --datavalue" +
                        " key-value pairs for static secret")
            os.Exit(1)
        }

        // get data key-values
        if ((datakeyProvided && !datavalueProvided) ||
            (!datakeyProvided && datavalueProvided)) {
                fmt.Println("Please provide both data key & values")
                os.Exit(1)
        }

        datakeyArray, _ := flags.GetStringArray("datakey")
        datavalueArray, _ := flags.GetStringArray("datavalue")
        if (len(datakeyArray) != len(datavalueArray)) {
            fmt.Println("Please provide equal number of data keys & values")
            os.Exit(1)
        }

        secretkvParams := map[string]interface{}{}
        for i := 0; i < len(datavalueArray); i +=1 {
            if (IsJSON(datavalueArray[i])) {
                secretkvParams[datakeyArray[i]] = JsonStrToMap(datavalueArray[i])
            } else {
                secretkvParams[datakeyArray[i]] = datavalueArray[i]
            }
        }
        params["secret_data"] = secretkvParams

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
            "type": "kv",
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
    rootCmd.AddCommand(createKVSecretCmd)
    createKVSecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    createKVSecretCmd.Flags().StringP("name", "n", "",
                                    "Name of the Secret")
    createKVSecretCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Secret")
    createKVSecretCmd.Flags().StringArrayP("datakey", "X", []string{},
                                         "The key to associate with the Secret data")
    createKVSecretCmd.Flags().StringArrayP("datavalue", "Y", []string{},
                                         "Value corresponding to specific Secret data")

    createKVSecretCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Secret. " +
                                         "This option is repeatable.")
    createKVSecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Secret. " +
                                         "This option is repeatable.")
    createKVSecretCmd.Flags().StringP("expires_at", "e", "",
                                    "Expiration time in RFC 3339 format.")

    // mark mandatory fields as required
    createKVSecretCmd.MarkFlagRequired("boxid")
    createKVSecretCmd.MarkFlagRequired("name")
}