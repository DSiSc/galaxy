package config

type ConsensusConfig struct {
	PolicyName       string
	Timeout          ConsensusTimeout
	EnableEmptyBlock bool
	SignVerifySwitch SignatureVerifySwitch
}

type SignatureVerifySwitch struct {
	SyncVerifySignature  bool
	LocalVerifySignature bool
}

type ConsensusTimeout struct {
	TimeoutToCollectResponseMsg int64
	TimeoutToWaitCommitMsg      int64
	TimeoutToChangeView         int64
}
