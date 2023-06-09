package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"strconv"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

var ctx = context.Background()
var cli = sui.NewSuiClient("https://fullnode.testnet.sui.io:443")
var IntentBytes = []byte{0, 0, 0}

func messageWithIntent(message []byte) []byte {
	intent := IntentBytes
	intentMessage := make([]byte, len(intent)+len(message))
	copy(intentMessage, intent)
	copy(intentMessage[len(intent):], message)
	return intentMessage
}

func SuiTransfer() {
	from := "acd6fe95b754f42268e27ffae0bb82c9e282e643d9dbd947567156dfab62fb8f"
	to := "fe04efc99cf36b57f4b8fe98e26975c634321cd99b5861721494727bbc67585d"
	var gas, amount int64 = 30000000, 100000000

	rsp, err := cli.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    from,
		CoinType: "0x2::sui::SUI",
		Limit:    10,
	})
	if err != nil {
		log.Fatal(err)
	}
	var object_id string
	for _, v := range rsp.Data {
		i, err := strconv.ParseInt(v.Balance, 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		if i > gas+amount {
			object_id = v.CoinObjectId
		}
	}
	gas_str := fmt.Sprintf("%d", gas)
	amount_str := fmt.Sprintf("%d", amount)

	rsp2, err := cli.TransferSui(ctx, models.TransferSuiRequest{
		Signer:      from,
		SuiObjectId: object_id,
		GasBudget:   gas_str,
		Recipient:   to,
		Amount:      amount_str,
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Println(rsp2.TxBytes)
	tx_raw_bytes, err := base64.StdEncoding.DecodeString(rsp2.TxBytes)
	if err != nil {
		log.Fatal(err)
	}

	tx_bytes := messageWithIntent(tx_raw_bytes)

	// tmp := base64.StdEncoding.EncodeToString(tx_bytes)
	// log.Println("---base64.StdEncoding.EncodeToString(tx_bytes)------ ", tmp)

	tx_str := hex.EncodeToString(tx_bytes)

	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "05e433db-ef8e-4e9c-8b55-1fd7f82c3567",
			"chain_type": "SUI",
			"address": "%s",
			"input": {
				"raw_data": "%s"
			},
			"key": {
				"Password": ""
			}
		}
	}`, from, tx_str)
	log.Println(json_tx)

	args := []string{
		"sign_tx",
		json_tx,
	}
	cmd2 := exec.Command("./hd-wallet", args...)
	output, err := cmd2.Output()
	if err != nil {
		log.Fatal(err)
	}

	var sig_json map[string]interface{}
	err = json.Unmarshal(output, &sig_json)
	if err != nil {
		log.Println(err)
		return
	}
	sig, ok := sig_json["signature"].(string)
	if !ok {
		log.Println("signature not ok")
		return
	}
	sig_bytes, err := hex.DecodeString(sig)
	if err != nil {
		log.Fatal(err)
	}

	base64_sig := base64.StdEncoding.EncodeToString(sig_bytes)
	rsp3, err := cli.SuiExecuteTransactionBlock(ctx, models.SuiExecuteTransactionBlockRequest{
		TxBytes:   rsp2.TxBytes,
		Signature: []string{base64_sig},
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
			ShowEvents:   true,
		},
		RequestType: "WaitForLocalExecution",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp3)

}

func SuiXGetCoins() {
	rsp, err := cli.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
		CoinType: "0x2::sui::SUI",
		Limit:    5,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}
