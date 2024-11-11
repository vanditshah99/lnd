package contractcourt

import (
	"context"
	"io"

	"github.com/btcsuite/btcd/wire"
	"github.com/vanditshah99/lnd/channeldb"
	"github.com/vanditshah99/lnd/channeldb/models"
	"github.com/vanditshah99/lnd/htlcswitch/hop"
	"github.com/vanditshah99/lnd/input"
	"github.com/vanditshah99/lnd/invoices"
	"github.com/vanditshah99/lnd/lntypes"
	"github.com/vanditshah99/lnd/lnwallet/chainfee"
	"github.com/vanditshah99/lnd/lnwire"
	"github.com/vanditshah99/lnd/sweep"
)

// Registry is an interface which represents the invoice registry.
type Registry interface {
	// LookupInvoice attempts to look up an invoice according to its 32
	// byte payment hash.
	LookupInvoice(context.Context, lntypes.Hash) (invoices.Invoice, error)

	// NotifyExitHopHtlc attempts to mark an invoice as settled. If the
	// invoice is a debug invoice, then this method is a noop as debug
	// invoices are never fully settled. The return value describes how the
	// htlc should be resolved. If the htlc cannot be resolved immediately,
	// the resolution is sent on the passed in hodlChan later.
	NotifyExitHopHtlc(payHash lntypes.Hash, paidAmount lnwire.MilliSatoshi,
		expiry uint32, currentHeight int32,
		circuitKey models.CircuitKey, hodlChan chan<- interface{},
		wireCustomRecords lnwire.CustomRecords,
		payload invoices.Payload) (invoices.HtlcResolution, error)

	// HodlUnsubscribeAll unsubscribes from all htlc resolutions.
	HodlUnsubscribeAll(subscriber chan<- interface{})
}

// OnionProcessor is an interface used to decode onion blobs.
type OnionProcessor interface {
	// ReconstructHopIterator attempts to decode a valid sphinx packet from
	// the passed io.Reader instance.
	ReconstructHopIterator(r io.Reader, rHash []byte,
		blindingInfo hop.ReconstructBlindingInfo) (hop.Iterator, error)
}

// UtxoSweeper defines the sweep functions that contract court requires.
type UtxoSweeper interface {
	// SweepInput sweeps inputs back into the wallet.
	SweepInput(input input.Input, params sweep.Params) (chan sweep.Result,
		error)

	// RelayFeePerKW returns the minimum fee rate required for transactions
	// to be relayed.
	RelayFeePerKW() chainfee.SatPerKWeight

	// UpdateParams allows updating the sweep parameters of a pending input
	// in the UtxoSweeper. This function can be used to provide an updated
	// fee preference that will be used for a new sweep transaction of the
	// input that will act as a replacement transaction (RBF) of the
	// original sweeping transaction, if any.
	UpdateParams(input wire.OutPoint, params sweep.Params) (
		chan sweep.Result, error)
}

// HtlcNotifier defines the notification functions that contract court requires.
type HtlcNotifier interface {
	// NotifyFinalHtlcEvent notifies the HtlcNotifier that the final outcome
	// for an htlc has been determined.
	NotifyFinalHtlcEvent(key models.CircuitKey,
		info channeldb.FinalHtlcInfo)
}
