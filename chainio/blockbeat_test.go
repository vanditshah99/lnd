package chainio

import (
	"errors"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/chainntnfs"
	"github.com/lightningnetwork/lnd/fn"
	"github.com/stretchr/testify/require"
)

var errDummy = errors.New("dummy error")

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
		beat.NotifyBlockProcessed(errDummy, nil)
	}()

	// We expect the error to be sent to the beat's error channel
	// immediately.
	_, err := fn.RecvOrTimeout(doneChan, time.Second)
	require.NoError(t, err, "timeout sending err")

	// We expect the error to be read from the error channel.
	_, err = fn.RecvOrTimeout(beat.errorChan, time.Second)
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
	beat.errorChan <- errDummy

	// Notify the block processed.
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		beat.NotifyBlockProcessed(errDummy, quitChan)
	}()

	// We expect the function to exit immediately once the quit chan is
	// closed.
	close(quitChan)
	_, err := fn.RecvOrTimeout(doneChan, time.Second)
	require.NoError(t, err)
}

// TestCopyBeat asserts the copied beat has the same data but different
// internal state than the original.
func TestCopyBeat(t *testing.T) {
	t.Parallel()

	// Create a testing epoch.
	epoch := chainntnfs.BlockEpoch{
		Height: 1,
	}

	// Create the beat.
	beat := NewBeat(epoch)

	// Send an error to the beat to occupy the error channel.
	beat.NotifyBlockProcessed(errDummy, nil)

	// Copy the beat.
	beatCopy := beat.copy()

	// Send a different error to the copied beat.
	errDummy2 := errors.New("different err")
	beatCopy.NotifyBlockProcessed(errDummy2, nil)

	// Assert the copied data is the same.
	require.Equal(t, beat.Height(), beatCopy.Height())
	require.Equal(t, beat.logger(), beatCopy.logger())

	// Assert the errors are different.
	receivedErr, err := fn.RecvOrTimeout(beat.errChan(), time.Second)
	require.NoError(t, err)
	require.ErrorIs(t, errDummy, receivedErr)

	receivedErr, err = fn.RecvOrTimeout(beatCopy.errChan(), time.Second)
	require.NoError(t, err)
	require.ErrorIs(t, errDummy2, receivedErr)
}
