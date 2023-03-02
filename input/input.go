package input

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightningnetwork/lnd/lntypes"
)

// Input represents an abstract UTXO which is to be spent using a sweeping
// transaction. The method provided give the caller all information needed to
// construct a valid input within a sweeping transaction to sweep this
// lingering UTXO.
type Input interface {
	// Outpoint returns the reference to the output being spent, used to
	// construct the corresponding transaction input.
	OutPoint() *wire.OutPoint

	// RequiredTxOut returns a non-nil TxOut if input commits to a certain
	// transaction output. This is used in the SINGLE|ANYONECANPAY case to
	// make sure any presigned input is still valid by including the
	// output.
	RequiredTxOut() *wire.TxOut

	// RequiredLockTime returns whether this input commits to a tx locktime
	// that must be used in the transaction including it.
	RequiredLockTime() (uint32, bool)

	// WitnessType returns an enum specifying the type of witness that must
	// be generated in order to spend this output.
	WitnessType() WitnessType

	// SignDesc returns a reference to a spendable output's sign
	// descriptor, which is used during signing to compute a valid witness
	// that spends this output.
	SignDesc() *SignDescriptor

	// CraftInputScript returns a valid set of input scripts allowing this
	// output to be spent. The returns input scripts should target the
	// input at location txIndex within the passed transaction. The input
	// scripts generated by this method support spending p2wkh, p2wsh, and
	// also nested p2sh outputs.
	CraftInputScript(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (*Script, error)

	// BlocksToMaturity returns the relative timelock, as a number of
	// blocks, that must be built on top of the confirmation height before
	// the output can be spent. For non-CSV locked inputs this is always
	// zero.
	BlocksToMaturity() uint32

	// HeightHint returns the minimum height at which a confirmed spending
	// tx can occur.
	HeightHint() uint32

	// UnconfParent returns information about a possibly unconfirmed parent
	// tx.
	UnconfParent() *TxInfo
}

// TxInfo describes properties of a parent tx that are relevant for CPFP.
type TxInfo struct {
	// Fee is the fee of the tx.
	Fee btcutil.Amount

	// Weight is the weight of the tx.
	Weight int64
}

// SignDetails is a struct containing information needed to resign certain
// inputs. It is used to re-sign 2nd level HTLC transactions that uses the
// SINGLE|ANYONECANPAY sighash type, as we have a signature provided by our
// peer, but we can aggregate multiple of these 2nd level transactions into a
// new transaction, that needs to be signed by us.
type SignDetails struct {
	// SignDesc is the sign descriptor needed for us to sign the input.
	SignDesc SignDescriptor

	// PeerSig is the peer's signature for this input.
	PeerSig Signature

	// SigHashType is the sighash signed by the peer.
	SigHashType txscript.SigHashType
}

type inputKit struct {
	outpoint        wire.OutPoint
	witnessType     WitnessType
	signDesc        SignDescriptor
	heightHint      uint32
	blockToMaturity uint32
	cltvExpiry      uint32

	// unconfParent contains information about a potential unconfirmed
	// parent transaction.
	unconfParent *TxInfo
}

// OutPoint returns the breached output's identifier that is to be included as
// a transaction input.
func (i *inputKit) OutPoint() *wire.OutPoint {
	return &i.outpoint
}

// RequiredTxOut returns a nil for the base input type.
func (i *inputKit) RequiredTxOut() *wire.TxOut {
	return nil
}

// RequiredLockTime returns whether this input commits to a tx locktime that
// must be used in the transaction including it. This will be false for the
// base input type since we can re-sign for any lock time.
func (i *inputKit) RequiredLockTime() (uint32, bool) {
	return i.cltvExpiry, i.cltvExpiry > 0
}

// WitnessType returns the type of witness that must be generated to spend the
// breached output.
func (i *inputKit) WitnessType() WitnessType {
	return i.witnessType
}

// SignDesc returns the breached output's SignDescriptor, which is used during
// signing to compute the witness.
func (i *inputKit) SignDesc() *SignDescriptor {
	return &i.signDesc
}

// HeightHint returns the minimum height at which a confirmed spending
// tx can occur.
func (i *inputKit) HeightHint() uint32 {
	return i.heightHint
}

// BlocksToMaturity returns the relative timelock, as a number of blocks, that
// must be built on top of the confirmation height before the output can be
// spent. For non-CSV locked inputs this is always zero.
func (i *inputKit) BlocksToMaturity() uint32 {
	return i.blockToMaturity
}

// Cpfp returns information about a possibly unconfirmed parent tx.
func (i *inputKit) UnconfParent() *TxInfo {
	return i.unconfParent
}

// BaseInput contains all the information needed to sweep a basic
// output (CSV/CLTV/no time lock).
type BaseInput struct {
	inputKit
}

// MakeBaseInput assembles a new BaseInput that can be used to construct a
// sweep transaction.
func MakeBaseInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32,
	unconfParent *TxInfo) BaseInput {

	return BaseInput{
		inputKit{
			outpoint:     *outpoint,
			witnessType:  witnessType,
			signDesc:     *signDescriptor,
			heightHint:   heightHint,
			unconfParent: unconfParent,
		},
	}
}

