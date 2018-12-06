package tools

import (
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/worker"
	"sync"
)

type SignData struct {
	Lock       sync.RWMutex
	Signatures [][]byte
	SignMap    map[account.Account][]byte
}

func (s *SignData) AddSignature(account account.Account, sign []byte) bool {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if _, ok := s.SignMap[account]; !ok {
		log.Info("add node %d signature.", account.Extension.Id)
		s.SignMap[account] = sign
		s.Signatures = append(s.Signatures, sign)
		return true
	}
	log.Info("node %d signature has exist.", account.Extension.Id)
	return false
}

func (s *SignData) SignatureNum() int {
	s.Lock.RLock()
	defer s.Lock.RUnlock()
	return len(s.Signatures)
}

func (s *SignData) GetSignMap() map[account.Account][]byte {
	s.Lock.RLock()
	defer s.Lock.RUnlock()
	return s.SignMap
}

func (s *SignData) GetSignByAccount(account account.Account) []byte {
	s.Lock.RLock()
	defer s.Lock.RUnlock()
	return s.SignMap[account]
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
		log.Error("archive signature occur error %v.", err)
		return nil, err
	}
	log.Info("archive signature for %x successfully with sign %x.", digest, sign)
	return sign, nil
}
