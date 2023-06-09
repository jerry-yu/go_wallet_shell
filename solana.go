package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/btcsuite/btcutil/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"

	"github.com/portto/solana-go-sdk/program/system"
	"github.com/portto/solana-go-sdk/types"
)

func Solana_transfer() {

	c := client.NewClient("https://api.devnet.solana.com")
	from := "EpEspbDY35RikKxz7tHDbtZVLjSwgrtgSEJL73F2nsHZ"
	from_pubkey := common.PublicKeyFromBytes(base58.Decode(from))
	recipient := common.PublicKeyFromString("32bZpr4eH6LkfxmDAxrQKyumxr3Kao5pfTQeiaCfiL1Q")

	// sk, err := hex.DecodeString("6391989cdca53c9d0488f33ecc26be5e52c9869964e622f31f15a8fdd13bccc0")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // Create an account for the payer
	// payer, err := types.AccountFromSeed(sk)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	log.Println(from_pubkey.String())
	// Create a public key for the recipient

	// Create an instruction to transfer 1 SOL from payer to recipient

	instruction := system.Transfer(system.TransferParam{
		From:   from_pubkey,
		To:     recipient,
		Amount: 100_000_000,
	})

	var ctx = context.Background()
	bhash, err := c.GetLatestBlockhash(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Create a message containing the instruction and the payer account
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        from_pubkey,
		Instructions:    []types.Instruction{instruction},
		RecentBlockhash: bhash.Blockhash,
	})

	tx_bytes, err := message.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	tx_str := hex.EncodeToString(tx_bytes)

	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "05e433db-ef8e-4e9c-8b55-1fd7f82c3567",
			"chain_type": "SOLANA",
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

	//sig_bytes := payer.Sign(data)

	tx := types.Transaction{
		Message:    message,
		Signatures: []types.Signature{sig_bytes},
	}

	// Create an unsigned transaction from the message
	// tx, err := types.NewTransaction(types.NewTransactionParam{
	// 	Message: message,
	// 	Signers: []types.Account{payer},
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Sign the transaction with the payer account
	// tx.Sign(payer)

	// Send the transaction to the network and wait for confirmation
	txID, err := c.SendTransaction(ctx, tx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction sent with ID:", txID)
}
