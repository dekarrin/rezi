# REZI

[![Go Reference](https://pkg.go.dev/badge/github.com/dekarrin/rezi.svg)](https://pkg.go.dev/github.com/dekarrin/rezi)

The Rarefied Encoding (Compressible) for Interchange (REZI) library performs
encoding of data to a binary format. It is able to encode and decode several
built-in Go types to bytes, and automatically handles decoding and encoding of
types that implement encoding.BinaryMarshaler and encoding.BinaryUnmarshaler.

Use is simple. Import the REZI library and call `Enc` to encode and `Dec` to
decode.

```golang
import "github.com/dekarrin/rezi"

...

// Take some values:

number := 612
name := "TEREZI"

// Call rezi.Enc to encode them:

numData, err := rezi.Enc(number)
if err != nil {
    panic(err)
}

nameData, err := rezi.Enc(name)
if err != nil {
    panic(err)
}

// Append the encoded together, if desired:
var data []byte
data = append(data, numData...)
data = append(data, nameData...)

// Use rezi.Dec to decode values:

var decodedNumber int
var decodedName string

var readByteCount int
var err error

readByteCount, err = rezi.Dec(data, &decodedNumber)
if err != nil {
    panic(err)
}
data = data[readByteCount:]

readByteCount, err = rezi.Dec(data, &decodedName)
if err != nil {
    panic(err)
}
data = data[readByteCount:]
```

All data is encoded in a deterministic fashion, or as deterministically as
possible. Any nondeterminism in the resulting encoded value will arise from
functions outside of the library's control; it is up to the user to ensure that,
for instance, calling MarshalBinary on a user-defined type passed to REZI for
encoding gives a determinstic result.

The REZI format was originally created for structs in the
[Ictiobus](https://github.com/dekarrin/ictiobus) project and eventually grew
into its own library for use with other projects.

### Supported Types

At this time REZI supports the built-in types `bool`, `int`, `uint`, `int8`,
`int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`, and `string`.
Additionally, any type that implements `encoding.BinaryMarshaler` can be
encoded, and any type that implements `encoding.BinaryUnmarshaler` with a
pointer receiver can be decoded. `float32`, `float64`, `complex64`, and
`complex128` are not supported.

REZI supports slice types and map types whose values are of one of any supported
type (including those whose values are themselves slice or map values). Maps
must additionally have a key of one of the type `string`, `bool`, or one of the
integer types listed above.

REZI can also handle encoding and decoding pointers to any supported type, with
any level of indirection.

### Compression

REZI does not currently support compression itself but its results can be
compressed. Automatic compression will come in a future release.
