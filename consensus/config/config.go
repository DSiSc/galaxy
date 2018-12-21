package config

type ConsensusConfig struct {
	PolicyName string
	Timeout    ConsensusTimeout
}

type ConsensusTimeout struct {
	TimeoutToCollectResponseMsg int64
	TimeoutToWaitCommitMsg      int64
	TimeoutToChangeView         int64
}
