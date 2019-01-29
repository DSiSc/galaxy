package solo

import (
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/tools/signature/keypair"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"
)

var events types.EventCenter

type Event struct {
	m           sync.RWMutex
	Subscribers map[types.EventType]map[types.Subscriber]types.EventFunc
}

func NewEvent() types.EventCenter {
	return &Event{
		Subscribers: make(map[types.EventType]map[types.Subscriber]types.EventFunc),
	}
}

//  adds a new subscriber to Event.
func (e *Event) Subscribe(eventType types.EventType, eventFunc types.EventFunc) types.Subscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.Subscribers[eventType]
	if !ok {
		e.Subscribers[eventType] = make(map[types.Subscriber]types.EventFunc)
	}
	e.Subscribers[eventType][sub] = eventFunc

	return sub
}

func (e *Event) UnSubscribe(eventType types.EventType, subscriber types.Subscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not exist")
		return
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return
}

func (e *Event) Notify(eventType types.EventType, value interface{}) (err error) {

	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not register")
		return
	}

	switch value.(type) {
	case error:
		log.Error("Receive errors is [%v].", value)
	}
	log.Info("Receive eventType is [%d].", eventType)

	for _, event := range subs {
		go e.NotifySubscriber(event, value)
	}
	return nil
}

func (e *Event) NotifySubscriber(eventFunc types.EventFunc, value interface{}) {
	if eventFunc == nil {
		return
	}

	// invoke subscriber event func
	eventFunc(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventType, _ := range e.Subscribers {
		if err := e.Notify(eventType, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// unsubscribe all event and subscriber elegant
func (e *Event) UnSubscribeAll() {
	for eventtype, _ := range e.Subscribers {
		subs, ok := e.Subscribers[eventtype]
		if !ok {
			continue
		}
		for subscriber, _ := range subs {
			delete(subs, subscriber)
			close(subscriber)
		}
	}
	// TODO: open it when txswitch and blkswith stop complete
	//e.Subscribers = make(map[types.EventType]map[types.Subscriber]types.EventFunc)
	return
}

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

var MockSignatureVerifySwitch = config.SignatureVerifySwitch{
	SyncVerifySignature:  true,
	LocalVerifySignature: true,
}

func TestNewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	sp, err := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	asserts.Nil(err)
	asserts.Equal(uint8(common.SoloConsensusNum), sp.tolerance)
	asserts.Equal(common.SoloPolicy, sp.name)
}

func Test_toSoloProposal(t *testing.T) {
	asserts := assert.New(t)
	p := mock_proposal()
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	proposal := sp.toSoloProposal(p)
	asserts.NotNil(proposal)
	asserts.Equal(common.Proposing, proposal.status)
	asserts.Equal(common.Version(1), proposal.version)
	asserts.NotNil(proposal.proposal)
}

func Test_prepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
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
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	err := sp.submitConsensus(proposal)
	asserts.NotNil(err)
	asserts.Equal(err, fmt.Errorf("proposal status must be Propose"))

	proposal.status = common.Propose
	err = sp.submitConsensus(proposal)
	asserts.Nil(err)
	asserts.Equal(common.Committed, proposal.status)
}

func TestSoloPolicy_ToConsensus(t *testing.T) {
	asserts := assert.New(t)
	event := NewEvent()
	blockSwitch := make(chan interface{})
	var subscriber1 types.EventFunc = func(v interface{}) {
		log.Info("TEST: consensus failed event func.")
	}
	sub1 := event.Subscribe(types.EventConsensusFailed, subscriber1)
	assert.NotNil(t, sub1)
	proposal := mock_proposal()
	sp, _ := NewSoloPolicy(mockAccounts[0], blockSwitch, true, MockSignatureVerifySwitch)
	sp.Initialization(mockAccounts[0], mockAccounts[:1], event, false)

	err := sp.ToConsensus(proposal)
	asserts.Equal(err, fmt.Errorf("local verify failed"))

	var v *validator.Validator
	monkey.PatchInstanceMethod(reflect.TypeOf(v), "ValidateBlock", func(*validator.Validator, *types.Block, bool) (*types.Header, error) {
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
	go func() {
		err = sp.ToConsensus(proposal)
		assert.Nil(t, err)
	}()
	block := <-blockSwitch
	asserts.NotNil(block)
	monkey.UnpatchAll()
}

func TestSolo_prepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	sp.version = math.MaxUint64
	proposal := sp.toSoloProposal(nil)
	err := sp.prepareConsensus(proposal)
	asserts.Nil(err)
}

func TestSoloPolicy_PolicyName(t *testing.T) {
	sp, err := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	assert.Nil(t, err)
	assert.NotNil(t, sp)
	assert.Equal(t, common.SoloPolicy, sp.PolicyName())
}

func Test_toConsensus(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	err := sp.toConsensus(nil)
	assert.Equal(t, false, err)
}

func TestSoloPolicy_Start(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	sp.Start()
}

func TestSoloPolicy_Halt(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	sp.Halt()
}

func TestSoloPolicy_Initialization(t *testing.T) {
	sp, _ := NewSoloPolicy(mockAccounts[0], nil, true, MockSignatureVerifySwitch)
	sp.Initialization(mockAccounts[0], mockAccounts[:1], nil, true)
}

func TestSoloPolicy_GetConsensusResult(t *testing.T) {
	mm := time.NewTimer(3 * time.Second)
	go waitTime(mm)
	time.Sleep(1 * time.Second)
	fmt.Print("1 .\n")
	mm.Reset(5 * time.Second)
	time.Sleep(3 * time.Second)
	fmt.Print("2 .\n")
}

func waitTime(timer *time.Timer) {
	for {
		select {
		case <-timer.C:
			fmt.Print("time out.\n")
		}
	}
}
