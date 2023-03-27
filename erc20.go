package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const erc20_url = "https://rpc.ankr.com/eth_goerli"

func Erc20_transfer() {
	my_address := "0x540d00F4bB9f26bBCA6CDC5Ffcf3ddA7e56731E5"
	contruct_address := "0x5a94Dc6cc85fdA49d8E9A8b85DDE8629025C42be"
	to_address := common.HexToAddress("0x62cBF86a4Fa9A5AAA25Bd3601d0Eb89e26F6FB66")

	nonce_req := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["%s","latest"],"id":1}`, my_address)
	args := []string{
		"-X", "POST",
		"--url", erc20_url,
		"--header", "accept: application/json",
		"--header", "content-type: application/json",
		"--data", nonce_req,
	}

	cmd := exec.Command("curl", args...)

	nonce_output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var nonce_data map[string]interface{}
	err = json.Unmarshal(nonce_output, &nonce_data)
	if err != nil {
		log.Fatal(err)
	}

	hex_nonce, _ := nonce_data["result"].(string)
	nonce, err := strconv.ParseInt(hex_nonce[2:], 16, 32)
	if err != nil {
		log.Fatal(err)
	}

	contractABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		panic(err)
	}
	amount := big.NewInt(100000)

	call_data, err := contractABI.Pack("transfer", to_address, amount)

	encoded := hex.EncodeToString(call_data)
	fmt.Println("call data", encoded)
	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "bf8fe2ca-351a-49ba-b78b-49e9de791b39",
			"chain_type": "ETHEREUM",
			"address": "%s",
			"input": {
				"nonce": "%d",
				"to": "%s",
				"value": "0",
				"gas_price": "200000000000",
				"gas": "60000",
				"data": "%s",
				"network": "GOERLI"
			},
			"key": {
				"Password": ""
			}
		}
	}`, my_address, nonce, contruct_address, encoded)
	log.Println(json_tx)
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
		log.Println(err)
		return
	}

	sig, _ := sig_json["signature"].(string)
	log.Println(sig)

	sign_str := fmt.Sprintf(`{"jsonrpc":"2.0",
		"method":"eth_sendRawTransaction",
			"params":["0x%s"],
			"id":3}`, sig)

	log.Println(sign_str)
	args3 := []string{
		"-X", "POST",
		"--url", erc20_url,
		"--header", "accept: application/json",
		"--header", "content-type: application/json",
		"--data", sign_str,
	}

	cmd3 := exec.Command("curl", args3...)

	output3, err := cmd3.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output3))
}
