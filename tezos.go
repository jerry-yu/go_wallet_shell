package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/goat-systems/go-tezos/v4/forge"
	"github.com/goat-systems/go-tezos/v4/keys"
	"github.com/goat-systems/go-tezos/v4/rpc"
)

func Tezos_transfer() {
	ghost_url := "https://ghostnet.tezos.marigold.dev"

	address := "tz1Sor6tc7wBtTJCpvcL4AzC2wZ8nDBxZrRT"
	to_address := "tz1Qf1pSbJzMN4VtGFfVJRgbXhBksRv36TxW"

	client, err := rpc.New(ghost_url)
	if err != nil {
		fmt.Printf("failed to initialize rpc client: %s\n", err.Error())
		os.Exit(1)
	}

	resp, counter, err := client.ContractCounter(rpc.ContractCounterInput{
		BlockID:    &rpc.BlockIDHead{},
		ContractID: address,
	})
	if err != nil {
		fmt.Printf("failed to get (%s) counter: %s\n", resp.Status(), err.Error())
		os.Exit(1)
	}
	counter++

	transaction := rpc.Transaction{
		Kind:         "transaction",
		Source:       address,
		Fee:          "2941",
		GasLimit:     "26283",
		StorageLimit: "1000",
		Counter:      strconv.Itoa(counter),
		Amount:       "10000",
		Destination:  to_address,
	}

	resp, head, err := client.Block(&rpc.BlockIDHead{})
	if err != nil {
		fmt.Printf("failed to get (%s) head block: %s\n", resp.Status(), err.Error())
		os.Exit(1)
	}

	op, err := forge.Encode(head.Hash, transaction.ToContent())
	if err != nil {
		log.Println("failed to forge transaction: ", err)
		os.Exit(1)
	}

	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "05e433db-ef8e-4e9c-8b55-1fd7f82c3567",
			"chain_type": "TEZOS",
			"address": "%s",
			"input": {
				"raw_data":"%s"
			},
			"key": {
				"Password": ""
			}
		}
	}`, address, op)
	log.Println(json_tx)
	args := []string{
		"sign_tx",
		json_tx,
	}
	cmd := exec.Command("./hd-wallet", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(output))

	var sig_json map[string]interface{}
	err = json.Unmarshal(output, &sig_json)
	if err != nil {
		log.Println(err)
		return
	}

	sig, _ := sig_json["signature"].(string)

	signature := new(keys.Signature)
	signature.Bytes, err = hex.DecodeString(sig)
	if err != nil {
		log.Println(err)
		return
	}

	// signature, err := key.SignHex(op)
	// if err != nil {
	// 	fmt.Printf("failed to sign operation: %s\n", err.Error())
	// 	os.Exit(1)
	// }

	// log.Println(hex.EncodeToString(signature.Bytes))

	resp, ophash, err := client.InjectionOperation(rpc.InjectionOperationInput{
		Operation: signature.AppendToHex(op),
	})
	if err != nil {
		fmt.Printf("failed to inject (%s): %s\n", resp.Status(), err.Error())
		os.Exit(1)
	}

	fmt.Println(ophash)
}
