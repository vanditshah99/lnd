package tlv_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/vanditshah99/lnd/tlv"
)

type nodeAmts struct {
	nodeID *btcec.PublicKey
	amt1   uint64
	amt2   uint64
}

func ENodeAmts(w io.Writer, val interface{}, buf *[8]byte) error {
	if t, ok := val.(*nodeAmts); ok {
		if err := tlv.EPubKey(w, &t.nodeID, buf); err != nil {
			return err
		}
		if err := tlv.EUint64T(w, t.amt1, buf); err != nil {
			return err
		}
		return tlv.EUint64T(w, t.amt2, buf)
	}
	return tlv.NewTypeForEncodingErr(val, "nodeAmts")
}

func DNodeAmts(r io.Reader, val interface{}, buf *[8]byte, l uint64) error {
	if t, ok := val.(*nodeAmts); ok && l == 49 {
		if err := tlv.DPubKey(r, &t.nodeID, buf, 33); err != nil {
			return err
		}
		if err := tlv.DUint64(r, &t.amt1, buf, 8); err != nil {
			return err
		}
		return tlv.DUint64(r, &t.amt2, buf, 8)
	}
	return tlv.NewTypeForDecodingErr(val, "nodeAmts", l, 49)
}

type N1 struct {
	amt       uint64
	scid      uint64
	nodeAmts  nodeAmts
	cltvDelta uint16

	alias []byte

	stream *tlv.Stream
}

func (n *N1) sizeAmt() uint64 {
	return tlv.SizeTUint64(n.amt)
}

func NewN1() *N1 {
	n := new(N1)

	n.stream = tlv.MustNewStream(
		tlv.MakeDynamicRecord(
			1, &n.amt, n.sizeAmt, tlv.ETUint64, tlv.DTUint64,
		),
		tlv.MakePrimitiveRecord(2, &n.scid),
		tlv.MakeStaticRecord(3, &n.nodeAmts, 49, ENodeAmts, DNodeAmts),
		tlv.MakePrimitiveRecord(254, &n.cltvDelta),
		tlv.MakePrimitiveRecord(401, &n.alias),
	)

	return n
}

func (n *N1) Encode(w io.Writer) error {
	return n.stream.Encode(w)
}

func (n *N1) Decode(r io.Reader) error {
	return n.stream.Decode(r)
}

type N2 struct {
	amt        uint64
	cltvExpiry uint32

	stream *tlv.Stream
}

func (n *N2) sizeAmt() uint64 {
	return tlv.SizeTUint64(n.amt)
}

func (n *N2) sizeCltv() uint64 {
	return tlv.SizeTUint32(n.cltvExpiry)
}

func NewN2() *N2 {
	n := new(N2)

	n.stream = tlv.MustNewStream(
		tlv.MakeDynamicRecord(
			0, &n.amt, n.sizeAmt, tlv.ETUint64, tlv.DTUint64,
		),
		tlv.MakeDynamicRecord(
			11, &n.cltvExpiry, n.sizeCltv, tlv.ETUint32, tlv.DTUint32,
		),
	)

	return n
}

func (n *N2) Encode(w io.Writer) error {
	return n.stream.Encode(w)
}

func (n *N2) Decode(r io.Reader) error {
	return n.stream.Decode(r)
}

