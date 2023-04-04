package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"encoding/hex"
	"encoding/json"
	"os/exec"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
)

func FilecoinTransfer() {
	url := "https://api.hyperspace.node.glif.io/rpc/v1"
	header := &http.Header{}
	rpc_client, _, err := client.NewFullNodeRPCV1(context.Background(),
		url, *header)
	if err != nil {
		panic(err)
	}

	from := "t1dvnvvpzhlv7uqm2tdvadwl2h3o2yiyk5cmdcxpq"
	to := "t3u4dgiuawlrd2edneg233yqdopivfxw6igmel23ih2zpf26aylkoxilqzbac4qbyzpxkctbadzilsnrwxw7ha"

	sender, err := address.NewFromString(from)
	if err != nil {
		panic(err)
	}

	nonce, err := rpc_client.MpoolGetNonce(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "05e433db-ef8e-4e9c-8b55-1fd7f82c3567",
			"chain_type": "FILECOIN",
			"address": "%s",
			"input": {
				"from": "%s",
				"to": "%s",
				"nonce": %d,
				"value": "1000000000000000",
				"gas_limit": 900000000,
				"method":0,
				"gas_fee_cap": "100",
				"gas_premium": "1",
				"params": ""
			},
			"key": {
				"Password": ""
			}
		}
	}`, from, from, to, nonce)
	log.Println(json_tx)

	// 创建一个交易消息，指定发送方，接收方，金额，手续费等参数
	// msg := &types.Message{
	// 	From:       sender,
	// 	To:         receiver,
	// 	Value:      big.NewInt(1000000000), // 1 filecoin
	// 	GasLimit:   1000000,
	// 	GasFeeCap:  big.NewInt(100),
	// 	GasPremium: big.NewInt(1),
	// 	Method:     0, // 普通转账
	// 	Params:     nil,
	// }

	// // 获取链上的nonce值，防止重放攻击

	// msg.Nonce = nonce

	args := []string{
		"sign_tx",
		json_tx,
	}
	cmd2 := exec.Command("./hd-wallet", args...)
	output, err := cmd2.Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("----------- ", string(output))
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
	log.Println(sig)
	sig_data, err := hex.DecodeString(sig)
	if err != nil {
		log.Println(err)
		return
	}

	signedMsg, err := types.DecodeSignedMessage(sig_data)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(signedMsg)
	// 将签名消息发送到链上，等待确认
	cid, err := rpc_client.MpoolPush(context.Background(), signedMsg)
	if err != nil {
		panic(err)
	}
	log.Println("get cid", cid)
}
