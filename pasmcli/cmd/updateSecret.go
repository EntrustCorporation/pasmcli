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


// updateSecretCmd represents the update-secret command
var updateSecretCmd = &cobra.Command{
    Use:   "update-secret",
    Short: "Update a Vault Secret within the specified box",
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

        // description
        if flags.Changed("description") {
            description, _ := flags.GetString("description")
	    if description == "unset" {
		params["desc"] = nil
	    } else {
	        params["desc"] = description
	    }
        }

        // lease params
        leaseDurationSet := flags.Changed("lease-duration")
        leaseRenewableSet := flags.Changed("lease-renewable")
        if (leaseDurationSet || leaseRenewableSet) {
            leaseParams := map[string]interface{}{}
            // lease duration
            if (leaseDurationSet) {
                leaseDuration, _ := flags.GetString("lease-duration")
                if (leaseDuration == "unset") {
                    leaseParams["duration"] = nil
                } else {
                    leaseParams["duration"] = leaseDuration
                }
            }
            // lease renewable
            if (leaseRenewableSet) {
                // remove this once we start supporting this
                fmt.Println("FOR FUTURE USE ONLY: --lease-renewable not supported yet")
                os.Exit(1)

                leaseRenewable, _ := flags.GetString("lease-renewable")
                if (leaseRenewable == "enable" ||
                    leaseRenewable == "disable" ||
                    leaseRenewable == "unset") {
                    if (leaseRenewable == "enable") {
                        leaseParams["renewable"] = true
                    }
                    if (leaseRenewable == "disable") {
                        leaseParams["renewable"] = false
                    }
                    if (leaseRenewable == "unset") {
                        leaseParams["renewable"] = nil
                    }
                } else {
                    fmt.Printf("\nInvalid -L, --lease-renewable option %s. " +
                                "Supported: enable, disable, unset\n", leaseRenewable)
                    os.Exit(1)
                }
            }
            // now update params with lease params
            params["lease"] = leaseParams
        }

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

        // rotation params
        rotationDurationSet := flags.Changed("rotation-duration")
        rotationForceSet := flags.Changed("rotation-force")
        rotationOnCheckinSet := flags.Changed("rotation-on-checkin")
        if (rotationDurationSet || rotationForceSet || rotationOnCheckinSet) {
            rotationParams := map[string]interface{}{}
            // rotation duration
            if (rotationDurationSet) {
                rotationDuration, _ := flags.GetString("rotation-duration")
                if (rotationDuration == "unset") {
                    rotationParams["duration"] = nil
                } else {
                    rotationParams["duration"] = rotationDuration
                }
            }
            // rotation force
            if (rotationForceSet) {
                rotationForce, _ := flags.GetString("rotation-force")
                if (rotationForce == "enable" ||
                    rotationForce == "disable" ||
                    rotationForce == "unset") {
                    if (rotationForce == "enable") {
                        rotationParams["force"] = true
                    }
                    if (rotationForce == "disable") {
                        rotationParams["force"] = false
                    }
                    if (rotationForce == "unset") {
                        rotationParams["force"] = nil
                    }
                } else {
                    fmt.Printf("\nInvalid -f, --rotation-force option %s. " +
                               "Supported: enable, disable, unset\n", rotationForce)
                    os.Exit(1)
                }
            }
            // rotation oncheckin
            if (rotationOnCheckinSet) {
                rotationOnCheckin, _ := flags.GetString("rotation-on-checkin")
                if (rotationOnCheckin == "enable" ||
                    rotationOnCheckin == "disable" ||
                    rotationOnCheckin == "unset") {
                    if (rotationOnCheckin == "enable") {
                        rotationParams["on_checkin"] = true
                    }
                    if (rotationOnCheckin == "disable") {
                        rotationParams["on_checkin"] = false
                    }
                    if (rotationOnCheckin == "unset") {
                        rotationParams["on_checkin"] = nil
                    }
                } else {
                    fmt.Printf("\nInvalid -o, --rotation-on-checkin option %s. " +
                               "Supported: enable, disable, unset\n", rotationOnCheckin)
                    os.Exit(1)
                }
            }
            // now update params with rotation params
            params["rotation"] = rotationParams
        }

        // exclusive checkout
        if flags.Changed("exclusive-checkout") {
            exclusiveCheckout, _ := flags.GetString("exclusive-checkout")
            if (exclusiveCheckout == "enable" ||
                exclusiveCheckout == "disable" ||
                exclusiveCheckout == "unset") {
                if (exclusiveCheckout == "enable") {
                    params["exclusive_checkout"] = true
                }
                if (exclusiveCheckout == "disable") {
                    params["exclusive_checkout"] = false
                }
                if (exclusiveCheckout == "unset") {
                    params["exclusive_checkout"] = nil
                }
            } else {
                fmt.Printf("\nInvalid -x, --exclusive-checkout option %s. " +
                           "Supported: enable, disable, unset\n", exclusiveCheckout)
                os.Exit(1)
            }
        }

        if flags.Changed("expires-at") {
            expiresAt, _ := flags.GetString("expires-at")
            if (expiresAt == "unset") {
                params["expires_at"] = nil
            } else {
                params["expires_at"] = expiresAt
            }
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "UpdateSecret")
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
    rootCmd.AddCommand(updateSecretCmd)
    updateSecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    updateSecretCmd.Flags().StringP("secretid", "s", "",
                                    "Id or name of the Secret to update")
    updateSecretCmd.Flags().IntP("revision", "R", 0,
                                 "Revision number of the Secret")

    // updateable fields
    updateSecretCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Secret. " +
                                    "\"unset\" to clear.")
    updateSecretCmd.Flags().StringP("lease-duration", "l", "",
                                    "Lease duration to enforce for the Secret. This property set " +
                                    "here, takes precedence over Box property" +
                                    "To clear this property, set it to \"unset\".")
    updateSecretCmd.Flags().StringP("lease-renewable", "L", "",
                                    "(FOR FUTURE USE ONLY) Whether lease on checked out Secret " +
                                    "is renewable or not. This property set here, takes precedence " +
                                    "over Box property. " +
                                    "Supports one of enable, disable or unset. Unset clears the property.")
    updateSecretCmd.Flags().StringP("rotation-duration", "r", "",
                                    "Duration on which Secret will be rotated. Behavior depends " +
                                    "on \"rotation force\". This property set " +
                                    "here, takes precedence over Box property. " +
                                    "To clear this property, set it to \"unset\".")
    updateSecretCmd.Flags().StringP("rotation-force", "f", "",
                                    "Force rotation of Secret. Behavior depends on " +
                                    "\"rotation duration\", \"rotation on checkin\". " +
                                    "Supports one of enable, disable or unset. " +
                                    "Setting \"unset\" clears this property" +
                                    "This property set here takes precedence over Box property")
    updateSecretCmd.Flags().StringP("rotation-on-checkin", "o", "",
                                  "If this flag is set, Secret rotation would be attempted " +
                                  "on Checkin. Behavior varies depending on " +
                                  " \"rotation force\" status. Supports one of enable, disable or unset. " +
                                  "Setting \"unset\" clears this property" +
                                  "This property set here takes precedence over Box property.")
    updateSecretCmd.Flags().StringP("exclusive-checkout", "x", "",
                                    "If this flag is set, all Secret checkouts " +
                                    "would be exclusive. Supports one of enable, disable or unset. " +
                                    "Setting \"unset\" clears this property. " +
                                    "This property set here takes precedence over Box property")
    updateSecretCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Secret. " +
                                         "This option is repeatable.")
    updateSecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Secret. " +
                                         "This option is repeatable.")
    updateSecretCmd.Flags().StringP("expires-at", "e", "",
                                    "Expiration time in RFC 3339 format. " +
                                    "Set \"unset\" to clear")

    // mark mandatory fields as required
    updateSecretCmd.MarkFlagRequired("boxid")
    updateSecretCmd.MarkFlagRequired("secretid")
    updateSecretCmd.MarkFlagRequired("revision")
}
