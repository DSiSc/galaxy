package participates

import (
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_conf() *config.ParticipateConfig {
	return &config.ParticipateConfig{
		PolicyName: "solo",
	}
}

func Test_NewParticipates(t *testing.T) {
	assert := assert.New(t)
	conf := mock_conf()
	participate, err := NewParticipates(conf)
	assert.NotNil(participate)
	assert.Nil(err)
}
