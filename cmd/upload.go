package cmd

import (
	"github.com/Ionian-Web3-Storage/ionian-client/common"
	"github.com/Ionian-Web3-Storage/ionian-client/contract"
	"github.com/Ionian-Web3-Storage/ionian-client/file"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	uploadArgs struct {
		file string
		tags string

		url      string
		contract string
		key      string

		node string

		force bool
	}

	uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload file to Ionian network",
		Run:   upload,
	}
)

func init() {
	uploadCmd.Flags().StringVar(&uploadArgs.file, "file", "", "File name to upload")
	uploadCmd.MarkFlagRequired("file")
	uploadCmd.Flags().StringVar(&uploadArgs.tags, "tags", "0x", "Tags of the file")

	uploadCmd.Flags().StringVar(&uploadArgs.url, "url", "", "Fullnode URL to interact with Ionian smart contract")
	uploadCmd.MarkFlagRequired("url")
	uploadCmd.Flags().StringVar(&uploadArgs.contract, "contract", "", "Ionian smart contract to interact with")
	uploadCmd.MarkFlagRequired("contract")
	uploadCmd.Flags().StringVar(&uploadArgs.key, "key", "", "Private key to interact with smart contract")
	uploadCmd.MarkFlagRequired("key")

	uploadCmd.Flags().StringVar(&uploadArgs.node, "node", "", "Ionian storage node URL")
	uploadCmd.MarkFlagRequired("node")

	uploadCmd.Flags().BoolVar(&uploadArgs.force, "force", false, "Force to upload file even already exists")

	rootCmd.AddCommand(uploadCmd)
}

func upload(*cobra.Command, []string) {
	client := common.MustNewWeb3(uploadArgs.url, uploadArgs.key)
	defer client.Close()
	contractAddr := ethCommon.HexToAddress(uploadArgs.contract)
	flow, err := contract.NewFlowExt(contractAddr, client)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create flow contract")
	}

	node := node.MustNewClient(uploadArgs.node)
	defer node.Close()

	uploader := file.NewUploader(flow, node)
	opt := file.UploadOption{
		Tags:  hexutil.MustDecode(uploadArgs.tags),
		Force: uploadArgs.force,
	}
	if err := uploader.Upload(uploadArgs.file, opt); err != nil {
		logrus.WithError(err).Fatal("Failed to upload file")
	}
}
