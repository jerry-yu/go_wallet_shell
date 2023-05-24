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
	"github.com/block-vision/sui-go-sdk/signer"
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

func SuiExecuteTransactionBlock() {
	rsp, err := cli.SuiExecuteTransactionBlock(ctx, models.SuiExecuteTransactionBlockRequest{
		TxBytes:   "AAACAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQEAAAAAAAAAAQEAjgDW4hJZlqvw654RGR3SdndKkdjoC0pzXQLxja/NUahLowQAAAAAACBEQGwClI9RQX68dzbN7PN29/Pw/Sc1hbtZwNAny7wZ+wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMKc3VpX3N5c3RlbRZyZXF1ZXN0X3dpdGhkcmF3X3N0YWtlAAIBAAABAQC3+Y0yfxn2dDR+HkBkFAglMULW5+UJOnyW7ajN/X2btQEqzrI5x8BMQ6LjmCSgAykfjisdYCcyTfW79nyzDB/PvtZBpwAAAAAAIAm+IREDziwoZLm7lc4ZKegZ2J5viEgoss9zgrFkHLh6t/mNMn8Z9nQ0fh5AZBQIJTFC1uflCTp8lu2ozf19m7XoAwAAAAAAAFDhjyoAAAAAAA==",
		Signature: []string{"ALISOTYXKlmBvQ1Uc4UrlWieczU9cGwkyT0Mg65Y2r6VvriElBGB63JDjqg1714Z8B0m84g3S4yrJIIws+leugKOjKY5Wf3dV/la/GVL26whJPWy7WsrWUH2wtmlmcgN6w=="},
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

	utils.PrettyPrint(rsp)
}

func SuiDryRunTransactionBlock() {
	rsp, err := cli.SuiDryRunTransactionBlock(ctx, models.SuiDryRunTransactionBlockRequest{
		TxBytes: "AAACAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQEAAAAAAAAAAQEAjgDW4hJZlqvw654RGR3SdndKkdjoC0pzXQLxja/NUahLowQAAAAAACBEQGwClI9RQX68dzbN7PN29/Pw/Sc1hbtZwNAny7wZ+wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMKc3VpX3N5c3RlbRZyZXF1ZXN0X3dpdGhkcmF3X3N0YWtlAAIBAAABAQC3+Y0yfxn2dDR+HkBkFAglMULW5+UJOnyW7ajN/X2btQEqzrI5x8BMQ6LjmCSgAykfjisdYCcyTfW79nyzDB/PvtZBpwAAAAAAIAm+IREDziwoZLm7lc4ZKegZ2J5viEgoss9zgrFkHLh6t/mNMn8Z9nQ0fh5AZBQIJTFC1uflCTp8lu2ozf19m7XoAwAAAAAAAFDhjyoAAAAAAA==",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiDevInspectTransactionBlock() {
	rsp, err := cli.SuiDevInspectTransactionBlock(ctx, models.SuiDevInspectTransactionBlockRequest{
		Sender:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		TxBytes:  "AAACAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQEAAAAAAAAAAQEAjgDW4hJZlqvw654RGR3SdndKkdjoC0pzXQLxja/NUahLowQAAAAAACBEQGwClI9RQX68dzbN7PN29/Pw/Sc1hbtZwNAny7wZ+wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMKc3VpX3N5c3RlbRZyZXF1ZXN0X3dpdGhkcmF3X3N0YWtlAAIBAAABAQC3+Y0yfxn2dDR+HkBkFAglMULW5+UJOnyW7ajN/X2btQEqzrI5x8BMQ6LjmCSgAykfjisdYCcyTfW79nyzDB/PvtZBpwAAAAAAIAm+IREDziwoZLm7lc4ZKegZ2J5viEgoss9zgrFkHLh6t/mNMn8Z9nQ0fh5AZBQIJTFC1uflCTp8lu2ozf19m7XoAwAAAAAAAFDhjyoAAAAAAA==",
		GasPrice: "1000",
		Epoch:    "87",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

// send signed transactions
func SignAndExecuteTransactionBlock() {
	signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	priKey := signerAccount.PriKey
	fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

	rsp, err := cli.TransferSui(ctx, models.TransferSuiRequest{
		Signer:      signerAccount.Address,
		SuiObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "100000000",
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Amount:      "1",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// see the successful transaction url: https://suivision.xyz/txblock/C7iYsH4tU5RdY1KBeNax4mCBn3XLZ5UswsuDpKrVkcH6
	rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: rsp,
		PriKey:      priKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp2)
}

func SuiGetTotalTransactionBlocks() {
	rsp, err := cli.SuiGetTotalTransactionBlocks(ctx)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiGetTransactionBlock() {
	rsp, err := cli.SuiGetTransactionBlock(ctx, models.SuiGetTransactionBlockRequest{
		Digest: "CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht",
		Options: models.SuiTransactionBlockOptions{
			ShowInput:          true,
			ShowRawInput:       true,
			ShowEffects:        true,
			ShowEvents:         true,
			ShowBalanceChanges: true,
			ShowObjectChanges:  true,
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiMultiGetTransactionBlocks() {
	rsp, err := cli.SuiMultiGetTransactionBlocks(ctx, models.SuiMultiGetTransactionBlocksRequest{
		Digests: []string{"CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht", "C2zZu2dpX2sLQy2234yt6ecRiNTVgQTXeQpgw9GhxGgo"},
		Options: models.SuiTransactionBlockOptions{
			ShowInput:          true,
			ShowRawInput:       true,
			ShowEffects:        true,
			ShowEvents:         true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, transactionBlock := range rsp {
		utils.PrettyPrint(*transactionBlock)
	}
}

func SuiXQueryTransactionBlocks() {
	rsp, err := cli.SuiXQueryTransactionBlocks(ctx, models.SuiXQueryTransactionBlocksRequest{
		SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
			TransactionFilter: models.TransactionFilter{
				"FromAddress": "0x7d20dcdb2bca4f508ea9613994683eb4e76e9c4ed371169677c1be02aaf0b58e",
			},
			Options: models.SuiTransactionBlockOptions{
				ShowInput:          true,
				ShowRawInput:       true,
				ShowEffects:        true,
				ShowEvents:         true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
		},
		Limit:           5,
		DescendingOrder: false,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, transactionBlock := range rsp.Data {
		utils.PrettyPrint(transactionBlock)
	}
}

func MoveCall() {
	rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
		Module:          "auction",
		Function:        "start_an_auction",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
			"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
			"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
			"web3",
			"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		},
		Gas:       "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget: "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func MergeCoins() {
	rsp, err := cli.MergeCoins(ctx, models.MergeCoinsRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		PrimaryCoin: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		CoinToMerge: "0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7",
		Gas:         "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SplitCoin() {
	rsp, err := cli.SplitCoin(ctx, models.SplitCoinRequest{
		Signer:       "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CoinObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		SplitAmounts: []string{"1000", "1000"},
		Gas:          "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SplitCoinEqual() {
	rsp, err := cli.SplitCoinEqual(ctx, models.SplitCoinEqualRequest{
		Signer:       "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CoinObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		SplitCount:   "2",
		Gas:          "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func Publish() {
	rsp, err := cli.Publish(ctx, models.PublishRequest{
		Sender: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CompiledModules: []string{
			"oRzrCwUAAAALAQAOAg4kAzJZBIsBFgWhAZoBB7sC5wEIogQoBsoECgrUBBoM7gSjAQ2RBgQAAAEBAQIBAwEEAQUBBgAHCAAACAgAAAkIAAIKDAEAAQQLAgAGDAIAAxIEAAEUBAEAAQANAAEAAA4CAQAADwMBAAAQBAEAAhUHCAEAAhYJCgEAARcLDAEAARgNCAEAAxkEDgAGGg8QAAUFEQEBCAEVEwgBAAIbFBUBAAMcDgEAAR0BDAEABR4ZAQEIBAYFBgYGBwYKAwsGDAYKFgoXDgYPGAMHCAEHCwMBCAQHCAUAAwYIAgcIAQcIBQEIAAEHCAUCBwsHAQgECwcBCAQBCAQBBgsDAQkAAQMBBwsDAQkAAQcLBwEJAAIHCwcBCQADAQsHAQkAAgcLBwEJAAsHAQkAAQgGAQYIBQEFAgkABQIDCwMBCAQBBgsHAQkAAwcLBwEJAAMHCAUBCwMBCQABCwMBCAQBCAIBCAEBCQAGZG9udXRzB2JhbGFuY2UEY29pbgZvYmplY3QDc3VpCHRyYW5zZmVyCnR4X2NvbnRleHQFRG9udXQJRG9udXRTaG9wDFNob3BPd25lckNhcARDb2luA1NVSQlUeENvbnRleHQJYnV5X2RvbnV0D2NvbGxlY3RfcHJvZml0cwllYXRfZG9udXQEaW5pdAJpZANVSUQFcHJpY2UHQmFsYW5jZQV2YWx1ZQtiYWxhbmNlX211dAVzcGxpdARqb2luA25ldwZzZW5kZXIEdGFrZQZkZWxldGUEemVybwxzaGFyZV9vYmplY3QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAwgAAAAAAAAAAAACAREIBgECAxEIBhMDAQsHAQgEAgIBEQgGAAEEAAUmCgEuOAAKABAAFCYDEAsAAQsBAQsCAQcAJwsBOAEMAwsDCgAQABQ4AgwECwAPAQsEOAMBCgIRCBIACwIuEQk4BAIBAQQAEhAKARABOAUMAwsBDwELAwoCOAYMBAsECwIuEQk4BwICAQQADgYLABMADAELARENAgMAAAABDgoAEQgSAgoALhEJOAgLABEIBugDAAAAAAAAOAkSATgKAgEBAQIA",
		},
		Dependencies: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Gas:          "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func TransferObject() {
	rsp, err := cli.TransferObject(ctx, models.TransferObjectRequest{
		Signer:    "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		ObjectId:  "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		Gas:       "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget: "1000",
		Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func TransferSui() {
	rsp, err := cli.TransferSui(ctx, models.TransferSuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "1000000",
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Amount:      "1",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func Pay() {
	rsp, err := cli.Pay(ctx, models.PayRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7"},
		Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
		Amount:      []string{"1"},
		Gas:         "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func PaySui() {
	rsp, err := cli.PaySui(ctx, models.PaySuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
		Amount:      []string{"1"},
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func PayAllSui() {
	rsp, err := cli.PayAllSui(ctx, models.PayAllSuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func RequestAddStake() {
	rsp, err := cli.RequestAddStake(ctx, models.AddStakeRequest{
		Signer:    "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Coins:     []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Amount:    "1",
		Validator: "0x884515e99dab69c4c28662149db81ca563ed4c36e0c8ce44a58e40e25a0a64a1",
		Gas:       "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget: "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func RequestWithdrawStake() {
	rsp, err := cli.RequestWithdrawStake(ctx, models.WithdrawStakeRequest{
		Signer:         "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		StakedObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		Gas:            "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:      "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func BatchTransaction() {
	rsp, err := cli.BatchTransaction(ctx, models.BatchTransactionRequest{
		Signer: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		RPCTransactionRequestParams: []models.RPCTransactionRequestParams{
			{
				MoveCallRequestParams: &models.MoveCallRequest{
					PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
					Module:          "auction",
					Function:        "start_an_auction",
					TypeArguments:   []interface{}{},
					Arguments: []interface{}{
						"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
						"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
						"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
						"web3",
						"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
					},
				},
			},
			{
				TransferObjectRequestParams: &models.TransferObjectRequest{
					ObjectId:  "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
					Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
				},
			},
		},
		Gas:                            "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:                      "1000",
		SuiTransactionBlockBuilderMode: "DevInspect",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}
