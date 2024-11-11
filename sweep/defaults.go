package sweep

import (
	"github.com/vanditshah99/lnd/lnwallet/chainfee"
)

var (
	// DefaultMaxFeeRate is the default maximum fee rate allowed within the
	// UtxoSweeper. The current value is equivalent to a fee rate of 1,000
	// sat/vbyte.
	DefaultMaxFeeRate chainfee.SatPerVByte = 1e3
)
