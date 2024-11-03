package channeldb

import (
	"net"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/require"
)

var (
	addr1 = &net.TCPAddr{IP: (net.IP)([]byte{0x1}), Port: 1}
	addr2 = &net.TCPAddr{IP: (net.IP)([]byte{0x2}), Port: 2}
	addr3 = &net.TCPAddr{IP: (net.IP)([]byte{0x3}), Port: 3}
)

// TestMultiAddrSource tests that the multiAddrSource correctly merges and
// deduplicates the results of a set of AddrSource implementations.
func TestMultiAddrSource(t *testing.T) {
	t.Parallel()

	var pk1 = newTestPubKey(t)

	t.Run("both sources have results", func(t *testing.T) {
		t.Parallel()

		var (
			src1 = newMockAddrSource()
			src2 = newMockAddrSource()
		)

		// Let source 1 know of 2 addresses (addr 1 and 2) for node 1.
		src1.setAddrs(pk1, addr1, addr2)

		// Let source 2 know of 2 addresses (addr 2 and 3) for node 1.
		src2.setAddrs(pk1, addr2, addr3)

		// Create a multi-addr source that consists of both source 1
		// and 2.
		multiSrc := NewMultiAddrSource(src1, src2)

		// Query it for the addresses known for node 1. The results
		// should contain addr 1, 2 and 3.
		known, addrs, err := multiSrc.AddrsForNode(pk1)
		require.NoError(t, err)
		require.True(t, known)
		require.ElementsMatch(t, addrs, []net.Addr{addr1, addr2, addr3})
	})

	t.Run("only once source has results", func(t *testing.T) {
		t.Parallel()

		var (
			src1 = newMockAddrSource()
			src2 = newMockAddrSource()
		)

		// Let source 1 know of address 1 for node 1.
		src1.setAddrs(pk1, addr1)

		// Create a multi-addr source that consists of both source 1
		// and 2.
		multiSrc := NewMultiAddrSource(src1, src2)

		// Query it for the addresses known for node 1. The results
		// should contain addr 1.
		known, addrs, err := multiSrc.AddrsForNode(pk1)
		require.NoError(t, err)
		require.True(t, known)
		require.ElementsMatch(t, addrs, []net.Addr{addr1})
	})

	t.Run("unknown address", func(t *testing.T) {
		t.Parallel()

		var (
			src1 = newMockAddrSource()
			src2 = newMockAddrSource()
		)

		// Create a multi-addr source that consists of both source 1
		// and 2. Neither source known of node 1.
		multiSrc := NewMultiAddrSource(src1, src2)

		// Query it for the addresses known for node 1. It should return
		// false to indicate that the node is unknown to all backing
		// sources.
		known, addrs, err := multiSrc.AddrsForNode(pk1)
		require.NoError(t, err)
		require.False(t, known)
		require.Empty(t, addrs)
	})
}

type mockAddrSource struct {
	addrs map[*btcec.PublicKey][]net.Addr
}

var _ AddrSource = (*mockAddrSource)(nil)

func newMockAddrSource() *mockAddrSource {
	return &mockAddrSource{
		make(map[*btcec.PublicKey][]net.Addr),
	}
}

func (m *mockAddrSource) AddrsForNode(pub *btcec.PublicKey) (bool, []net.Addr,
	error) {

	addrs, ok := m.addrs[pub]

	return ok, addrs, nil
}

func (m *mockAddrSource) setAddrs(pub *btcec.PublicKey, addrs ...net.Addr) {
	m.addrs[pub] = addrs
}

func newTestPubKey(t *testing.T) *btcec.PublicKey {
	priv, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	return priv.PubKey()
}
