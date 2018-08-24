package participates

import (
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
	}
}

func Test_NewParticipates(t *testing.T) {
	assert := assert.New(t)
	conf := mock_conf("solo")
	participate, err := NewParticipates(conf)
	assert.NotNil(participate)
	assert.Nil(err)

	conf = mock_conf("random")
	participate, err = NewParticipates(conf)
	assert.NotNil(err)
	assert.Nil(participate)
}
