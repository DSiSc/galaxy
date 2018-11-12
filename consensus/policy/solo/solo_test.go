package solo

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/tools/signature/keypair"
	"github.com/stretchr/testify/assert"
	"math"
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

func mock_proposal() *common.Proposal {
	return &common.Proposal{
		Block: &types.Block{
			Header: &types.Header{
				Height:  1,
				SigData: make([][]byte, 0),
			},
		},
	}
}

func mock_solo_proposal() *SoloProposal {
	return &SoloProposal{
		proposal: nil,
		version:  0,
		status:   common.Proposing,
	}
}

func TestNewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	sp, err := NewSoloPolicy(mockAccounts[0])
	asserts.Nil(err)
	asserts.Equal(uint8(common.SOLO_CONSENSUS_NUM), sp.tolerance)
	asserts.Equal(common.SOLO_POLICY, sp.name)
}

func Test_toSoloProposal(t *testing.T) {
	asserts := assert.New(t)
	p := mock_proposal()
	sp, _ := NewSoloPolicy(mockAccounts[0])
	proposal := sp.toSoloProposal(p)
	asserts.NotNil(proposal)
	asserts.Equal(common.Proposing, proposal.status)
	asserts.Equal(common.Version(1), proposal.version)
	asserts.NotNil(proposal.proposal)
}

func Test_prepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(mockAccounts[0])
	proposal := mock_solo_proposal()

	err := sp.prepareConsensus(proposal)
	asserts.NotNil(err)

	proposal.version = 1
	err = sp.prepareConsensus(proposal)
	asserts.Nil(err)
	asserts.Equal(common.Propose, proposal.status)

	proposal.status = common.Propose
	err = sp.prepareConsensus(proposal)
	asserts.NotNil(err)

	sp.version = math.MaxUint64
	proposal = sp.toSoloProposal(nil)
	err = sp.prepareConsensus(proposal)
	asserts.Nil(err)
}

func Test_submitConsensus(t *testing.T) {
	asserts := assert.New(t)
	proposal := mock_solo_proposal()
	sp, _ := NewSoloPolicy(mockAccounts[0])
	err := sp.submitConsensus(proposal)
	asserts.NotNil(err)
	asserts.Equal(err, fmt.Errorf("proposal status must be proposaling"))

	proposal.status = common.Propose
	err = sp.submitConsensus(proposal)
	asserts.Nil(err)
	asserts.Equal(common.Committed, proposal.status)
}

func TestSoloPolicy_ToConsensus(t *testing.T) {
	asserts := assert.New(t)
	proposal := mock_proposal()
	sp, _ := NewSoloPolicy(mockAccounts[0])

	err := sp.ToConsensus(proposal)
	asserts.Equal(err, fmt.Errorf("local verify failed"))

	var role = make(map[account.Account]commonr.Roler)
	role[mockAccounts[0]] = commonr.Master
	err = sp.Initialization(role, mockAccounts[:1])
	var v *validator.Validator
	monkey.PatchInstanceMethod(reflect.TypeOf(v), "ValidateBlock", func(*validator.Validator, *types.Block) (*types.Header, error) {
		return nil, nil
	})
	var fakeSignature = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	proposal.Block.Header.SigData = append(proposal.Block.Header.SigData, fakeSignature)
	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[2].Address, fmt.Errorf("invalid signature")
	})
	err = sp.ToConsensus(proposal)
	asserts.Equal(err, fmt.Errorf("not enough valid signature"))

	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[2].Address, nil
	})
	err = sp.ToConsensus(proposal)
	asserts.Equal(err, fmt.Errorf("absence self signature"))

	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[0].Address, nil
	})
	err = sp.ToConsensus(proposal)
	asserts.Nil(err)

}

func TestSolo_prepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(mockAccounts[0])
	sp.version = math.MaxUint64
	proposal := sp.toSoloProposal(nil)
	err := sp.prepareConsensus(proposal)
	asserts.Nil(err)
}

func TestSoloPolicy_PolicyName(t *testing.T) {
	sp, err := NewSoloPolicy(mockAccounts[0])
	assert.Nil(t, err)
	assert.NotNil(t, sp)
	assert.Equal(t, common.SOLO_POLICY, sp.PolicyName())
}

func Test_toConsensus(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0])
	err := sp.toConsensus(nil)
	assert.Equal(t, false, err)
}

func TestSoloPolicy_Start(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0])
	sp.Start()
}

func TestSoloPolicy_Halt(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0])
	sp.Halt()
}

func TestSoloPolicy_Initialization(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0])
	var role = make(map[account.Account]commonr.Roler)
	role[mockAccounts[0]] = commonr.Master
	err := sp.Initialization(role, mockAccounts[:2])
	assert.Equal(t, err, fmt.Errorf("role and peers not in consistent"))

	err = sp.Initialization(role, mockAccounts[:1])
	assert.Nil(t, err)

	role[mockAccounts[1]] = commonr.Slave
	err = sp.Initialization(role, mockAccounts[:2])
	assert.Equal(t, err, fmt.Errorf("solo policy only support one participate"))
}
