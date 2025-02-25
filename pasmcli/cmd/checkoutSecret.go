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
    "time"
    "bytes"
    "encoding/json"
    // external
    "github.com/spf13/cobra"
)

type lease_Info struct {
	leaseId   string
	expiresAt string
    renewable string
}

func TwelveHourTime(expiration time.Time) string {
	// convert time to 12 hour with AMPM
	hour, min, sec := expiration.Clock()
	ampm := "AM"
	if hour > 11 {
		ampm = "PM"
	}
	if hour == 0 {
		hour = 12
	} else if hour > 12 {
		hour = hour - 12
	}

	year, month, day := expiration.Date()

	return fmt.Sprintf("%s, %02d %s %d %02d:%02d:%02d %s", expiration.Weekday(),
		day, month, year, hour, min, sec, ampm)
}

func isSecretFile(secret_subtype_info map[string]interface{}) bool {
    // will not execute if type is not in the map
    if _, ok := secret_subtype_info["type"]; ok {
        return secret_subtype_info["type"] == "file"
    }
    return false
}

func getFilename(secret_subtype_info map[string]interface{}) string {
    // assume this function will only be called after
    // confirming the secret is of type file
    info := secret_subtype_info["info"].(map[string]interface{})
    return info["filename"].(string)
}

func processRespInfo(retMap map[string]interface{}, printResp bool) lease_Info {
    var lInfo lease_Info
    if secretData, secretPresent := retMap["secret_data"]; secretPresent {

        switch secretDataVal := secretData.(type) {
        // can be string
        case string:
            if (printResp) {
                secret_subtype_info := retMap["secret_subtype_info"].(map[string]interface{})
                if !isSecretFile(secret_subtype_info) {
                    fmt.Println("\nSecret data: " + secretDataVal)
		}
            }
        // or map
        case map[string]interface{}:
            secretMapStr, err := JSONMarshalIndent(secretDataVal)
            if err != nil {
                fmt.Println("Error parsing secret data\n")
                os.Exit(5)
            } else {
                if (printResp) {
                    fmt.Println("Secret data:\n")
                    fmt.Println(string(secretMapStr))
                }
            }
        // unexpected(shouldn't happen)
        default:
            fmt.Println("Invalid secret value\n")
            os.Exit(5)
        }
    } else {
        fmt.Println("\nError during Secret checkout\n")
        os.Exit(5)
    }

    // lease details(print only if printResp is True, else goes to file)
    if lease, leasePresent := retMap["lease"]; leasePresent {
        leaseMap := lease.(map[string]interface{})

        if (printResp) {
            fmt.Println("\nLease:\n")
        }

        // lease expires at
        if leaseExpiresAt, leaseExpiresPresent := leaseMap["expires_at"];  leaseExpiresPresent {
            leaseExpiresAtVal := fmt.Sprintf("%v", leaseExpiresAt)

            // convert to local time
            localTime := true
            var location *time.Location
            location, err := time.LoadLocation("Local")
            if err != nil {
                // lets print timestamp in UTC instead of failing the call
                localTime = false
            }

            var leaseExpiresAtTime time.Time
            leaseExpiresAtTime, err = time.Parse(time.RFC3339, leaseExpiresAtVal)
            // error parsing, just print out what was returned, instead of failing
            if err != nil {
                if (printResp) {
                    fmt.Println("Expires at (UTC): " + leaseExpiresAtVal)
                }
                lInfo.expiresAt = leaseExpiresAtVal
            } else {
                var ts string
                if localTime {
                    leaseExpiresAtTime := leaseExpiresAtTime.In(location)
                    //ts = leaseExpiresAtTime.Format(time.RFC1123)
                    ts = TwelveHourTime(leaseExpiresAtTime)
                } else {
                    //ts = leaseExpiresAtTime.Format(time.RFC1123)
                    ts = TwelveHourTime(leaseExpiresAtTime)
                    ts += " UTC"
                }

                // print
                if (printResp) {
                    if localTime {
                        fmt.Println("Expires at: " + ts)
                    } else {
                        fmt.Println("Expires at (UTC): " + ts)
                    }
                }
                lInfo.expiresAt = ts
            }
        }

        // lease id
        if leaseId, leaseIdPresent := leaseMap["lease_id"];  leaseIdPresent {
            leaseIdVal := fmt.Sprintf("%v", leaseId)

            if (printResp) {
                fmt.Println("Lease id: " + leaseIdVal)
            }
            lInfo.leaseId = leaseIdVal
        } else {
            // this shouldn't happen
            fmt.Println("Lease id not found.\n")
            os.Exit(5)
        }

        // lease renewable
        if leaseRenewable, leaseRenewablePresent := leaseMap["renewable"];  leaseRenewablePresent {
            leaseRenewableVal := fmt.Sprintf("%v", leaseRenewable)

            if (printResp) {
                fmt.Println("Renewable: " + leaseRenewableVal)
            }
            lInfo.renewable = leaseRenewableVal
        }
        if (printResp) {
            fmt.Println()
        }
    }
    return lInfo
}

