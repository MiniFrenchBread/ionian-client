package main

import (
	"fmt"

	"github.com/Ionian-Web3-Storage/ionian-client/node"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

const KvClientAddr = "http://127.0.0.1:6789"

func main() {
	client, err := node.NewClient(KvClientAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	streamId := ethCommon.HexToHash("0x000000000000000000000000000000000000000000000000000000000000f2bd")
	key := ethCommon.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	account := ethCommon.HexToAddress("0x578dd2bfc41bb66e9f0ae0802c613996440c9597")
	kvClient := client.KV()
	val, _ := kvClient.GetValue(streamId, key, 0, 1000)
	fmt.Println(string(val.Data))
	fmt.Println(kvClient.GetTransactionResult(2))
	fmt.Println(kvClient.GetHoldingStreamIds())
	fmt.Println(kvClient.HasWritePermission(account, streamId, key))
	fmt.Println(kvClient.IsAdmin(account, streamId))
	fmt.Println(kvClient.IsSpecialKey(streamId, key))
	fmt.Println(kvClient.IsWriterOfKey(account, streamId, key))
	fmt.Println(kvClient.IsWriterOfStream(account, streamId))
}
