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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultTokenFilename = "pasm_token.txt"

const PASMCLIDataSubdir = "pasmcli.data"

type leaseInfo struct {
	BoxId     string `json:"box_id"`
	SecretId  string `json:"secret_id"`
	LeaseId   string `json:"lease_id"`
	ExpiresAt string `json:"expires_at"`
	Renewable string `json:"renewable"`
	Version   int    `json:"version"`
}

type proxyPortInfo struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	UserName string `json:"username"`
}

type proxyInfo struct {
	PasmUserName string          `json:"pasm_user_name"`
	ProxyInfos   []proxyPortInfo `json:"proxy_infos"`
}

var gLeaseInfo leaseInfo

func GetDataDir() (string, error) {
	baseDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	pasmCLIDir := filepath.Join(baseDir, PASMCLIDataSubdir)

	_, err = os.Stat(pasmCLIDir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(pasmCLIDir, 0755)
		if errDir != nil {
			return "", err
		}
	}
	return pasmCLIDir, nil
}

func GetEndPoint(server string, version string, action string) string {
	if server == "" {
		server = GetServer()
	}
	return fmt.Sprintf("https://%s/vault/%s/%s/", server, version, action)
}

func GetEndPoint2(server string, version string, action string) string {
	if server == "" {
		server = GetServer()
	}
	return fmt.Sprintf("https://%s/vault/%s/%s", server, version, action)
}

func GetLeaseFilePath(boxId string, secretId string, version int) (string, error) {
	vaultDataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}

	var identifier string
	if version != -1 {
		identifier = fmt.Sprintf("%s|%s|%d", boxId, secretId, version)
	} else {
		identifier = fmt.Sprintf("%s|%s", boxId, secretId)
	}

	b64Identifier := base64.StdEncoding.EncodeToString([]byte(identifier))
	return fmt.Sprintf("%s/vault_lease_%s.txt", vaultDataDir, b64Identifier), nil
}

func SaveLeaseInfo(leaseFile string,
	boxId string,
	secretId string,
	leaseId string,
	expiresAt string,
	renewable string,
	version int) (string, error) {
	info := leaseInfo{
		BoxId:     boxId,
		SecretId:  secretId,
		LeaseId:   leaseId,
		ExpiresAt: expiresAt,
		Renewable: renewable,
		Version:   version}

	file, err := os.Create(leaseFile)
	if err != nil {
		return leaseFile, err
	}
	defer file.Close()
	return leaseFile, json.NewEncoder(file).Encode(&info)
}

func GetLeaseId(LeaseFile string) (string, error) {
	file, err := os.Open(LeaseFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lInfo leaseInfo
	err = json.NewDecoder(file).Decode(&lInfo)
	if err != nil {
		return "", err
	}

	if lInfo.LeaseId == "" {
		return "", fmt.Errorf("Invalid or corrupt lease file - missing lease_id")
	}

	return lInfo.LeaseId, nil
}

func AuthTokenKV() map[string]string {
	return map[string]string{"X-VAULT-AUTH": GetAccessToken()}
}

// TODO: Refactor import csv command to make use of this function. To be done post 10.2
func uploadCsv(params map[string]interface{}) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	}
	endpoint := GetEndPoint("", "1.0", "ImportCSVSecrets")
	ret, err := DoPostFormData(endpoint,
		GetCACertFile(),
		AuthTokenKV(),
		jsonParams,
		"application/json")
	if err != nil {
		fmt.Printf("\nHTTP request failed: %s\n", err)
		os.Exit(4)
	}
	// type assertion
	retBytes := ret["data"].(*bytes.Buffer)
	retStatus := ret["status"].(int)
	retStr := retBytes.String()

	if retStr == "" && retStatus == 404 {
		fmt.Println("\nAction denied\n")
		os.Exit(5)
	}

	// make a decision on what to exit with
	retMap := JsonStrToMap(retStr)
	if _, present := retMap["error"]; present {
		os.Exit(3)
	}
}