// checkoutSecretCmd represents the checkout-secret command
var checkoutSecretCmd = &cobra.Command{
    Use:   "checkout-secret",
    Short: "Checkout Secret",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        dontSaveLease, _ := flags.GetBool("dont-save-lease")
        leaseFile, _ := flags.GetString("lease-file")
        if (leaseFile != "" && dontSaveLease) {
            fmt.Println("Cannot set both dont-save-lease and lease-file")
            os.Exit(1)
        }

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretId, _ := flags.GetString("secretid")
        params["secret_id"] = secretId

        // secret version
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
        endpoint := GetEndPoint("", "1.0", "CheckoutSecret")
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

            if retStatus != 200 {
                fmt.Println("\n" + retStr + "\n")
                os.Exit(5)
            }

            var lInfo lease_Info
            retMap := JsonStrToMap(retStr)
            secret_subtype_info := retMap["secret_subtype_info"].(map[string]interface{})

            JSONOutput, _ := flags.GetBool("json-output")
            // expected raw output, print
            if (JSONOutput) {
                fmt.Println("\n" + retStr + "\n")
            }

            if retVal, present := retMap["error"]; present {
                if (!JSONOutput) {
                    fmt.Println("\n" + retVal.(string) + "\n")
                }
                os.Exit(3)
            } else {
                if isSecretFile(secret_subtype_info) {
                    if (JSONOutput) {
                        fmt.Println("This is a file secret, secret_data above contains " +
			            "base64 of file secret. Do a base64 decode " +
				    "to get actual file content.\n")
                    }
                }

                if (!JSONOutput) {
                    lInfo = processRespInfo(retMap, true)
                } else {
                    if (!dontSaveLease) {
                        // we call processRespInfo() with printResp set to
                        // False, since "jsonOutput" was requested, and we
                        // have already printed the json raw output. However,
                        // "dontSaveLease" is False, which means, we still
                        // need "lInfo" so we save it to a file
                        lInfo = processRespInfo(retMap, false)
                    }
                }

                // save lease info to file (also check & proceed only if
                // "lease id" is present)
                if (!dontSaveLease && (lInfo.leaseId != "")) {
                    // get file path
                    if (leaseFile == "") {
                        versionVal := -1
                        if flags.Changed("version") {
                            versionVal, _ = flags.GetInt("version")
                        }
                        leaseFile, err = GetLeaseFilePath(boxid,
                                                                secretId,
                                                                versionVal)
                        if err != nil {
                            fmt.Println("Error getting lease file path: " + err.Error())
                            os.Exit(1)
                        }
                    }

                    version, _ := flags.GetInt("version")
                    leaseFile, err = SaveLeaseInfo(leaseFile,
                                                         boxid,
                                                         secretId,
                                                         lInfo.leaseId,
                                                         lInfo.expiresAt,
                                                         lInfo.renewable,
                                                         version)
                    if err != nil {
                        fmt.Printf("\nError saving lease info to %s - %v\n", leaseFile, err)
                        os.Exit(4)
                    }

                    fmt.Printf("\nLease id saved in %s. Pass this file if checking in " +
                               "the secret with --lease-file option.\n", leaseFile)
                }
                if isSecretFile(secret_subtype_info) {
                    filename := getFilename(secret_subtype_info)
                    b64content := retMap["secret_data"].(string)
                    content := B64Decode(b64content)
                    f, err := os.Create(filename)
                    if err != nil {
                        fmt.Printf("\nUnable to write to %s\n", filename)
                        os.Exit(4)
                    }
                    defer f.Close()
                    _, err2 := f.WriteString(content)
                    if err2 != nil {
                        fmt.Printf("\nUnable to write to %s\n", filename)
                        os.Exit(4)
                    }
                    fmt.Printf("\nSuccessfully downloaded %s\n", filename)
                }
                fmt.Println()
                os.Exit(0)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(checkoutSecretCmd)
    checkoutSecretCmd.Flags().StringP("boxid", "b", "",
                                 "Id or name of the Box where the Secret is")
    checkoutSecretCmd.Flags().StringP("secretid", "s", "",
                                 "Id or name of the Secret to fetch")
    checkoutSecretCmd.Flags().IntP("version", "v", 0,
                              "Version of the Secret to fetch. If not " +
                              "specified, fetch the latest version")
    checkoutSecretCmd.Flags().BoolP("json-output", "j", false,
                                 "Show JSON formatted output")
    checkoutSecretCmd.Flags().StringP("lease-file", "l", "",
                                    "File to save lease details to, on successful " +
                                    "checkout, if lease id present. If this option " +
                                    "and \"dont-save-lease\" " +
                                    "is not provided, lease details will be stored " +
                                    "by default at $HOMEDIR/vault.data/vault_lease_<id>.txt")
    checkoutSecretCmd.Flags().BoolP("dont-save-lease", "d", false,
                                    "Do not save lease info in a file")

    // mark mandatory fields as required
    checkoutSecretCmd.MarkFlagRequired("boxid")
    checkoutSecretCmd.MarkFlagRequired("secretid")
}
