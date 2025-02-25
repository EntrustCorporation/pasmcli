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


// createSSHKeySecretCmd represents the create-secret command
var createSSHKeySecretCmd = &cobra.Command{
    Use:   "create-ssh-key-secret",
    Short: "Create SSH key-based Secret within the specified Box",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}
        params["secret_type"] = "managed"

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret name
        name, _ := flags.GetString("name")
        params["name"] = name

	// get SSH secret values
        secretData := map[string]interface{}{}

	// host
	host, _ := flags.GetString("host")
        secretData["host"] = host

	// port
	if flags.Changed("port") {
	    port, _ := flags.GetInt("port")
            secretData["port"] = port
	} else {
	    secretData["port"] = 22
	}

	// user
	user, _ := flags.GetString("user")
        secretData["username"] = user

	// private key file
	keyFile, _ := flags.GetString("key-file")
	b64PrivateKey, err := B64File(keyFile)
	if err != nil {
            fmt.Printf("\nKey file error: %s\n", err)
            os.Exit(4)
	}
	secretData["private_key"] = b64PrivateKey

	if flags.Changed("key-pwd") {
	    keyPwd, _ := flags.GetString("key-pwd")
	    secretData["private_key_pwd"] = keyPwd
	}
        params["secret_data"] = secretData

        // secret config
        secretConfig := map[string]interface{}{}
        secretConfig["type"] = "SSH key endpoint"
	params["secret_config"] = secretConfig

        // description
        if flags.Changed("description") {
            description, _ := flags.GetString("description")
            params["desc"] = description
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
                if !(rotationForce == "enable" || rotationForce == "disable") {
                    fmt.Printf("\nInvalid option value provided for -f, --rotation-force option %s. " +
                                "Supported values: enable, disable\n", rotationForce)
                    os.Exit(1)
                }
                if (rotationForce == "enable") {
                    rotationParams["force"] = true
                }
                if (rotationForce == "disable") {
                    rotationParams["force"] = false
                }
                
            }
            // rotation oncheckin
            if (rotationOnCheckinSet) {
                rotationOnCheckin, _ := flags.GetString("rotation-on-checkin")
                if !(rotationOnCheckin == "enable" || rotationOnCheckin == "disable") {
                    fmt.Printf("\nInvalid option value provided for -o, --rotation-on-checkin option %s. " +
                               "Supported: enable, disable\n", rotationOnCheckin)
                    os.Exit(1)
                }
                if (rotationOnCheckin == "enable") {
                    rotationParams["on_checkin"] = true
                }
                if (rotationOnCheckin == "disable") {
                    rotationParams["on_checkin"] = false
                }                
            }
            // now update params with rotation params
            params["rotation"] = rotationParams
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

        if flags.Changed("expires_at") {
            expiresAt, _ := flags.GetString("expires_at")
            params["expires_at"] = expiresAt
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
    rootCmd.AddCommand(createSSHKeySecretCmd)
    createSSHKeySecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    createSSHKeySecretCmd.Flags().StringP("name", "n", "",
                                    "Name of the Secret")
    createSSHKeySecretCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Secret")

    createSSHKeySecretCmd.Flags().StringP("host", "H", "",
                                    "SSH endpoint IP/hostname")
    createSSHKeySecretCmd.Flags().IntP("port", "P", 0,
                                    "SSH port")
    createSSHKeySecretCmd.Flags().StringP("user", "U", "",
                                    "SSH username")
    createSSHKeySecretCmd.Flags().StringP("key-file", "K", "",
                                    "Private key file for SSH endpoint access")
    createSSHKeySecretCmd.Flags().StringP("key-pwd", "W", "",
                                    "password of private key, if encrypted")
    createSSHKeySecretCmd.Flags().StringP("master-boxid", "B", "",
                                    "Box id or name of master secret(optional)")
    createSSHKeySecretCmd.Flags().StringP("master-secretid", "I", "",
                                    "Master Secret id or name(optional)")

    createSSHKeySecretCmd.Flags().StringP("lease-duration", "l", "",
                                    "Lease duration to enforce for the Secret. This property set " +
                                    "here, takes precedence over Box property")
    createSSHKeySecretCmd.Flags().StringP("lease-renewable", "L", "",
                                    "(FOR FUTURE USE ONLY) Whether lease on checked out Secret " +
                                    "is renewable or not. This property set here, takes precedence " +
                                    "over Box property. " +
                                    "Supports one of enable or disable options")
    createSSHKeySecretCmd.Flags().StringP("rotation-duration", "r", "",
                                    "Duration on which Secret will be rotated. Behavior depends " +
                                    "on \"rotation force\". This property set " +
                                    "here, takes precedence over Box property." +
                                    "rotation-duration value to be specified in ISO 8601 format")
    createSSHKeySecretCmd.Flags().StringP("rotation-force", "f", "",
                                    "Force rotation of Secret. Behavior depends on " +
                                    "\"rotation duration\", \"rotation on checkin\". " +
                                    "Supports one of enable or disable. " +
                                    "This property set here takes precedence over Box property")
    createSSHKeySecretCmd.Flags().StringP("rotation-on-checkin", "o", "",
                                  "If this flag is set, Secret rotation would be attempted " +
                                  "on Checkin. Behavior varies depending on " +
                                  " \"rotation force\" status. Supports one of enable or disable. " +
                                  "This property set here takes precedence over Box property.")                                    
    createSSHKeySecretCmd.Flags().StringP("exclusive-checkout", "x", "",
                                    "If this flag is set, all Secret checkouts " +
                                    "would be exclusive. Supports one of enable or disable. " +
                                    "This property set here takes precedence over Box property")
    createSSHKeySecretCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Secret. " +
                                         "This option is repeatable.")
    createSSHKeySecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Secret. " +
                                         "This option is repeatable.")
    createSSHKeySecretCmd.Flags().StringP("expires_at", "e", "",
                                    "Expiration time in RFC 3339 format.")

    // mark mandatory fields as required
    createSSHKeySecretCmd.MarkFlagRequired("boxid")
    createSSHKeySecretCmd.MarkFlagRequired("name")
    createSSHKeySecretCmd.MarkFlagRequired("host")
    createSSHKeySecretCmd.MarkFlagRequired("user")
    createSSHKeySecretCmd.MarkFlagRequired("key-file")
}
