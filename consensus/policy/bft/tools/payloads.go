package tools

import (
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/worker"
)

type SignData struct {
	Signatures [][]byte
	SignMap    map[account.Account][]byte
}

func (s *SignData) AddSignature(account account.Account, sign []byte) {
	log.Info("add %x signature.", account.Address)
	s.SignMap[account] = sign
	s.Signatures = append(s.Signatures, sign)
}

func VerifyPayload(payload *types.Block) (types.Receipts, error) {
	blockStore, err := blockchain.NewBlockChainByBlockHash(payload.Header.PrevBlockHash)
	if nil != err {
		log.Error("Get NewBlockChainByBlockHash failed.")
		return nil, err
	}
	worker := worker.NewWorker(blockStore, payload)
	err = worker.VerifyBlock()
	if err != nil {
		log.Error("The block %d verified failed with err %v.", payload.Header.Height, err)
		return nil, err
	}

	return worker.GetReceipts(), nil
}

func SignPayload(account account.Account, digest types.Hash) ([]byte, error) {
	sign, err := signature.Sign(&account, digest[:])
	if nil != err {
		log.Error("archive signature occur error %x.", err)
		return nil, err
	}
	log.Info("archive signature for %x successfully with sign %x.", digest, sign)
	return sign, nil
}
