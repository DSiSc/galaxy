package participates

import (
	"github.com/DSiSc/contractsManage/contracts"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/monkey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
	}
}

func Test_NewParticipates(t *testing.T) {
	asserts := assert.New(t)
	conf := mock_conf(common.SoloPolicy)
	participate, err := NewParticipates(conf)
	asserts.NotNil(participate)
	asserts.Nil(err)

	conf = mock_conf("random")
	participate, err = NewParticipates(conf)
	asserts.NotNil(err)
	asserts.Nil(participate)

	conf = mock_conf(common.DposPolicy)
	monkey.Patch(contracts.NewVotingContract, func() contracts.Voting {
		return &contracts.VotingContract{}
	})
	participate, err = NewParticipates(conf)
	asserts.Nil(err)
	asserts.NotNil(participate)
	asserts.Equal(common.DposPolicy, participate.PolicyName())
}
