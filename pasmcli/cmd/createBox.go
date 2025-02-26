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


// createBoxCmd represents the create-box command
var createBoxCmd = &cobra.Command{
    Use:   "create-box",
    Short: "Create a Vault Box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box name
        name, _ := flags.GetString("name")
        params["name"] = name

        // description
        if flags.Changed("description") {
            description, _ := flags.GetString("description")
            params["description"] = description
        }

        // lease params
        leaseDurationSet := flags.Changed("lease-duration")
        leaseRenewableSet := flags.Changed("lease-renewable")
        if (leaseDurationSet || leaseRenewableSet) {
            leaseParams := map[string]interface{}{}
            // lease duration
            if (leaseDurationSet) {
                leaseDuration, _ := flags.GetString("lease-duration")
                leaseParams["duration"] = leaseDuration
            }
            // lease renewable
            if (leaseRenewableSet) {
                // remove this once we start supporting this
                fmt.Println("FOR FUTURE USE ONLY: --lease-renewable not supported yet")
                os.Exit(1)

                leaseRenewable, _ := flags.GetString("lease-renewable")
                if (leaseRenewable == "enable" || leaseRenewable == "disable") {
                    if (leaseRenewable == "enable") {
                        leaseParams["renewable"] = true
                    }
                    if (leaseRenewable == "disable") {
                        leaseParams["renewable"] = false
                    }
                } else {
                    fmt.Printf("\nInvalid -L, --lease-renewable option %s. " +
                                "Supported: enable, disable\n", leaseRenewable)
                    os.Exit(1)
                }
            }
            // now update params with lease params
            params["lease"] = leaseParams
        }

        // max secret versions
        if flags.Changed("max-secret-versions") {
            maxSecretVersions, _ := flags.GetInt("max-secret-versions")
            params["max_secret_versions"] = maxSecretVersions
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
                rotationParams["duration"] = rotationDuration
            }
            // rotation force
            if (rotationForceSet) {
                rotationForce, _ := flags.GetString("rotation-force")
                if (rotationForce == "enable" || rotationForce == "disable") {
                    if (rotationForce == "enable") {
                        rotationParams["force"] = true
                    }
                    if (rotationForce == "disable") {
                        rotationParams["force"] = false
                    }
                } else {
                    fmt.Printf("\nInvalid -f, --rotation-force option %s. " +
                                "Supported: enable, disable\n", rotationForce)
                    os.Exit(1)
                }
            }
            // rotation oncheckin
            if (rotationOnCheckinSet) {
                rotationOnCheckin, _ := flags.GetString("rotation-on-checkin")
                if (rotationOnCheckin == "enable" || rotationOnCheckin == "disable") {
                    if (rotationOnCheckin == "enable") {
                        rotationParams["on_checkin"] = true
                    }
                    if (rotationOnCheckin == "disable") {
                        rotationParams["on_checkin"] = false
                    }
                } else {
                    fmt.Printf("\nInvalid -o, --rotation-on-checkin option %s. " +
                               "Supported: enable, disable\n", rotationOnCheckin)
                    os.Exit(1)
                }
            }
            // now update params with rotation params
            params["rotation"] = rotationParams
        }

        // exclusive checkout
        if flags.Changed("exclusive-checkout") {
            exclusiveCheckout, _ := flags.GetString("exclusive-checkout")
            if (exclusiveCheckout == "enable" || exclusiveCheckout == "disable") {
                if (exclusiveCheckout == "enable") {
                    params["exclusive_checkout"] = true
                }
                if (exclusiveCheckout == "disable") {
                    params["exclusive_checkout"] = false
                }
            } else {
                fmt.Printf("\nInvalid -x, --exclusive-checkout option %s. " +
                            "Supported: enable, disable\n", exclusiveCheckout)
                os.Exit(1)
            }
        }

        // secret duration
        if flags.Changed("secret-duration") {
            secretDuration, _ := flags.GetString("secret-duration")
            params["secret_duration"] = secretDuration
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "CreateBox")
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
    rootCmd.AddCommand(createBoxCmd)
    createBoxCmd.Flags().StringP("name", "n", "",
                                 "Name of the Box")
    createBoxCmd.Flags().StringP("description", "d", "",
                                 "Short description for the Box")
    createBoxCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                      "Tag key to associate with the Box. " +
                                      "This option is repeatable.")
    createBoxCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                      "Tag value to associate with the Box. " +
                                      "This option is repeatable.")
    createBoxCmd.Flags().IntP("max-secret-versions", "m", 0,
                              "Maximum number of Secret versions to persist")
    createBoxCmd.Flags().StringP("lease-duration", "l", "",
                                 "Lease duration to enforce for all " +
                                 "checked out Secrets within the Box. This property " +
                                 "if set in a Secret takes precedence over Box property.")
    createBoxCmd.Flags().StringP("lease-renewable", "L", "",
                               "(FOR FUTURE USE ONLY) Whether lease on checked out Secrets " +
                               "within the Box is renewable or not. " +
                               "Supports one of enable or disable. This property " +
                               "if set in a Secret takes precedence over Box property")
    createBoxCmd.Flags().StringP("rotation-duration", "r", "",
                                 "Duration on which Secrets in the Box " +
                                 "will be rotated. Behavior depends on \"rotation-force\". " +
                                 "This property if set in a Secret takes precedence over Box property")
    createBoxCmd.Flags().StringP("rotation-on-checkin", "o", "",
                                 "Secret will be rotated on checkin. " +
                                 "\"rotation-force\" flag determines the rotation behavior if the " +
                                 "checkout lease expires. The parameter value can be either " +
                                 "\"enable\" or \"disable\".")
    createBoxCmd.Flags().StringP("rotation-force", "f", "",
                               "If this flag is set, forces rotation " +
                               "of Secrets in the Box. Behavior varies for \"rotation duration\" " +
                               "and \"rotation on checkin\". Supports one of enable or disable. " +
                               "This property if set in a Secret takes precedence over Box property")
    createBoxCmd.Flags().StringP("exclusive-checkout", "x", "",
                               "If this flag is set, all Secret checkouts " +
                               "would be exclusive. Supports one of enable or disable. " +
                               "This property if set in a Secret takes precedence over Box property")
    createBoxCmd.Flags().StringP("secret-duration", "D", "",
                                 "Expiration duration for Secrets in the Box, if --expires-at option" +
                                 " not provided while doing create-secret")

    // mark mandatory fields as required
    createBoxCmd.MarkFlagRequired("name")
}
