package bft

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
)

type bftcore struct {
	id uint64
}

func NewBFTCore(id uint64) tools.Receiver {
	return &bftcore{
		id: id,
	}
}

func (instance *bftcore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.id)
	switch et := e.(type) {
	case messages.Request:
		log.Info("receive request from replica %d.", instance.id)
	case messages.Propoasl:
		log.Info("receive proposal from replica %d.", instance.id)
	default:
		log.Warn("replica %d received an unknown message type %T", instance.id, et)
	}

	if err != nil {
		log.Warn(err.Error())
	}

	return nil
}
