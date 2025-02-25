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


// createEsxiHostSecretCmd represents the create-secret command
var createEsxiHostSecretCmd = &cobra.Command{
    Use:   "create-esxi-host-secret",
    Short: "Create a ESXi Host Secret within the specified Box",
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

	ESXiHost, _ := flags.GetString("ESXihost")
	ESXiUser, _ := flags.GetString("ESXiuser")
	ESXiPasswd, _ := flags.GetString("ESXipasswd")
	secretData := map[string]interface{}{}
	secretData["host"] = ESXiHost
	secretData["uid"] = ESXiUser
	secretData["password"] = ESXiPasswd
	params["secret_data"] = secretData

	// secret config
	secretConfig := map[string]interface{}{}
	// secret config type
	secretConfig["type"] = "ESXi Host Account"

	// secret config master
	if ((flags.Changed("master-boxid") && !flags.Changed("master-secretid")) ||
	    (!flags.Changed("master-boxid") && flags.Changed("master-secretid"))) {
	    fmt.Println("Please provide both master-boxid as well as master-secretid")
	    os.Exit(1)
	}
	if (flags.Changed("master-boxid") && flags.Changed("master-secretid")) {
	    masterBoxId, _ := flags.GetString("master-boxid")
	    masterSecretId, _ := flags.GetString("master-secretid")

	    masterSecretConfig := map[string]interface{}{}
	    masterSecretConfig["box_id"] = masterBoxId
	    masterSecretConfig["secret_id"] = masterSecretId
	    secretConfig["master_secret"] = masterSecretConfig
	}
	// secret config ESXi TLS version
	if (flags.Changed("ESXi-tls-version")) {
	    ESXiTLSVersion, _ := flags.GetString("ESXi-tls-version")
	    secretConfig["tls_version"] = ESXiTLSVersion
	}
	// secret config ESXi CA Certificate
	if (flags.Changed("ESXicacert")) {
	    ESXiCACert, _ := flags.GetString("ESXicacert")
	    secretConfig["ca_cert"] = ESXiCACert
	}
	params["secret_config"] = secretConfig

	// set secret type now
	params["secret_type"] = "managed"

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
    rootCmd.AddCommand(createEsxiHostSecretCmd)
    createEsxiHostSecretCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    createEsxiHostSecretCmd.Flags().StringP("name", "n", "",
                                    "Name of the ESXi Host Secret")
    createEsxiHostSecretCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Secret")

    createEsxiHostSecretCmd.Flags().StringP("ESXihost", "H", "",
                                    "ESXi host address")
    createEsxiHostSecretCmd.Flags().StringP("ESXiuser", "U", "",
                                    "ESXi username")
    createEsxiHostSecretCmd.Flags().StringP("ESXipasswd", "P", "",
                                    "ESXi password")
    createEsxiHostSecretCmd.Flags().StringP("ESXi-tls-version", "T", "",
                                    "TLS version(optional) to use while connecting to ESXi")
    createEsxiHostSecretCmd.Flags().StringP("ESXicacert", "c", "",
                                    "CA Certificate(optional) to use while connecting to ESXi")

    createEsxiHostSecretCmd.Flags().StringP("master-boxid", "B", "",
                                    "(optional) Box id or name of master secret")
    createEsxiHostSecretCmd.Flags().StringP("master-secretid", "I", "",
                                    "(optional) Master Secret id or name")

    createEsxiHostSecretCmd.Flags().StringP("lease-duration", "l", "",
                                    "Lease duration to enforce for the Secret. This property set " +
                                    "here, takes precedence over Box property")
    createEsxiHostSecretCmd.Flags().StringP("lease-renewable", "L", "",
                                    "(FOR FUTURE USE ONLY) Whether lease on checked out Secret " +
                                    "is renewable or not. This property set here, takes precedence " +
                                    "over Box property. " +
                                    "Supports one of enable or disable options")
    createEsxiHostSecretCmd.Flags().StringP("rotation-duration", "r", "",
                                    "Duration on which Secret will be rotated. Behavior depends " +
                                    "on \"rotation force\". This property set " +
                                    "here, takes precedence over Box property." +
                                    "rotation-duration value to be specified in ISO 8601 format")
    createEsxiHostSecretCmd.Flags().StringP("rotation-force", "f", "",
                                    "Force rotation of Secret. Behavior depends on " +
                                    "\"rotation duration\", \"rotation on checkin\". " +
                                    "Supports one of enable or disable. " +
                                    "This property set here takes precedence over Box property")
    createEsxiHostSecretCmd.Flags().StringP("rotation-on-checkin", "o", "",
                                  "If this flag is set, Secret rotation would be attempted " +
                                  "on Checkin. Behavior varies depending on " +
                                  " \"rotation force\" status. Supports one of enable or disable. " +
                                  "This property set here takes precedence over Box property.")
    createEsxiHostSecretCmd.Flags().StringP("exclusive-checkout", "x", "",
                                    "If this flag is set, all Secret checkouts " +
                                    "would be exclusive. Supports one of enable or disable. " +
                                    "This property set here takes precedence over Box property")
    createEsxiHostSecretCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Secret. " +
                                         "This option is repeatable.")
    createEsxiHostSecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Secret. " +
                                         "This option is repeatable.")
    createEsxiHostSecretCmd.Flags().StringP("expires_at", "e", "",
                                    "Expiration time in RFC 3339 format.")

    // mark mandatory fields as required
    createEsxiHostSecretCmd.MarkFlagRequired("boxid")
    createEsxiHostSecretCmd.MarkFlagRequired("name")

    createEsxiHostSecretCmd.MarkFlagRequired("ESXihost")
    createEsxiHostSecretCmd.MarkFlagRequired("ESXiuser")
    createEsxiHostSecretCmd.MarkFlagRequired("ESXipasswd")
}
