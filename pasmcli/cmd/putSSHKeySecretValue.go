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


// putSSHKeySecretValueCmd represents the create-secret command
var putSSHKeySecretValueCmd = &cobra.Command{
    Use:   "put-ssh-key-secret-value",
    Short: "Update existing SSH key-based Secret with new Secret value",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}
        params["secret_type"] = "managed"

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret name
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

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

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "PutSecretValue")
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
                fmt.Println("\nSecret denied\n")
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
    rootCmd.AddCommand(putSSHKeySecretValueCmd)
    putSSHKeySecretValueCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    putSSHKeySecretValueCmd.Flags().StringP("secretid", "s", "",
                                    "Id or name of the Secret")

    putSSHKeySecretValueCmd.Flags().StringP("host", "H", "",
                                    "SSH endpoint IP/hostname")
    putSSHKeySecretValueCmd.Flags().IntP("port", "P", 0,
                                    "SSH port")
    putSSHKeySecretValueCmd.Flags().StringP("user", "U", "",
                                    "SSH username")
    putSSHKeySecretValueCmd.Flags().StringP("key-file", "K", "",
                                    "Private key file for SSH endpoint access")
    putSSHKeySecretValueCmd.Flags().StringP("key-pwd", "W", "",
                                    "password of private key, if encrypted")
    putSSHKeySecretValueCmd.Flags().StringP("master-boxid", "B", "",
                                    "Box id or name of master secret(optional)")
    putSSHKeySecretValueCmd.Flags().StringP("master-secretid", "I", "",
                                    "Master Secret id or name(optional)")

    // mark mandatory fields as required
    putSSHKeySecretValueCmd.MarkFlagRequired("boxid")
    putSSHKeySecretValueCmd.MarkFlagRequired("secretid")
    putSSHKeySecretValueCmd.MarkFlagRequired("host")
    putSSHKeySecretValueCmd.MarkFlagRequired("user")
    putSSHKeySecretValueCmd.MarkFlagRequired("key-file")
}
