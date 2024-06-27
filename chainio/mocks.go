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
