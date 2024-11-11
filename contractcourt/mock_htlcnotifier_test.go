package contractcourt

import (
	"github.com/vanditshah99/lnd/channeldb"
	"github.com/vanditshah99/lnd/channeldb/models"
)

type mockHTLCNotifier struct {
	HtlcNotifier
}

func (m *mockHTLCNotifier) NotifyFinalHtlcEvent(key models.CircuitKey,
	info channeldb.FinalHtlcInfo) {

}
