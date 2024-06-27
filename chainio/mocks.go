package chainio

import "github.com/stretchr/testify/mock"

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
