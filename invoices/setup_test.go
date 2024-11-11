package invoices

import (
	"testing"

	"github.com/vanditshah99/lnd/kvdb"
)

func TestMain(m *testing.M) {
	kvdb.RunTests(m)
}
