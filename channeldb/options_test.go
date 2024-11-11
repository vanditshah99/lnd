package channeldb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vanditshah99/lnd/channeldb"
	"github.com/vanditshah99/lnd/kvdb"
)

// TestDefaultOptions tests the default options are created as intended.
func TestDefaultOptions(t *testing.T) {
	opts := channeldb.DefaultOptions()

	require.True(t, opts.NoFreelistSync)
	require.False(t, opts.AutoCompact)
	require.Equal(
		t, kvdb.DefaultBoltAutoCompactMinAge, opts.AutoCompactMinAge,
	)
	require.Equal(t, kvdb.DefaultDBTimeout, opts.DBTimeout)
	require.Equal(
		t, channeldb.DefaultRejectCacheSize, opts.RejectCacheSize,
	)
	require.Equal(
		t, channeldb.DefaultChannelCacheSize, opts.ChannelCacheSize,
	)
}