var tlvDecodingFailureTests = []struct {
	name   string
	bytes  []byte
	expErr error

	// skipN2 if true, will cause the test to only be executed on N1.
	skipN2 bool
}{
	{
		name:   "type truncated",
		bytes:  []byte{0xfd},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "type truncated",
		bytes:  []byte{0xfd, 0x01},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "not minimally encoded type",
		bytes:  []byte{0xfd, 0x00, 0x01}, // spec has trailing 0x00
		expErr: tlv.ErrVarIntNotCanonical,
	},
	{
		name:   "missing length",
		bytes:  []byte{0xfd, 0x01, 0x01},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "length truncated",
		bytes:  []byte{0x0f, 0xfd},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "length truncated",
		bytes:  []byte{0x0f, 0xfd, 0x26},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "missing value",
		bytes:  []byte{0x0f, 0xfd, 0x26, 0x02},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "not minimally encoded length",
		bytes:  []byte{0x0f, 0xfd, 0x00, 0x01}, // spec has trailing 0x00
		expErr: tlv.ErrVarIntNotCanonical,
	},
	{
		name: "value truncated",
		bytes: []byte{0x0f, 0xfd, 0x02, 0x01,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		expErr: io.ErrUnexpectedEOF,
	},
	{
		name:   "greater than encoding length for n1's amt",
		bytes:  []byte{0x01, 0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		expErr: tlv.NewTypeForDecodingErr(new(uint64), "uint64", 9, 8),
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x01, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x02, 0x00, 0x01},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x03, 0x00, 0x01, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x04, 0x00, 0x01, 0x00, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "encoding for n1's amt is not minimal",
		bytes:  []byte{0x01, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		expErr: tlv.ErrTUintNotMinimal,
		skipN2: true,
	},
	{
		name:   "less than encoding length for n1's scid",
		bytes:  []byte{0x02, 0x07, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
		expErr: tlv.NewTypeForDecodingErr(new(uint64), "uint64", 7, 8),
		skipN2: true,
	},
	{
		name:   "less than encoding length for n1's scid",
		bytes:  []byte{0x02, 0x09, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
		expErr: tlv.NewTypeForDecodingErr(new(uint64), "uint64", 9, 8),
		skipN2: true,
	},
	{
		name: "less than encoding length for n1's nodeAmts",
		bytes: []byte{0x03, 0x29,
			0x02, 0x3d, 0xa0, 0x92, 0xf6, 0x98, 0x0e, 0x58, 0xd2,
			0xc0, 0x37, 0x17, 0x31, 0x80, 0xe9, 0xa4, 0x65, 0x47,
			0x60, 0x26, 0xee, 0x50, 0xf9, 0x66, 0x95, 0x96, 0x3e,
			0x8e, 0xfe, 0x43, 0x6f, 0x54, 0xeb, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x01,
		},
		expErr: tlv.NewTypeForDecodingErr(new(nodeAmts), "nodeAmts", 41, 49),
		skipN2: true,
	},
	{
		name: "less than encoding length for n1's nodeAmts",
		bytes: []byte{0x03, 0x30,
			0x02, 0x3d, 0xa0, 0x92, 0xf6, 0x98, 0x0e, 0x58, 0xd2,
			0xc0, 0x37, 0x17, 0x31, 0x80, 0xe9, 0xa4, 0x65, 0x47,
			0x60, 0x26, 0xee, 0x50, 0xf9, 0x66, 0x95, 0x96, 0x3e,
			0x8e, 0xfe, 0x43, 0x6f, 0x54, 0xeb, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x01,
		},
		expErr: tlv.NewTypeForDecodingErr(new(nodeAmts), "nodeAmts", 48, 49),
		skipN2: true,
	},
	{
		name: "n1's node_id is not a valid point",
		bytes: []byte{0x03, 0x31,
			0x04, 0x3d, 0xa0, 0x92, 0xf6, 0x98, 0x0e, 0x58, 0xd2,
			0xc0, 0x37, 0x17, 0x31, 0x80, 0xe9, 0xa4, 0x65, 0x47,
			0x60, 0x26, 0xee, 0x50, 0xf9, 0x66, 0x95, 0x96, 0x3e,
			0x8e, 0xfe, 0x43, 0x6f, 0x54, 0xeb, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x02,
		},
		expErr: secp.Error{
			Err:         secp.ErrPubKeyInvalidFormat,
			Description: "invalid public key: unsupported format: 4",
		},
		skipN2: true,
	},
	{
		name: "greater than encoding length for n1's nodeAmts",
		bytes: []byte{0x03, 0x32,
			0x02, 0x3d, 0xa0, 0x92, 0xf6, 0x98, 0x0e, 0x58, 0xd2,
			0xc0, 0x37, 0x17, 0x31, 0x80, 0xe9, 0xa4, 0x65, 0x47,
			0x60, 0x26, 0xee, 0x50, 0xf9, 0x66, 0x95, 0x96, 0x3e,
			0x8e, 0xfe, 0x43, 0x6f, 0x54, 0xeb, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		},
		expErr: tlv.NewTypeForDecodingErr(new(nodeAmts), "nodeAmts", 50, 49),
		skipN2: true,
	},
	{
		name:   "less than encoding length for n1's cltvDelta",
		bytes:  []byte{0xfd, 0x00, 0x0fe, 0x00},
		expErr: tlv.NewTypeForDecodingErr(new(uint16), "uint16", 0, 2),
		skipN2: true,
	},
	{
		name:   "less than encoding length for n1's cltvDelta",
		bytes:  []byte{0xfd, 0x00, 0xfe, 0x01, 0x01},
		expErr: tlv.NewTypeForDecodingErr(new(uint16), "uint16", 1, 2),
		skipN2: true,
	},
	{
		name:   "greater than encoding length for n1's cltvDelta",
		bytes:  []byte{0xfd, 0x00, 0xfe, 0x03, 0x01, 0x01, 0x01},
		expErr: tlv.NewTypeForDecodingErr(new(uint16), "uint16", 3, 2),
		skipN2: true,
	},
	{
		name: "valid records but invalid ordering",
		bytes: []byte{0x02, 0x08,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x26, 0x01,
			0x01, 0x2a,
		},
		expErr: tlv.ErrStreamNotCanonical,
		skipN2: true,
	},
	{
		name: "duplicate tlv type",
		bytes: []byte{0x02, 0x08,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x31, 0x02,
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x51,
		},
		expErr: tlv.ErrStreamNotCanonical,
		skipN2: true,
	},
	{
		name:   "duplicate ignored tlv type",
		bytes:  []byte{0x1f, 0x00, 0x1f, 0x01, 0x2a},
		expErr: tlv.ErrStreamNotCanonical,
		skipN2: true,
	},
	{
		name:   "type wraparound",
		bytes:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00},
		expErr: tlv.ErrStreamNotCanonical,
	},
}

// TestTLVDecodingSuccess asserts that the TLV parser fails to decode invalid
// TLV streams.
func TestTLVDecodingFailures(t *testing.T) {
	for _, test := range tlvDecodingFailureTests {
		t.Run(test.name, func(t *testing.T) {
			n1 := NewN1()
			r := bytes.NewReader(test.bytes)

			err := n1.Decode(r)
			if !reflect.DeepEqual(err, test.expErr) {
				t.Fatalf("expected N1 decoding failure: %v, "+
					"got: %v", test.expErr, err)
			}

			if test.skipN2 {
				return
			}

			n2 := NewN2()
			r = bytes.NewReader(test.bytes)

			err = n2.Decode(r)
			if !reflect.DeepEqual(err, test.expErr) {
				t.Fatalf("expected N2 decoding failure: %v, "+
					"got: %v", test.expErr, err)
			}
		})
	}
}

var tlvDecodingSuccessTests = []struct {
	name   string
	bytes  []byte
	skipN2 bool
}{
	{
		name: "empty",
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0x21, 0x00},
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0xfd, 0x02, 0x01, 0x00},
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0xfd, 0x00, 0xfd, 0x00},
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0xfd, 0x00, 0xff, 0x00},
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0xfe, 0x02, 0x00, 0x00, 0x01, 0x00},
	},
	{
		name:  "unknown odd type",
		bytes: []byte{0xff, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
	},
	{
		name:   "N1 amt=0",
		bytes:  []byte{0x01, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=1",
		bytes:  []byte{0x01, 0x01, 0x01},
		skipN2: true,
	},
	{
		name:   "N1 amt=256",
		bytes:  []byte{0x01, 0x02, 0x01, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=65536",
		bytes:  []byte{0x01, 0x03, 0x01, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=16777216",
		bytes:  []byte{0x01, 0x04, 0x01, 0x00, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=4294967296",
		bytes:  []byte{0x01, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=1099511627776",
		bytes:  []byte{0x01, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=281474976710656",
		bytes:  []byte{0x01, 0x07, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 amt=72057594037927936",
		bytes:  []byte{0x01, 0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		skipN2: true,
	},
	{
		name:   "N1 scid=0x0x550",
		bytes:  []byte{0x02, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x26},
		skipN2: true,
	},
	{
		name: "N1 node_id=023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb amount_msat_1=1 amount_msat_2=2",
		bytes: []byte{0x03, 0x31,
			0x02, 0x3d, 0xa0, 0x92, 0xf6, 0x98, 0x0e, 0x58, 0xd2,
			0xc0, 0x37, 0x17, 0x31, 0x80, 0xe9, 0xa4, 0x65, 0x47,
			0x60, 0x26, 0xee, 0x50, 0xf9, 0x66, 0x95, 0x96, 0x3e,
			0x8e, 0xfe, 0x43, 0x6f, 0x54, 0xeb, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x02},
		skipN2: true,
	},
	{
		name:   "N1 cltv_delta=550",
		bytes:  []byte{0xfd, 0x00, 0xfe, 0x02, 0x02, 0x26},
		skipN2: true,
	},
}

// TestTLVDecodingSuccess asserts that the TLV parser is able to successfully
// decode valid TLV streams.
func TestTLVDecodingSuccess(t *testing.T) {
	for _, test := range tlvDecodingSuccessTests {
		t.Run(test.name, func(t *testing.T) {
			n1 := NewN1()
			r := bytes.NewReader(test.bytes)

			err := n1.Decode(r)
			if err != nil {
				t.Fatalf("expected N1 decoding success, got: %v",
					err)
			}

			if test.skipN2 {
				return
			}

			n2 := NewN2()
			r = bytes.NewReader(test.bytes)

			err = n2.Decode(r)
			if err != nil {
				t.Fatalf("expected N2 decoding success, got: %v",
					err)
			}
		})
	}
}
