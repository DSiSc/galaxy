package tools

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/worker"
	"github.com/DSiSc/validator/worker/common"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var mockAccounts = []account.Account{
	account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "172.0.0.1:8081"},
	},
	account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "172.0.0.1:8082",
		},
	},

	account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "172.0.0.1:8083",
		},
	},
}

var mockHash = types.Hash{
	0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

var mockSignset = [][]byte{
	{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x37, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
}

func TestSignData_AddSignature(t *testing.T) {
	sign := &SignData{
		Signatures: make([][]byte, 0),
		SignMap:    make(map[account.Account][]byte),
	}
	sign.AddSignature(mockAccounts[0], mockSignset[0])
	assert.Equal(t, 1, len(sign.Signatures))
	assert.Equal(t, len(sign.Signatures), len(sign.SignMap))
	assert.Equal(t, sign.SignMap[mockAccounts[0]], mockSignset[0])
	sign.AddSignature(mockAccounts[0], mockSignset[0])
}

func TestSignData_GetSignMap(t *testing.T) {
	sign := &SignData{
		Signatures: make([][]byte, 0),
		SignMap:    make(map[account.Account][]byte),
	}
	sign.AddSignature(mockAccounts[0], mockSignset[0])
	sign.AddSignature(mockAccounts[1], mockSignset[1])
	sign.AddSignature(mockAccounts[2], mockSignset[2])
	sigMap := sign.GetSignMap()
	assert.Equal(t, mockSignset[0], sigMap[mockAccounts[0]])
	assert.Equal(t, mockSignset[1], sigMap[mockAccounts[1]])
	assert.Equal(t, mockSignset[2], sigMap[mockAccounts[2]])

	assert.Equal(t, 3, sign.SignatureNum())
	assert.Equal(t, mockSignset[0], sign.GetSignByAccount(mockAccounts[0]))
	assert.Equal(t, mockSignset[1], sign.GetSignByAccount(mockAccounts[1]))
	assert.Equal(t, mockSignset[2], sign.GetSignByAccount(mockAccounts[2]))
}

func TestVerifyPayload(t *testing.T) {
	block := &types.Block{
		Header: &types.Header{
			Height: 1,
		},
	}
	receipt, err := VerifyPayload(block)
	assert.NotNil(t, err)
	assert.Nil(t, receipt)

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})
	receipt, err = VerifyPayload(block)
	assert.Nil(t, receipt)
	assert.Equal(t, fmt.Errorf("verify block failed"), err)

	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})
	var rr types.Receipts
	r := common.NewReceipt(nil, false, uint64(10))
	rr = append(rr, r)
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "GetReceipts", func(*worker.Worker) types.Receipts {
		return rr
	})
	receipt, err = VerifyPayload(block)
	assert.Nil(t, err)
	assert.Equal(t, rr, receipt)
	monkey.UnpatchAll()
}

var fakeSignature = []byte{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
}

func TestSignPayload(t *testing.T) {
	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return make([]byte, 0), fmt.Errorf("sign error")
	})
	signatures, err := SignPayload(mockAccounts[0], mockHash)
	assert.Equal(t, 0, len(signatures))
	assert.Equal(t, fmt.Errorf("sign error"), err)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
	})
	signatures, err = SignPayload(mockAccounts[0], mockHash)
	assert.Nil(t, err)
	assert.Equal(t, fakeSignature, signatures)
}
