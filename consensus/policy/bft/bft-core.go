package bft

import (
	"github.com/DSiSc/craft/log"
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
	log.Debug("Replica %d processing event", instance.id)
	switch et := e.(type) {
	default:
		log.Warn("Replica %d received an unknown message type %T", instance.id, et)
	}

	if err != nil {
		log.Warn(err.Error())
	}

	return nil
}