// NewBaseInput allocates and assembles a new *BaseInput that can be used to
// construct a sweep transaction.
func NewBaseInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32) *BaseInput {

	input := MakeBaseInput(
		outpoint, witnessType, signDescriptor, heightHint, nil,
	)

	return &input
}

// NewCsvInput assembles a new csv-locked input that can be used to
// construct a sweep transaction.
func NewCsvInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32,
	blockToMaturity uint32) *BaseInput {

	return &BaseInput{
		inputKit{
			outpoint:        *outpoint,
			witnessType:     witnessType,
			signDesc:        *signDescriptor,
			heightHint:      heightHint,
			blockToMaturity: blockToMaturity,
		},
	}
}

// NewCsvInputWithCltv assembles a new csv and cltv locked input that can be
// used to construct a sweep transaction.
func NewCsvInputWithCltv(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32,
	csvDelay uint32, cltvExpiry uint32) *BaseInput {

	return &BaseInput{
		inputKit{
			outpoint:        *outpoint,
			witnessType:     witnessType,
			signDesc:        *signDescriptor,
			heightHint:      heightHint,
			blockToMaturity: csvDelay,
			cltvExpiry:      cltvExpiry,
			unconfParent:    nil,
		},
	}
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returned input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (bi *BaseInput) CraftInputScript(signer Signer, txn *wire.MsgTx,
	hashCache *txscript.TxSigHashes,
	prevOutputFetcher txscript.PrevOutputFetcher, txinIdx int) (*Script,
	error) {

	signDesc := bi.SignDesc()
	signDesc.PrevOutputFetcher = prevOutputFetcher
	witnessFunc := bi.witnessType.WitnessGenerator(signer, signDesc)

	return witnessFunc(txn, hashCache, txinIdx)
}

// HtlcSucceedInput constitutes a sweep input that needs a pre-image. The input
// is expected to reside on the commitment tx of the remote party and should
// not be a second level tx output.
type HtlcSucceedInput struct {
	inputKit

	preimage []byte
}

// MakeHtlcSucceedInput assembles a new redeem input that can be used to
// construct a sweep transaction.
func MakeHtlcSucceedInput(outpoint *wire.OutPoint,
	signDescriptor *SignDescriptor, preimage []byte, heightHint,
	blocksToMaturity uint32) HtlcSucceedInput {

	return HtlcSucceedInput{
		inputKit: inputKit{
			outpoint:        *outpoint,
			witnessType:     HtlcAcceptedRemoteSuccess,
			signDesc:        *signDescriptor,
			heightHint:      heightHint,
			blockToMaturity: blocksToMaturity,
		},
		preimage: preimage,
	}
}

// MakeTaprootHtlcSucceedInput creates a new HtlcSucceedInput that can be used
// to spend an HTLC output for a taproot channel on the remote party's
// commitment transaction.
func MakeTaprootHtlcSucceedInput(op *wire.OutPoint, signDesc *SignDescriptor,
	preimage []byte, heightHint, blocksToMaturity uint32) HtlcSucceedInput {

	return HtlcSucceedInput{
		inputKit: inputKit{
			outpoint:        *op,
			witnessType:     TaprootHtlcAcceptedRemoteSuccess,
			signDesc:        *signDesc,
			heightHint:      heightHint,
			blockToMaturity: blocksToMaturity,
		},
		preimage: preimage,
	}
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returns input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (h *HtlcSucceedInput) CraftInputScript(signer Signer, txn *wire.MsgTx,
	hashCache *txscript.TxSigHashes,
	prevOutputFetcher txscript.PrevOutputFetcher, txinIdx int) (*Script,
	error) {

	desc := h.signDesc
	desc.SigHashes = hashCache
	desc.InputIndex = txinIdx
	desc.PrevOutputFetcher = prevOutputFetcher

	isTaproot := txscript.IsPayToTaproot(desc.Output.PkScript)

	var (
		witness wire.TxWitness
		err     error
	)
	if isTaproot {
		if desc.ControlBlock == nil {
			return nil, fmt.Errorf("ctrl block must be set")
		}

		desc.SignMethod = TaprootScriptSpendSignMethod

		witness, err = SenderHTLCScriptTaprootRedeem(
			signer, &desc, txn, h.preimage, nil, nil,
		)
	} else {
		witness, err = SenderHtlcSpendRedeem(
			signer, &desc, txn, h.preimage,
		)
	}
	if err != nil {
		return nil, err
	}

	return &Script{
		Witness: witness,
	}, nil
}

// HtlcsSecondLevelAnchorInput is an input type used to spend HTLC outputs
// using a re-signed second level transaction, either via the timeout or success
// paths.
type HtlcSecondLevelAnchorInput struct {
	inputKit

	// SignedTx is the original second level transaction signed by the
	// channel peer.
	SignedTx *wire.MsgTx

	// createWitness creates a witness allowing the passed transaction to
	// spend the input.
	createWitness func(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (wire.TxWitness, error)
}

// RequiredTxOut returns the tx out needed to be present on the sweep tx for
// the spend of the input to be valid.
func (i *HtlcSecondLevelAnchorInput) RequiredTxOut() *wire.TxOut {
	return i.SignedTx.TxOut[0]
}

// RequiredLockTime returns the locktime needed for the sweep tx for the spend
// of the input to be valid. For a second level HTLC timeout this will be the
// CLTV expiry, for HTLC success it will be zero.
func (i *HtlcSecondLevelAnchorInput) RequiredLockTime() (uint32, bool) {
	return i.SignedTx.LockTime, true
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returns input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (i *HtlcSecondLevelAnchorInput) CraftInputScript(signer Signer,
	txn *wire.MsgTx, hashCache *txscript.TxSigHashes,
	prevOutputFetcher txscript.PrevOutputFetcher, txinIdx int) (*Script,
	error) {

	witness, err := i.createWitness(
		signer, txn, hashCache, prevOutputFetcher, txinIdx,
	)
	if err != nil {
		return nil, err
	}

	return &Script{
		Witness: witness,
	}, nil
}

// MakeHtlcSecondLevelTimeoutAnchorInput creates an input allowing the sweeper
// to spend the HTLC output on our commit using the second level timeout
// transaction.
func MakeHtlcSecondLevelTimeoutAnchorInput(signedTx *wire.MsgTx,
	signDetails *SignDetails, heightHint uint32) HtlcSecondLevelAnchorInput {

	// Spend an HTLC output on our local commitment tx using the
	// 2nd timeout transaction.
	createWitness := func(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (wire.TxWitness, error) {

		desc := signDetails.SignDesc
		desc.SigHashes = txscript.NewTxSigHashes(txn, prevOutputFetcher)
		desc.InputIndex = txinIdx
		desc.PrevOutputFetcher = prevOutputFetcher

		return SenderHtlcSpendTimeout(
			signDetails.PeerSig, signDetails.SigHashType, signer,
			&desc, txn,
		)
	}

	return HtlcSecondLevelAnchorInput{
		inputKit: inputKit{
			outpoint:    signedTx.TxIn[0].PreviousOutPoint,
			witnessType: HtlcOfferedTimeoutSecondLevelInputConfirmed,
			signDesc:    signDetails.SignDesc,
			heightHint:  heightHint,

			// CSV delay is always 1 for these inputs.
			blockToMaturity: 1,
		},
		SignedTx:      signedTx,
		createWitness: createWitness,
	}
}

// MakeHtlcSecondLevelTimeoutTaprootInput creates an input that allows the
// sweeper to spend an HTLC output to the second level on our commitment
// transaction. The sweeper is also able to generate witnesses on demand to
// sweep the second level HTLC aggregated with other transactions.
func MakeHtlcSecondLevelTimeoutTaprootInput(signedTx *wire.MsgTx,
	signDetails *SignDetails, heightHint uint32) HtlcSecondLevelAnchorInput {

	createWitness := func(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (wire.TxWitness, error) {

		desc := signDetails.SignDesc
		if desc.ControlBlock == nil {
			return nil, fmt.Errorf("ctrl block must be set")
		}

		desc.SigHashes = txscript.NewTxSigHashes(txn, prevOutputFetcher)
		desc.InputIndex = txinIdx
		desc.PrevOutputFetcher = prevOutputFetcher

		desc.SignMethod = TaprootScriptSpendSignMethod

		return SenderHTLCScriptTaprootTimeout(
			signDetails.PeerSig, signDetails.SigHashType, signer,
			&desc, txn, nil, nil,
		)
	}

	return HtlcSecondLevelAnchorInput{
		inputKit: inputKit{
			outpoint:    signedTx.TxIn[0].PreviousOutPoint,
			witnessType: TaprootHtlcLocalOfferedTimeout,
			signDesc:    signDetails.SignDesc,
			heightHint:  heightHint,

			// CSV delay is always 1 for these inputs.
			blockToMaturity: 1,
		},
		SignedTx:      signedTx,
		createWitness: createWitness,
	}
}

// MakeHtlcSecondLevelSuccessAnchorInput creates an input allowing the sweeper
// to spend the HTLC output on our commit using the second level success
// transaction.
func MakeHtlcSecondLevelSuccessAnchorInput(signedTx *wire.MsgTx,
	signDetails *SignDetails, preimage lntypes.Preimage,
	heightHint uint32) HtlcSecondLevelAnchorInput {

	// Spend an HTLC output on our local commitment tx using the 2nd
	// success transaction.
	createWitness := func(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (wire.TxWitness, error) {

		desc := signDetails.SignDesc
		desc.SigHashes = hashCache
		desc.InputIndex = txinIdx
		desc.PrevOutputFetcher = prevOutputFetcher

		return ReceiverHtlcSpendRedeem(
			signDetails.PeerSig, signDetails.SigHashType,
			preimage[:], signer, &desc, txn,
		)
	}

	return HtlcSecondLevelAnchorInput{
		inputKit: inputKit{
			outpoint:    signedTx.TxIn[0].PreviousOutPoint,
			witnessType: HtlcAcceptedSuccessSecondLevelInputConfirmed,
			signDesc:    signDetails.SignDesc,
			heightHint:  heightHint,

			// CSV delay is always 1 for these inputs.
			blockToMaturity: 1,
		},
		SignedTx:      signedTx,
		createWitness: createWitness,
	}
}

// MakeHtlcSecondLevelSuccessTaprootInput creates an input that allows the
// sweeper to spend an HTLC output to the second level on our taproot
// commitment transaction.
func MakeHtlcSecondLevelSuccessTaprootInput(signedTx *wire.MsgTx,
	signDetails *SignDetails, preimage lntypes.Preimage,
	heightHint uint32) HtlcSecondLevelAnchorInput {

	createWitness := func(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		prevOutputFetcher txscript.PrevOutputFetcher,
		txinIdx int) (wire.TxWitness, error) {

		desc := signDetails.SignDesc
		if desc.ControlBlock == nil {
			return nil, fmt.Errorf("ctrl block must be set")
		}

		desc.SigHashes = txscript.NewTxSigHashes(txn, prevOutputFetcher)
		desc.InputIndex = txinIdx
		desc.PrevOutputFetcher = prevOutputFetcher

		desc.SignMethod = TaprootScriptSpendSignMethod

		return ReceiverHTLCScriptTaprootRedeem(
			signDetails.PeerSig, signDetails.SigHashType,
			preimage[:], signer, &desc, txn, nil, nil,
		)
	}

	return HtlcSecondLevelAnchorInput{
		inputKit: inputKit{
			outpoint:    signedTx.TxIn[0].PreviousOutPoint,
			witnessType: TaprootHtlcAcceptedLocalSuccess,
			signDesc:    signDetails.SignDesc,
			heightHint:  heightHint,

			// CSV delay is always 1 for these inputs.
			blockToMaturity: 1,
		},
		SignedTx:      signedTx,
		createWitness: createWitness,
	}
}

// Compile-time constraints to ensure each input struct implement the Input
// interface.
var _ Input = (*BaseInput)(nil)
var _ Input = (*HtlcSucceedInput)(nil)
var _ Input = (*HtlcSecondLevelAnchorInput)(nil)
