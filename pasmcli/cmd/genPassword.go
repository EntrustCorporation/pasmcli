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


// genPasswordCmd represents the get-secret command
var genPasswordCmd = &cobra.Command{
    Use:   "gen-passwd",
    Short: "Generate password",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        // length
        length, _ := flags.GetInt("length")
        params["length"] = length

        // conditions
	upperCondition := flags.Changed("upper")
	lowerCondition := flags.Changed("lower")
	numsCondition := flags.Changed("nums")
	symbolsCondition := flags.Changed("symbols")
	if (upperCondition || lowerCondition || numsCondition || symbolsCondition) {
            conditionParams := map[string]interface{}{}
	    if (upperCondition) {
                upperMinLength, _ := flags.GetInt("upper")
                conditionParams["upper"] = upperMinLength
	    }
	    if (lowerCondition) {
                lowerMinLength, _ := flags.GetInt("lower")
                conditionParams["lower"] = lowerMinLength
	    }
	    if (numsCondition) {
                numMinLength, _ := flags.GetInt("nums")
                conditionParams["nums"] = numMinLength
	    }
	    if (symbolsCondition) {
                symbolMinLength, _ := flags.GetInt("symbols")
                conditionParams["symbols"] = symbolMinLength
	    }
	    params["conditions"] = conditionParams
	}

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "GeneratePassword")
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
    rootCmd.AddCommand(genPasswordCmd)
    genPasswordCmd.Flags().IntP("length", "l", 0,
                                "Length of the password")
    genPasswordCmd.Flags().IntP("upper", "u", 0,
                                "Minimum number of upper case characters in password")
    genPasswordCmd.Flags().IntP("lower", "w", 0,
                                "Minimum number of lower case characters in password")
    genPasswordCmd.Flags().IntP("symbols", "s", 0,
                                "Minimum number of special characters in password")
    genPasswordCmd.Flags().IntP("nums", "n", 0,
                                "Minimum number of digits in password")

    // mark mandatory fields as required
    genPasswordCmd.MarkFlagRequired("length")
}
