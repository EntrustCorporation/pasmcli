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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	// external
	"github.com/spf13/cobra"
)

// createTerrafromSecretCmd represents the create-secret command
var createTerrafromSecretCmd = &cobra.Command{
	Use:   "create-terraform-secret",
	Short: "Create a Terraform cloud Secret within the specified Box",
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
		if leaseDurationSet {
			leaseParams := map[string]interface{}{}
			// lease duration
			leaseDuration, _ := flags.GetString("lease-duration")
			leaseParams["duration"] = leaseDuration
			// now update params with lease params
			params["lease"] = leaseParams
		}

		terraformToken, _ := flags.GetString("terraformToken")
		tokenType, _ := flags.GetString("tokenType")
		secretData := map[string]interface{}{}
		secretData["token"] = terraformToken
		secretData["type"] = tokenType
		params["secret_data"] = secretData

		// secret config
		secretConfig := map[string]interface{}{}
		// secret config type
		secretConfig["type"] = "Terraform_secrets"

		params["secret_config"] = secretConfig

		// set secret type now
		params["secret_type"] = "managed"

		// tags
		if (flags.Changed("tagkey") && !flags.Changed("tagvalue")) ||
			(!flags.Changed("tagkey") && flags.Changed("tagvalue")) {
			fmt.Println("Please provide both tag key & values")
			os.Exit(1)
		}

		if flags.Changed("tagkey") && flags.Changed("tagvalue") {
			tagkeyArray, _ := flags.GetStringArray("tagkey")
			tagvalueArray, _ := flags.GetStringArray("tagvalue")
			if len(tagkeyArray) != len(tagvalueArray) {
				fmt.Println("Please provide equal number of tag keys & values")
				os.Exit(1)
			}

			tagParams := map[string]interface{}{}
			for i := 0; i < len(tagvalueArray); i += 1 {
				if IsJSON(tagvalueArray[i]) {
					tagParams[tagkeyArray[i]] = JsonStrToMap(tagvalueArray[i])
				} else {
					tagParams[tagkeyArray[i]] = tagvalueArray[i]
				}
			}
			params["tags"] = tagParams
		}

		// rotation params
		rotationDurationSet := flags.Changed("rotation-duration")
		forceRotationSet := flags.Changed("rotation-force")
		rotationOnCheckinSet := flags.Changed("rotation-on-checkin")
		if rotationDurationSet || forceRotationSet || rotationOnCheckinSet {
			rotationParams := map[string]interface{}{}
			// rotation duration
			if rotationDurationSet {
				rotationDuration, _ := flags.GetString("rotation-duration")
				rotationParams["duration"] = rotationDuration
			}
			// rotation force
			if forceRotationSet {
				forceRotation, _ := flags.GetString("rotation-force")
				if forceRotation != "enable" && forceRotation != "disable" {
					fmt.Printf("\nInvalid -f, --rotation-force option %s. "+
						"Supported: enable, disable\n", forceRotation)
					os.Exit(1)	
				}
				rotationParams["force"] = forceRotation == "enable"
			}
			// rotation oncheckin
			if rotationOnCheckinSet {
				rotationOnCheckin, _ := flags.GetString("rotation-on-checkin")
				if rotationOnCheckin != "enable" && rotationOnCheckin != "disable" {
					fmt.Printf("\nInvalid -o, --rotation-on-checkin option %s. "+
						"Supported: enable, disable\n", rotationOnCheckin)
					os.Exit(1)
				}
				rotationParams["on_checkin"] = rotationOnCheckin == "enable"
			}
			// now update params with rotation params
			params["rotation"] = rotationParams
		}

		// exclusive checkout
		if flags.Changed("exclusive-checkout") {
			exclusiveCheckout, _ := flags.GetString("exclusive-checkout")
			if exclusiveCheckout != "enable" && exclusiveCheckout != "disable" {
				fmt.Printf("\nInvalid -x, --exclusive-checkout option %s. "+
					"Supported: enable, disable\n", exclusiveCheckout)
				os.Exit(1)
			}
			params["exclusive_checkout"] = exclusiveCheckout == "enable"
		}

		if flags.Changed("expires_at") {
			expiresAt, _ := flags.GetString("expires_at")
			params["expires_at"] = expiresAt
		}

		// JSONify
		jsonParams, err := json.Marshal(params)
		if err != nil {
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

			if retStr == "" && retStatus == 404 {
				fmt.Println("\nAction denied\n")
				os.Exit(5)
			}

			fmt.Println("\n" + retStr + "\n")

			// make a decision on what to exit with
			retMap := JsonStrToMap(retStr)
			if _, present := retMap["error"]; present {
				fmt.Println("\nError while mapping secret data\n")
				os.Exit(3)
			} else {
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createTerrafromSecretCmd)
	createTerrafromSecretCmd.Flags().StringP("boxid", "b", "",
		"Id or name of the Box")
	createTerrafromSecretCmd.Flags().StringP("name", "n", "",
		"Name of the terraform Secret")
	createTerrafromSecretCmd.Flags().StringP("description", "d", "",
		"Short description for the Secret")

	createTerrafromSecretCmd.Flags().StringP("terraformToken", "T", "",
		"Terraform Token")
	createTerrafromSecretCmd.Flags().StringP("tokenType", "t", "",
		"Terraform type of token")
	createTerrafromSecretCmd.Flags().StringP("lease-duration", "l", "",
		"Lease duration to enforce for the Secret. This property set "+
			"here, takes precedence over Box property")
	createTerrafromSecretCmd.Flags().StringP("rotation-duration", "r", "",
		"Duration on which Secret will be rotated. Behavior depends "+
			"on \"rotation force\". This property set "+
			"here, takes precedence over Box property."+
			"rotation-duration value to be specified in ISO 8601 format")
	createTerrafromSecretCmd.Flags().StringP("rotation-force", "f", "",
		"Force rotation of Secret. Behavior depends on "+
			"\"rotation duration\", \"rotation on checkin\". "+
			"Supports one of enable or disable. "+
			"This property set here takes precedence over Box property")
	createTerrafromSecretCmd.Flags().StringP("rotation-on-checkin", "o", "",
		"If this flag is set, Secret rotation would be attempted "+
			"on Checkin. Behavior varies depending on "+
			" \"rotation force\" status. Supports one of enable or disable. "+
			"This property set here takes precedence over Box property.")
	createTerrafromSecretCmd.Flags().StringP("exclusive-checkout", "x", "",
		"If this flag is set, all Secret checkouts "+
			"would be exclusive. Supports one of enable or disable. "+
			"This property set here takes precedence over Box property")
	createTerrafromSecretCmd.Flags().StringArrayP("tagkey", "k", []string{},
		"Tag key to associate with the Secret. "+
			"This option is repeatable.")
	createTerrafromSecretCmd.Flags().StringArrayP("tagvalue", "v", []string{},
		"Tag value to associate with the Secret. "+
			"This option is repeatable.")
	createTerrafromSecretCmd.Flags().StringP("expires_at", "e", "",
		"Expiration time in RFC 3339 format.")

	// mark mandatory fields as required
	createTerrafromSecretCmd.MarkFlagRequired("boxid")
	createTerrafromSecretCmd.MarkFlagRequired("name")

	createTerrafromSecretCmd.MarkFlagRequired("terraformToken")
	createTerrafromSecretCmd.MarkFlagRequired("tokenType")
}
