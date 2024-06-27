package chainio

import (
	"testing"

	"github.com/lightningnetwork/lnd/chainntnfs"
	"github.com/stretchr/testify/require"
)

// TestRegisterQueue tests the RegisterQueue function.
func TestRegisterQueue(t *testing.T) {
	t.Parallel()

	// Create two mock consumers.
	consumer1 := &MockConsumer{}
	defer consumer1.AssertExpectations(t)
	consumer1.On("Name").Return("mocker1")

	consumer2 := &MockConsumer{}
	defer consumer2.AssertExpectations(t)
	consumer2.On("Name").Return("mocker2")

	consumers := []Consumer{consumer1, consumer2}

	// Create a mock chain notifier.
	mockNotifier := &chainntnfs.MockChainNotifier{}
	defer mockNotifier.AssertExpectations(t)

	// Create a new dispatcher.
	b := NewBlockbeatDispatcher(mockNotifier)

	// Register the consumers.
	b.RegisterQueue(consumers)

	// Assert that the consumers have been registered.
	//
	// We should have one queue.
	require.Len(t, b.consumerQueues, 1)

	// The queue should have two consumers.
	queue, ok := b.consumerQueues[1]
	require.True(t, ok)
	require.Len(t, queue, 2)
}

// TestNotifyQueuesSuccess checks when the dispatcher successfully notifies all
// the queues, no error is returned.
func TestNotifyQueuesSuccess(t *testing.T) {
	t.Parallel()

	// Create two mock consumers.
	consumer1 := &MockConsumer{}
	defer consumer1.AssertExpectations(t)
	consumer1.On("Name").Return("mocker1")

	consumer2 := &MockConsumer{}
	defer consumer2.AssertExpectations(t)
	consumer2.On("Name").Return("mocker2")

	// Create two queues.
	queue1 := []Consumer{consumer1}
	queue2 := []Consumer{consumer2}

	// Create a mock chain notifier.
	mockNotifier := &chainntnfs.MockChainNotifier{}
	defer mockNotifier.AssertExpectations(t)

	// Create a mock beat.
	mockBeat := &MockBlockbeat{}
	defer mockBeat.AssertExpectations(t)

	// Create a new dispatcher.
	b := NewBlockbeatDispatcher(mockNotifier)

	// Register the queues.
	b.RegisterQueue(queue1)
	b.RegisterQueue(queue2)

	// Attach the blockbeat.
	b.beat = mockBeat

	// Mock the blockbeat to return nil error on DispatchSequential for
	// both queues.
	mockBeat.On("DispatchSequential", queue1).Return(nil)
	mockBeat.On("DispatchSequential", queue2).Return(nil)

	// Notify the queues. The mockers will be asserted in the end to
	// validate the calls.
	err := b.notifyQueues()
	require.NoError(t, err)
}

// TestNotifyQueuesError checks when one of the queue returns an error, this
// error is returned by the method.
func TestNotifyQueuesError(t *testing.T) {
	t.Parallel()

	// Create a mock consumer.
	consumer := &MockConsumer{}
	defer consumer.AssertExpectations(t)
	consumer.On("Name").Return("mocker1")

	// Create one queue.
	queue := []Consumer{consumer}

	// Create a mock chain notifier.
	mockNotifier := &chainntnfs.MockChainNotifier{}
	defer mockNotifier.AssertExpectations(t)

	// Create a mock beat.
	mockBeat := &MockBlockbeat{}
	defer mockBeat.AssertExpectations(t)

	// Create a new dispatcher.
	b := NewBlockbeatDispatcher(mockNotifier)

	// Register the queues.
	b.RegisterQueue(queue)

	// Attach the blockbeat.
	b.beat = mockBeat

	// Mock the blockbeat to return an error on DispatchSequential.
	mockBeat.On("DispatchSequential", queue).Return(dummyErr)

	// Notify the queues. The mockers will be asserted in the end to
	// validate the calls.
	err := b.notifyQueues()
	require.ErrorIs(t, err, dummyErr)
}
