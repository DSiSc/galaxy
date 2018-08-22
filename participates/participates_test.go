package participates

import (
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_conf() *config.ParticipatePolicy {
	return &config.ParticipatePolicy{
		PolicyName: "solo",
	}
}

func Test_NewParticipatePolicy(t *testing.T) {
	assert := assert.New(t)
	conf := mock_conf()
	participate, err := NewParticipatePolicy(conf)
	assert.NotNil(participate)
	assert.Nil(err)
}
