package chainio

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightningnetwork/lnd/chainntnfs"
	"github.com/stretchr/testify/mock"
)

type MockConsumer struct {
	mock.Mock
}

// Compile-time constraint to ensure MockConsumer implements Consumer.
var _ Consumer = (*MockConsumer)(nil)

// Name returns a human-readable string for this subsystem.
func (m *MockConsumer) Name() string {
	args := m.Called()
	return args.String(0)
}

// ProcessBlock takes a blockbeat and processes it. A receive-only error chan
// must be returned.
func (m *MockConsumer) ProcessBlock(b Beat) <-chan error {
	args := m.Called(b)

	return args.Get(0).(chan error)
}

type MockBlockbeat struct {
	mock.Mock
}

// Compile-time constraint to ensure MockBlockbeat implements Blockbeat.
var _ Blockbeat = (*MockBlockbeat)(nil)

// NotifyBlockProcessed signals that the block has been processed. It takes an
// error resulted from processing the block, and a quit chan of the subsystem.
func (m *MockBlockbeat) NotifyBlockProcessed(err error, quitChan chan struct{}) {
	m.Called(err, quitChan)
}

// Height returns the current block height.
func (m *MockBlockbeat) Height() int32 {
	args := m.Called()

	return args.Get(0).(int32)
}

// DispatchConcurrent sends the blockbeat to the specified consumers
// concurrently.
func (m *MockBlockbeat) DispatchConcurrent(consumers []Consumer) error {
	args := m.Called(consumers)

	return args.Error(0)
}

// DispatchSequential sends the blockbeat to the specified consumers
// sequentially.
func (m *MockBlockbeat) DispatchSequential(consumers []Consumer) error {
	args := m.Called(consumers)

	return args.Error(0)
}

// HasOutpointSpentByScript queries the block to find a spending tx that spends
// the given outpoint using the pkScript.
func (m *MockBlockbeat) HasOutpointSpentByScript(outpoint wire.OutPoint,
	pkScript txscript.PkScript) (*chainntnfs.SpendDetail, error) {

	args := m.Called(outpoint, pkScript)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*chainntnfs.SpendDetail), args.Error(1)
}

// HasOutpointSpent queries the block to find a spending tx that spends the
// given outpoint. Returns the spend details if found, otherwise nil.
func (m *MockBlockbeat) HasOutpointSpent(
	outpoint wire.OutPoint) *chainntnfs.SpendDetail {

	args := m.Called(outpoint)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(*chainntnfs.SpendDetail)
}
