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


// tagBoxCmd represents the tag-box command
var tagBoxCmd = &cobra.Command{
    Use:   "tag-box",
    Short: "Add/update tags to Box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // revision
        revision, _ := flags.GetInt("revision")
        params["revision"] = revision

        // tags
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

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "TagBox")
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
                fmt.Println("\nBox not found\n")
                os.Exit(5)
            }

            if retStatus == 204 {
                fmt.Println("\nTagging successful\n")
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
    rootCmd.AddCommand(tagBoxCmd)
    tagBoxCmd.Flags().StringP("boxid", "b", "",
                              "Id or name of the Box to tag")
    tagBoxCmd.Flags().IntP("revision", "R", 0,
                           "Revision number of the box")
    tagBoxCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                   "Tag key to associate with the Box. " +
                                   "This option is repeatable.")
    tagBoxCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                   "Tag value to associate with the Box. " +
                                   "This option is repeatable.")

    // mark mandatory fields as required
    tagBoxCmd.MarkFlagRequired("boxid")
    tagBoxCmd.MarkFlagRequired("revision")
    tagBoxCmd.MarkFlagRequired("tagkey")
    tagBoxCmd.MarkFlagRequired("tagvalue")
}
