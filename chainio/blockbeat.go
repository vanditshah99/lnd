package chainio

import (
	"fmt"

	"github.com/btcsuite/btclog/v2"
	"github.com/lightningnetwork/lnd/build"
	"github.com/lightningnetwork/lnd/chainntnfs"
	"github.com/lightningnetwork/lnd/fn"
)

// Beat implements the Blockbeat interface. It contains the block epoch and a
// buffered error chan.
//
// TODO(yy): extend this to check for confirmation status - which serves as the
// single source of truth, to avoid the potential race between receiving blocks
// and `GetTransactionDetails/RegisterSpendNtfn/RegisterConfirmationsNtfn`.
type Beat struct {
	// epoch is the current block epoch the blockbeat is aware of.
	epoch chainntnfs.BlockEpoch

	// errorChan is a buffered chan that receives an error returned from
	// processing this block.
	errorChan chan error

	// log is the customized logger for the blockbeat which prints the
	// block height.
	log btclog.Logger
}

// Compile-time check to ensure Beat satisfies the Blockbeat interface.
var _ Blockbeat = (*Beat)(nil)

// NewBeat creates a new beat with the specified block epoch and a buffered
// error chan.
func NewBeat(epoch chainntnfs.BlockEpoch) *Beat {
	b := &Beat{
		epoch:     epoch,
		errorChan: make(chan error, 1),
	}

	// Create a customized logger for the blockbeat.
	logPrefix := fmt.Sprintf("Height[%6d]:", b.Height())
	b.log = build.NewPrefixLog(logPrefix, clog)

	return b
}

// Height returns the height of the block epoch.
//
// NOTE: Part of the Blockbeat interface.
func (b *Beat) Height() int32 {
	return b.epoch.Height
}

// NotifyBlockProcessed sends a signal to the BlockbeatDispatcher to notify the
// block has been processed.
//
// NOTE: Part of the Blockbeat interface.
func (b *Beat) NotifyBlockProcessed(err error, quitChan chan struct{}) {
	fn.SendOrQuit(b.errorChan, err, quitChan)
}

// copy returns a deep copy of the blockbeat.
//
// NOTE: Part of the Blockbeat interface.
func (b *Beat) copy() Blockbeat {
	return &Beat{
		epoch:     b.epoch,
		errorChan: make(chan error, 1),
		log:       b.log,
	}
}

// logger returns the logger for the blockbeat.
//
// NOTE: Part of the private blockbeat interface.
func (b *Beat) logger() btclog.Logger {
	return b.log
}

// errChan returns a read-only chan used by the blockbeat.
//
// NOTE: Part of the private blockbeat interface.
func (b *Beat) errChan() <-chan error {
	return b.errorChan
}
