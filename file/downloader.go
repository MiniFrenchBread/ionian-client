package file

import (
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/file/download"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Downloader struct {
	clients []*node.Client
}

func NewDownloader(clients ...*node.Client) *Downloader {
	if len(clients) == 0 {
		panic("storage node not specified")
	}

	return &Downloader{
		clients: clients,
	}
}

func (downloader *Downloader) Download(root, filename string, proof bool) error {
	hash := common.HexToHash(root)

	// Query file info from storage node
	info, err := downloader.queryFile(hash)
	if err != nil {
		return errors.WithMessage(err, "Failed to query file info")
	}

	// Check file existence before downloading
	if err = downloader.checkExistence(filename, hash); err != nil {
		return errors.WithMessage(err, "Failed to check file existence")
	}

	// Download segments
	if err = downloader.downloadFile(filename, hash, int64(info.Tx.Size), proof); err != nil {
		return errors.WithMessage(err, "Failed to download file")
	}

	// Validate the downloaded file
	if err = downloader.validateDownloadFile(root, filename, int64(info.Tx.Size)); err != nil {
		return errors.WithMessage(err, "Failed to validate downloaded file")
	}

	return nil
}

func (downloader *Downloader) queryFile(root common.Hash) (info *node.FileInfo, err error) {
	// requires file finalized on all storage nodes
	for _, v := range downloader.clients {
		info, err = v.Ionian().GetFileInfo(root)
		if err != nil {
			return nil, errors.WithMessagef(err, "Failed to get file info on node %v", v.URL())
		}

		if info == nil {
			return nil, errors.WithMessagef(err, "File not found on node %v", v.URL())
		}

		if !info.Finalized {
			return nil, errors.WithMessagef(err, "File not finalized on node %v", v.URL())
		}
	}

	logrus.WithField("file", info).Debug("File found by root hash")

	return
}

func (downloader *Downloader) checkExistence(filename string, hash common.Hash) error {
	file, err := Open(filename)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return errors.WithMessage(err, "Failed to open file")
	}

	defer file.Close()

	tree, err := file.MerkleTree()
	if err != nil {
		return errors.WithMessage(err, "Failed to create file merkle tree")
	}

	if tree.Root().Hex() == hash.Hex() {
		return errors.New("File already exists")
	}

	return errors.New("File already exists with different hash")
}

func (downloader *Downloader) downloadFile(filename string, root common.Hash, size int64, proof bool) error {
	file, err := download.CreateDownloadingFile(filename, root, size)
	if err != nil {
		return errors.WithMessage(err, "Failed to create downloading file")
	}
	defer file.Close()

	logrus.WithField("threads", len(downloader.clients)).Info("Begin to download file from storage node")

	sd, err := NewSegmentDownloader(downloader.clients, file, proof)
	if err != nil {
		return errors.WithMessage(err, "Failed to create segment downloader")
	}

	if err = sd.Download(); err != nil {
		return errors.WithMessage(err, "Failed to download file")
	}

	if err := file.Seal(); err != nil {
		return errors.WithMessage(err, "Failed to seal downloading file")
	}

	logrus.Info("Completed to download file")

	return nil
}

func (downloader *Downloader) validateDownloadFile(root, filename string, fileSize int64) error {
	file, err := Open(filename)
	if err != nil {
		return errors.WithMessage(err, "Failed to open file")
	}
	defer file.Close()

	if file.Size() != fileSize {
		return errors.Errorf("File size mismatch: expected = %v, downloaded = %v", fileSize, file.Size())
	}

	tree, err := file.MerkleTree()
	if err != nil {
		return errors.WithMessage(err, "Failed to create merkle tree")
	}

	if rootHex := tree.Root().Hex(); rootHex != root {
		return errors.Errorf("Merkle root mismatch, downloaded = %v", rootHex)
	}

	logrus.Info("Succeeded to validate the downloaded file")

	return nil
}
