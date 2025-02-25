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
    "strings"
    "encoding/json"
    // external
    "github.com/spf13/cobra"
)


// updatePolicyCmd represents the update-policy command
var updatePolicyCmd = &cobra.Command{
    Use:   "update-policy",
    Short: "Update a given Vault Policy",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        // create request payload
        policyid, _ := flags.GetString("policyid")
        params["policy_id"] = policyid

        revision, _ := flags.GetInt("revision")
        params["revision"] = revision


        if flags.Changed("description") {
            description, _ := flags.GetString("description")
            if description == "unset" {
                params["desc"] = nil
            } else {
                params["desc"] = description
            }
        }

        if (flags.Changed("local-user") || flags.Changed("ad-upn") ||
            flags.Changed("ad-logon-name") || flags.Changed("ad-group")) {
            principalParams := []interface{}{}

            if flags.Changed("local-user") {
                localUserArray, _ := flags.GetStringArray("local-user")
                for i := 0; i < len(localUserArray); i +=1 {
                    userMap := map[string]interface{}{}
                    userMap["username"] = localUserArray[i]
                    localUser := map[string]interface{}{}
                    localUser["local_user"] = userMap
                    principalParams = append(principalParams, localUser)
                }
            }

            if flags.Changed("ad-upn") {
                upnArray, _ := flags.GetStringArray("ad-upn")
                for i := 0; i < len(upnArray); i +=1 {
                    upnMap := map[string]interface{}{}
                    upnMap["upn"] = upnArray[i]
                    adUser := map[string]interface{}{}
                    adUser["ad_user"] = upnMap
                    principalParams = append(principalParams, adUser)
                }
            }

            if flags.Changed("ad-logon-name") {
                samArray, _ := flags.GetStringArray("ad-logon-name")
                for i := 0; i < len(samArray); i +=1 {
                    samMap := map[string]interface{}{}
                    samMap["logon_name"] = samArray[i]
                    adUser := map[string]interface{}{}
                    adUser["ad_user"] = samMap
                    principalParams = append(principalParams, adUser)
                }
            }

            if flags.Changed("ad-group") {
                adGroupArray, _ := flags.GetStringArray("ad-group")
                for i := 0; i < len(adGroupArray); i +=1 {
                    groupList := strings.Split(adGroupArray[i], "||")
                    // remove trailing/leading whitespace
                    for index := range groupList {
                        groupList[index] = strings.TrimSpace(groupList[index])
                    }

                    if len(groupList) != 2 {
                        fmt.Printf("\nInvalid ad-group argument: %s\n", adGroupArray[i])
                        os.Exit(1)
                    }
                    adGroupMap := map[string]interface{}{}
                    adGroupMap["dn"] = groupList[0]
                    adGroupMap["display_name"] = groupList[1]
                    adGroup := map[string]interface{}{}
                    adGroup["ad_group"] = adGroupMap
                    principalParams = append(principalParams, adGroup)
                }
            }
            params["principals"] = principalParams
        }


        if (flags.Changed("resource")) {
            resourceParams := []interface{}{}
            resourceArray, _ := flags.GetStringArray("resource")

            for i := 0; i < len(resourceArray); i +=1 {
                resource := strings.Split(resourceArray[i], ",")
                // remove trailing/leading whitespace
                for index := range resource {
                    resource[index] = strings.TrimSpace(resource[index])
                }

                if len(resource) < 2 {
                    fmt.Printf("\nInvalid resource argument: %s\n", resourceArray[i])
                    os.Exit(1)
                }

                resourceMap := map[string]interface{}{}
                resourceMap["box_id"] = resource[0]

                secretIdList := []string{}
                for i := 1; i < len(resource); i +=1 {
                    secretIdList = append(secretIdList, resource[i])
                }
                resourceMap["secret_id"] = secretIdList
                resourceParams = append(resourceParams, resourceMap)
            }
            params["resources"] = resourceParams
        }


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

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "UpdatePolicy")
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
                fmt.Println("\nPolicy not found\n")
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
    rootCmd.AddCommand(updatePolicyCmd)
    updatePolicyCmd.Flags().StringP("policyid", "p", "",
                                    "Id or name of the Policy to update")
    updatePolicyCmd.Flags().IntP("revision", "R", 0,
                                 "Revision number of the Policy")
    updatePolicyCmd.Flags().StringP("description", "d", "",
                                    "Short description for the Policy. " +
                                    "\"unset\" to clear.")
    updatePolicyCmd.Flags().StringArrayP("local-user", "l", []string{},
                                         "Local user name of users to be added " +
                                         "as principals. This option is repeatable.")
    updatePolicyCmd.Flags().StringArrayP("ad-upn", "u", []string{},
                                         "UPN of AD users to be added as " +
                                         "principals. This option is repeatable.")
    updatePolicyCmd.Flags().StringArrayP("ad-logon-name", "L", []string{},
                                         "Logon Name of AD users to be added " +
                                         "as principals. This option is repeatable.")
    updatePolicyCmd.Flags().StringArrayP("ad-group", "g", []string{},
                                    "|| separated string containing DN & " +
                                    "display name of AD groups to be added " +
                                    "as principals.")
    updatePolicyCmd.Flags().StringArrayP("resource", "e", []string{},
                                    "Comma-separated string containing box id " +
                                    "and secret ids subsequently, to be added " +
                                    "as resources. Give * after box id to " +
                                    "include all Secrets in the Box.")
    updatePolicyCmd.Flags().StringArrayP("tagkey", "t", []string{},
                                         "Tag key to associate with the Policy." +
                                         " This option is repeatable.")
    updatePolicyCmd.Flags().StringArrayP("tagvalue", "v", []string{},
                                         "Tag value to associate with the Policy." +
                                         "This option is repeatable.")

    // mark mandatory fields as required
    updatePolicyCmd.MarkFlagRequired("policyid")
    updatePolicyCmd.MarkFlagRequired("revision")
}
