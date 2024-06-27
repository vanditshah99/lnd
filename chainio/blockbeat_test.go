package chainio

import (
	"errors"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/chainntnfs"
	"github.com/lightningnetwork/lnd/fn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var dummyErr = errors.New("dummy error")

// TestNewBeat tests the NewBeat and Height functions.
func TestNewBeat(t *testing.T) {
	t.Parallel()

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{
		Height: 1,
	}

	// Create the beat and check the internal state.
	beat := NewBeat(epoch)
	require.Equal(t, epoch, beat.epoch)

	// Check the height function.
	require.Equal(t, epoch.Height, beat.Height())
}

// TestNotifyBlockProcessedSendErr asserts the error can be sent and read by
// the beat via NotifyBlockProcessed.
func TestNotifyBlockProcessedSendErr(t *testing.T) {
	t.Parallel()

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Notify the block processed.
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		beat.NotifyBlockProcessed(dummyErr, nil)
	}()

	// We expect the error to be sent to the beat's error channel
	// immediately.
	_, err := fn.RecvOrTimeout(doneChan, time.Second)
	require.NoError(t, err, "timeout sending err")

	// We expect the error to be read from the error channel.
	_, err = fn.RecvOrTimeout(beat.errChan, time.Second)
	require.NoError(t, err, "timeout reading err")
}

// TestNotifyBlockProcessedOnQuit asserts NotifyBlockProcessed exits
// immediately when the quit channel is closed.
func TestNotifyBlockProcessedOnQuit(t *testing.T) {
	t.Parallel()

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Create a quit channel to test the quitting behavior.
	quitChan := make(chan struct{})

	// Send an error to the beat to occupy the error channel.
	beat.errChan <- dummyErr

	// Notify the block processed.
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		beat.NotifyBlockProcessed(dummyErr, quitChan)
	}()

	// We expect the function to exit immediately once the quit chan is
	// closed.
	close(quitChan)
	_, err := fn.RecvOrTimeout(doneChan, time.Second)
	require.NoError(t, err)
}

// TestNotifyAndWaitOnCunsumerErr asserts when the consumer returns an error,
// it's returned by notifyAndWait.
func TestNotifyAndWaitOnCunsumerErr(t *testing.T) {
	t.Parallel()

	// Create a mock consumer.
	consumer := &MockConsumer{}
	defer consumer.AssertExpectations(t)
	consumer.On("Name").Return("mocker")

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Mock the consumer to return an error.
	//
	// Create an error chan, which is returned by the consumer.
	dummyErrChan := make(chan error, 1)
	dummyErrChan <- dummyErr

	// Mock ProcessBlock to return the error chan.
	consumer.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)
		}).Once()

	// Call the method under test.
	errChan := make(chan error, 1)
	go func() {
		errChan <- beat.notifyAndWait(consumer)
	}()

	// We expect the error to be sent to the error channel.
	result, err := fn.RecvOrTimeout(errChan, time.Second)
	require.NoError(t, err, "timeout receiving error")
	require.ErrorIs(t, result, dummyErr)
}

// TestNotifyAndWaitOnCunsumerErr asserts when the consumer successfully
// processed the beat, no error is returned.
func TestNotifyAndWaitOnCunsumerSuccess(t *testing.T) {
	t.Parallel()

	// Create a mock consumer.
	consumer := &MockConsumer{}
	defer consumer.AssertExpectations(t)
	consumer.On("Name").Return("mocker")

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Mock the consumer to return an error.
	//
	// Create an error chan, which is returned by the consumer.
	dummyErrChan := make(chan error, 1)
	dummyErrChan <- nil

	// Mock ProcessBlock to return the error chan, which contains a nil.
	consumer.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)
		}).Once()

	// Call the method under test.
	errChan := make(chan error, 1)
	go func() {
		errChan <- beat.notifyAndWait(consumer)
	}()

	// We expect a nil error to be sent to the error channel.
	result, err := fn.RecvOrTimeout(errChan, time.Second)
	require.NoError(t, err, "timeout receiving error")
	require.NoError(t, result)
}

// TestNotifyAndWaitOnCunsumerTimeout asserts when the consumer times out
// processing the block, the timeout error is returned.
func TestNotifyAndWaitOnCunsumerTimeout(t *testing.T) {
	t.Parallel()

	// Change the default timeout to be 10ms.
	DefaultProcessBlockTimeout = 10 * time.Millisecond

	// Create a mock consumer.
	consumer := &MockConsumer{}
	defer consumer.AssertExpectations(t)
	consumer.On("Name").Return("mocker")

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Mock the consumer to return an error.
	//
	// Create an error chan, which is returned by the consumer.
	dummyErrChan := make(chan error, 1)

	// Mock ProcessBlock to return the error chan, which blocks on sending.
	consumer.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)
		}).Once()

	// Call the method under test.
	errChan := make(chan error, 1)
	go func() {
		errChan <- beat.notifyAndWait(consumer)
	}()

	// We expect a timeout error to be sent to the error channel.
	result, err := fn.RecvOrTimeout(errChan, time.Second)
	require.NoError(t, err, "timeout receiving error")
	require.ErrorIs(t, result, ErrProcessBlockTimeout)
}

// TestDispatchSequential checks that the beat is sent to the consumers
// sequentially.
func TestDispatchSequential(t *testing.T) {
	t.Parallel()

	// Create three mock consumers.
	consumer1 := &MockConsumer{}
	defer consumer1.AssertExpectations(t)
	consumer1.On("Name").Return("mocker1")

	consumer2 := &MockConsumer{}
	defer consumer2.AssertExpectations(t)
	consumer2.On("Name").Return("mocker2")

	consumer3 := &MockConsumer{}
	defer consumer3.AssertExpectations(t)
	consumer3.On("Name").Return("mocker3")

	consumers := []Consumer{consumer1, consumer2, consumer3}

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{}

	// Create the beat.
	beat := NewBeat(epoch)

	// Create an error chan, which is returned by the consumers. We also
	// send a nil error to this channel so ProcessBlock will return
	// immediately.
	dummyErrChan := make(chan error, len(consumers))
	for i := 0; i < len(consumers); i++ {
		dummyErrChan <- nil
	}

	// prevConsumer specifies the previous consumer that was called.
	var prevConsumer string

	// Mock the ProcessBlock on consumers.
	consumer1.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)

			// Check the order of the consumers.
			//
			// The first consumer should have no previous consumer.
			require.Empty(t, prevConsumer)

			// Set the consumer as the previous consumer.
			prevConsumer = consumer1.Name()
		}).Once()

	consumer2.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)

			// Check the order of the consumers.
			//
			// The second consumer should see consumer1.
			require.Equal(t, consumer1.Name(), prevConsumer)

			// Set the consumer as the previous consumer.
			prevConsumer = consumer2.Name()
		}).Once()

	consumer3.On("ProcessBlock", mock.Anything).Return(dummyErrChan).Run(
		func(args mock.Arguments) {
			// Check the beat used in the consumer is a copy of the
			// original beat.
			beatUsed := args.Get(0).(Beat)
			require.Equal(t, beat.epoch, beatUsed.epoch)

			// Check the order of the consumers.
			//
			// The third consumer should see consumer2.
			require.Equal(t, consumer2.Name(), prevConsumer)

			// Set the consumer as the previous consumer.
			prevConsumer = consumer3.Name()
		}).Once()

	// Call the method under test.
	err := beat.DispatchSequential(consumers)
	require.NoError(t, err)

	// Check the previous consumer is the last consumer.
	require.Equal(t, consumer3.Name(), prevConsumer)
}
