package main

import (
	"encoding/json"
	"fmt"

	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/common"
	"log"
	"os/exec"
)

func Tron_transfer() {
	args := []string{
		"-X", "POST",
		"--url", "https://api.shasta.trongrid.io/wallet/createtransaction",
		"--header", "accept: application/json",
		"--header", "content-type: application/json",
		"--data", `{"owner_address": "TE6RbYPvFZPS4F96yLCwebU6aSu1BM6VZ6","to_address": "TCDQGqTFf1aBUP6pMuZML1eL23mw4J2xu2","visible":true,"amount": 1000}`,
	}
	cmd := exec.Command("curl", args...)

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		fmt.Println(err)
		return
	}

	hex, _ := data["raw_data_hex"].(string)
	json_tx := fmt.Sprintf(`{"method":"sign_tx","param":{"id":"c48c3353-ceed-4be0-a397-72661027aa55","chain_type":"TRON","address":"TE6RbYPvFZPS4F96yLCwebU6aSu1BM6VZ6","input":{"raw_data":"%s"},"key":{"Password":""}}}`, hex)
	log.Println(json_tx)
	//fmt.Println()
	args2 := []string{
		"sign_tx",
		json_tx,
	}
	cmd2 := exec.Command("./hd-wallet", args2...)

	output2, err2 := cmd2.Output()
	if err2 != nil {
		log.Fatal(err)
	}

	var data2 map[string]interface{}
	err = json.Unmarshal(output2, &data2)
	if err != nil {
		fmt.Println(err)
		return
	}

	sig, _ := data2["signatures"]
	fmt.Println(sig)
	delete(data, "raw_data_hex")
	data["signature"] = sig
	tron_sign, err3 := json.Marshal(data)
	if err3 != nil {
		fmt.Println(err)
		return
	}

	sign_str := string(tron_sign)
	fmt.Println(string(sign_str))

	args3 := []string{
		"-X", "POST",
		"--url", "https://api.shasta.trongrid.io/wallet/broadcasttransaction",
		"--header", "accept: application/json",
		"--header", "content-type: application/json",
		"--data", sign_str,
	}

	cmd3 := exec.Command("curl", args3...)

	output3, err4 := cmd3.Output()
	if err4 != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output3))
}
