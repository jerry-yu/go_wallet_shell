package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"log"
	"os/exec"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const abiString = `[{
	"constant": false,
	"inputs": [
		{
			"name": "_to",
			"type": "address"
		},
		{
			"name": "_value",
			"type": "uint256"
		}
	],
	"name": "transfer",
	"outputs": [
		{
			"name": "",
			"type": "bool"
		}
	],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}]`

func call_param() (string, error) {
	tronAddress := "TUGUZJ5d7dty15AzFQ72edMDU9KcWM1pX4"
	base58Decoded, err := base58Decode(tronAddress)
	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(base58Decoded[1:21]))

	toAddress := common.BytesToAddress(base58Decoded[1:21])
	fmt.Println(toAddress)
	amount := big.NewInt(100000000)
	contractABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		panic(err)
	}

	method := contractABI.Methods["transfer"]
	input, err := method.Inputs.Pack(toAddress, amount)
	if err != nil {
		panic(err)
	}

	// data, err := contractABI.Pack("transfer", toAddress, amount)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(contractABI.Methods["transfer"].Inputs)

	// t_bytes, err := contractABI.Methods["transfer"].Inputs.Pack(args)
	// if err != nil {
	// 	panic(err)
	// }
	// data := append(methodID[:4], common.LeftPadBytes(t_bytes, 32)...)

	hexEncoded := hex.EncodeToString(input)

	return hexEncoded, nil
}

func base58Decode(address string) ([]byte, error) {
	const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	result := big.NewInt(0)
	radix := big.NewInt(58)
	for i := 0; i < len(address); i++ {
		charIndex := strings.IndexByte(base58Alphabet, address[i])
		if charIndex == -1 {
			return nil, fmt.Errorf("invalid character '%c' in Tron address", address[i])
		}
		result.Mul(result, radix)
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	resultBytes := result.Bytes()
	//resultBytes = append(bytes.Repeat([]byte{0x00}, len(address)-len(resultBytes)), resultBytes...)
	return resultBytes, nil
}

func Tron20_transfer() {
	param, err := call_param()
	call_data := fmt.Sprintf(`{"owner_address": "TE6RbYPvFZPS4F96yLCwebU6aSu1BM6VZ6",
	"contract_address": "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs",
	"visible":true,
	"function_selector": "transfer(address,uint256)",
	"call_value": 0,
	"fee_limit" : 8000000,
	"parameter": %q }`, param)

	args := []string{
		"-X", "POST",
		"--url", "https://api.shasta.trongrid.io/wallet/triggersmartcontract",
		"--header", "accept: application/json",
		"--header", "content-type: application/json",
		"--data", call_data,
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
	var real_tx_data map[string]interface{}
	real_tx_data = data["transaction"].(map[string]interface{})

	hex, _ := real_tx_data["raw_data_hex"].(string)
	json_tx := fmt.Sprintf(`{"method":"sign_tx","param":{"id":"c48c3353-ceed-4be0-a397-72661027aa55","chain_type":"TRON","address":"TE6RbYPvFZPS4F96yLCwebU6aSu1BM6VZ6","input":{"raw_data":"%s"},"key":{"Password":""}}}`, hex)
	log.Println(json_tx)
	//fmt.Println()
	args2 := []string{
		"sign_tx",
		json_tx,
	}
	cmd2 := exec.Command("./hd-wallet", args2...)

	output2, err := cmd2.Output()
	if err != nil {
		log.Fatal(err)
	}

	var sig_json map[string]interface{}
	err = json.Unmarshal(output2, &sig_json)
	if err != nil {
		fmt.Println(err)
		return
	}

	sig, _ := sig_json["signatures"]
	fmt.Println(sig)

	delete(real_tx_data, "raw_data_hex")
	real_tx_data["signature"] = sig
	tron_sign, err := json.Marshal(real_tx_data)
	if err != nil {
		fmt.Println(err)
		return
	}

	sign_str := string(tron_sign)

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
		log.Fatal(err4)
	}
	fmt.Println(string(output3))
}
