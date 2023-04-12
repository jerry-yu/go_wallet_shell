package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os/exec"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

func Deeper_transfer() {
	// This sample shows how to create a transaction to make a transfer from one an account to another.
	// Instantiate the API
	api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	if err != nil {
		panic(err)
	}

	from := "5GgG8kyJHWf1ZXQxvaNSoN5iV8GWx4NmaJqYBAP7zXUpDM1c"
	from_bytes, err := base58Decode(from)
	if err != nil {
		panic(err)
	}
	from_public := from_bytes[1 : 1+32]

	to := "5H9aLtLYyCcQUpGTHbqXhzDWzPkQt7ADV6KZjCmpMSxrbHxz"
	to_bytes, err := base58Decode(to)
	if err != nil {
		panic(err)
	}
	to_public := to_bytes[1 : 1+32]

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		log.Fatal(err)
	}

	to_addr, err := types.NewMultiAddressFromAccountID(to_public)
	if err != nil {
		log.Fatal(err)
	}

	// 1 unit of transfer
	bal, ok := new(big.Int).SetString("10000000000000000000", 10)
	if !ok {
		panic(fmt.Errorf("failed to convert balance"))
	}

	c, err := types.NewCall(meta, "Balances.transfer", to_addr, types.NewUCompact(bal))
	if err != nil {
		log.Fatal(err)
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", from_public)
	if err != nil {
		log.Fatal(err)
	}

	var accountInfo types.AccountInfo
	ok, err = api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		log.Println(accountInfo)
		panic(err)
	}

	nonce := uint32(accountInfo.Nonce)

	mb, err := codec.Encode(c)
	if err != nil {
		panic(err)
	}
	payload := types.ExtrinsicPayloadV4{
		ExtrinsicPayloadV3: types.ExtrinsicPayloadV3{
			Method:      mb,
			Era:         types.ExtrinsicEra{IsImmortalEra: true},
			Nonce:       types.NewUCompactFromUInt(uint64(nonce)),
			Tip:         types.NewUCompactFromUInt(0),
			SpecVersion: rv.SpecVersion,
			GenesisHash: genesisHash,
			BlockHash:   genesisHash,
		},
		TransactionVersion: rv.TransactionVersion,
	}

	payload_bytes, err := codec.Encode(payload)
	if err != nil {
		panic(err)
	}

	json_tx := fmt.Sprintf(`{
		"method": "sign_tx",
		"param": {
			"id": "05e433db-ef8e-4e9c-8b55-1fd7f82c3567",
			"chain_type": "DEEPER",
			"address": "%s",
			"input": {
				"raw_data":"%s"
			},
			"key": {
				"Password": ""
			}
		}
	}`, from, hex.EncodeToString(payload_bytes))
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

	var sig_json map[string]interface{}
	err = json.Unmarshal(output, &sig_json)
	if err != nil {
		log.Println(err)
		return
	}

	sig, _ := sig_json["signature"].(string)
	// skip 0x
	log.Println(sig[2:])
	sig_bytes, err := hex.DecodeString(sig[2:])
	if err != nil {
		log.Println(err)
		return
	}

	signerPubKey, err := types.NewMultiAddressFromAccountID(from_public)
	if err != nil {
		panic(err)
	}

	extSig := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(sig_bytes)},
		Era:       types.ExtrinsicEra{IsImmortalEra: true},
		Nonce:     types.NewUCompactFromUInt(uint64(nonce)),
		Tip:       types.NewUCompactFromUInt(0),
	}

	ext.Signature = extSig

	// mark the extrinsic as signed
	ext.Version |= types.ExtrinsicBitSigned

	ext_bytes, _ := codec.EncodeToHex(ext)
	fmt.Printf("ext_bytes %s \n", ext_bytes)

	//Send the extrinsic

	// o := types.SignatureOptions{
	// 	BlockHash:          genesisHash,
	// 	Era:                types.ExtrinsicEra{IsMortalEra: false},
	// 	GenesisHash:        genesisHash,
	// 	Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
	// 	SpecVersion:        rv.SpecVersion,
	// 	Tip:                types.NewUCompactFromUInt(0),
	// 	TransactionVersion: rv.TransactionVersion,
	// }

	// // Sign the transaction using Alice's default account
	// err = ext.Sign(signature.TestKeyringPairAlice, o)
	// if err != nil {
	// 	panic(err)
	// }

	// ext2_bytes, _ := codec.EncodeToHex(ext)

	// fmt.Printf("ext222_bytes %s \n", ext2_bytes)

	_, err = api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Balance transferred : %v\n", bal.String())
}
