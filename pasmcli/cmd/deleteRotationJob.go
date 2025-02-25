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


// deleteRotationJobCmd represents the delete-rotation-job command
var deleteRotationJobCmd = &cobra.Command{
    Use:   "delete-rotation-job",
    Short: "Delete rotation job set for the box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxId, _ := flags.GetString("boxid")
        params["box_id"] = boxId

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "DeleteRotationJob")
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
                fmt.Println("\nJob not found\n")
                os.Exit(5)
            }

            retMap := JsonStrToMap(retStr)
            if retStatus == 200 {
                fmt.Printf("\nRotation job %v deleted successfully\n\n", retMap["job_id"])
                os.Exit(0)
            } else {
                fmt.Println("\n" + retStr + "\n")

                // make a decision on what to exit with
                if _, present := retMap["error"]; present {
                    os.Exit(3)
                } else {
                    os.Exit(100)
                }
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(deleteRotationJobCmd)
    deleteRotationJobCmd.Flags().StringP("boxid", "b", "",
                               "Id or name of the Box to be deleted")

    // mark mandatory fields as required
    deleteRotationJobCmd.MarkFlagRequired("boxid")
}
