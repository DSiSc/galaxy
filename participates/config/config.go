package config

import (
	"github.com/DSiSc/validator/tools/account"
)

type ParticipateConfig struct {
	PolicyName   string
	Delegates    uint64
	Participates []account.Account
}
