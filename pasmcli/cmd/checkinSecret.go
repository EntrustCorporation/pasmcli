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


// checkinSecretCmd represents the checkin-secret command
var checkinSecretCmd = &cobra.Command{
    Use:   "checkin-secret [flags] - Please specify one of \"leaseid\", \"lease-file\" or secret identifiers (\"boxid\", \"secretid\", \"version\"(if provided during checkin)) options",
    Short: "Checkin Secret",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        var leaseId string
        var err error
        // lease id
        if flags.Changed("leaseid") {
            // validate command options
            if flags.Changed("lease-file") {
                fmt.Println("Cannot specify both \"leaseid\" & \"lease-file\" options")
                os.Exit(1)
            }

            if (flags.Changed("boxid") || flags.Changed("secretid") || flags.Changed("version")){
                fmt.Println("Cannot specify both \"leaseid\" & Secret identifiers(\"boxid\", " +
                            "\"secretid\", \"version\") options")
                os.Exit(1)
            }

            leaseId, _ = flags.GetString("leaseid")
        // lease from file
        } else {
            var leaseFile string
            if flags.Changed("lease-file") {
                if (flags.Changed("boxid") || flags.Changed("secretid") || flags.Changed("version")) {
                    fmt.Println("Cannot specify both \"lease-file\" & Secret identifiers" +
                                "(\"boxid\", \"secretid\", \"version\")")
                    os.Exit(1)
                }

                leaseFile, _ = flags.GetString("lease-file")
            } else {
                if (!(flags.Changed("boxid") && flags.Changed("secretid"))) {
                    if flags.Changed("boxid") {
                        fmt.Println("Please specify \"secretid\" of the Secret")
                        os.Exit(1)
                    }

                    if flags.Changed("secretid") {
                        fmt.Println("Please specify \"boxid\" of the Secret")
                        os.Exit(1)
                    }

                    // nothing is specified
                    fmt.Println("Please specify one of \"leaseid\", \"lease-file\" or secret identifiers " +
                                "(\"boxid\", \"secretid\", \"version\"(if provided during checkin)) options")
                    os.Exit(1)
                }

                boxId, _ := flags.GetString("boxid")
                secretId, _ := flags.GetString("secretid")

                // get lease file path
                versionVal := -1
                if flags.Changed("version") {
                    versionVal, _ = flags.GetInt("version")
                }
                leaseFile, err = GetLeaseFilePath(boxId,
                                                        secretId,
                                                        versionVal)
                if err != nil {
                    fmt.Println("Error getting lease file path: " + err.Error())
                    os.Exit(1)
                }
            }

            // get lease id now
            leaseId, err = GetLeaseId(leaseFile)
            if err != nil {
                fmt.Println("Error getting lease id: " + err.Error())
                os.Exit(1)
            }
        }
        params["lease_id"] = leaseId

        
        

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "CheckinSecret")
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
                fmt.Println("\nCheckin successful\n")
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
    rootCmd.AddCommand(checkinSecretCmd)
    checkinSecretCmd.Flags().StringP("leaseid", "l", "",
                                 "Lease Id with which to Checkin the Secret")
    checkinSecretCmd.Flags().StringP("lease-file", "f", "",
                                 "Lease Id file to checkin the secret with")
    checkinSecretCmd.Flags().StringP("boxid", "b", "",
                                 "Box id or name of the Secret we are checking in")
    checkinSecretCmd.Flags().StringP("secretid", "s", "",
                                 "Secret id or name of the secret we are checking in")
    checkinSecretCmd.Flags().IntP("version", "v", 0,
                                 "Version of the secret we are checking in")
}
