# REZI

![Tests Status Badge](https://github.com/dekarrin/rezi/actions/workflows/tests.yml/badge.svg?branch=main&event=push)
[![Go Reference](https://pkg.go.dev/badge/github.com/dekarrin/rezi.svg)](https://pkg.go.dev/github.com/dekarrin/rezi)

The Rarefied Encoding (Compressible) for Interchange (REZI) library performs
binary marshaling of data to REZI-format bytes. It can encode and decode several
built-in Go types to bytes, and automatically handles decoding and encoding of
types that implement `encoding.BinaryMarshaler` and
`encoding.BinaryUnmarshaler`.

All data is encoded in a deterministic fashion, or as deterministically as
possible. Any non-determinism in the resulting encoded value will arise from
functions outside of the library's control; it is up to the user to ensure that,
for instance, calling MarshalBinary on a user-defined type passed to REZI for
encoding gives a deterministic result.

The REZI format was originally created for structs in the
[Ictiobus](https://github.com/dekarrin/ictiobus) project and eventually grew
into a separate library for use with other projects.

### Installation

Install REZI into your project using standard Go tools:

```bash
go get -u github.com/dekarrin/rezi@latest
```

And import the REZI library to use it:

```golang
import "github.com/dekarrin/rezi"
```

### Usage

The primary REZI format functions are `Enc()` for encoding data and `Dec` to
decode it. Both of these work similar to the `Marshal` and `Unmarshal` functions
in the `json` library. For encoding, the value to be encoded is passed in
directly, and for decoding, a pointer to a value of the correct type is passed
in.

#### Encoding

To encode a value using REZI, pass it to `Enc()`. This will return a slice of
bytes holding the encoded value.

```golang
import (
    "fmt"

    "github.com/dekarrin/rezi"
)

...

name := "Terezi"

nameData, err := rezi.Enc(name)
if err != nil {
    panic(err)
}

fmt.Println(nameData) // this will print out the encoded data bytes
```

Multiple encoded values are joined into a single slice of REZI-compatible bytes
by appending the results of `Enc()` together.

```golang
// A new value to encode:
number := 612

numData, err := rezi.Enc(number)
if err != nil {
    panic(err)
}

// we'll append the two data slices together in a new slice containing both the
// encoded name and number:

var data []byte
data = append(data, nameData...)
data = append(data, numData...)
```

You'll need to keep the order of the encoded values in mind when decoding. In
the above example, the `data` slice contains the encoded name, followed by the
encoded number.

#### Decoding

To decode data from a slice of bytes containing REZI-format data, pass it along
with a pointer to receive the value to the `Dec()` function. The data can
contain more than one value in sequence; `Dec()` will decode the one that begins
at the start of the slice, and return the number of bytes it decoded.

```golang
import (
    "fmt"

    "github.com/dekarrin/rezi"
)

...

var decodedName string
var decodedNumber int

var readByteCount int
var err error

// assume data is the []byte from the end of the Enc() example. It contains a
// REZI-format string, followed by a REZI-format int.

// decode the first value, the name:
readByteCount, err = rezi.Dec(data, &decodedName)
if err != nil {
    panic(err)
}

// skip ahead the number of bytes that were just read so that the start of data
// now points at the next REZI-encoded value
data = data[readByteCount:]

// decode the second value, the number:
readByteCount, err = rezi.Dec(data, &decodedNumber)
if err != nil {
    panic(err)
}

fmt.Println(decodedName) // "Terezi"
fmt.Println(decodedNumber) // 612
```

### Supported Types

At this time REZI supports the built-in types `bool`, `int`, `uint`, `int8`,
`int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`, and `string`.
Additionally, any type that implements `encoding.BinaryMarshaler` can be
encoded, and any type that implements `encoding.BinaryUnmarshaler` with a
pointer receiver can be decoded. `float32`, `float64`, `complex64`, and
`complex128` are not supported.

REZI supports slice types and map types whose values are of any supported type
(including those whose values are themselves slice or map values). Maps must
additionally have keys of type `string`, `bool`, or one of the integer types
listed above.

REZI can also handle encoding and decoding pointers to any supported type, with
any level of indirection.

#### User-Defined Types
REZI supports encoding any custom type that implements
`encoding.BinaryMarshaler`, and it supports decoding any custom type that
implements `encoding.BinaryUnmarshaler` with a pointer receiver. In fact, the
lack of built-in facilities in Go for binary encoding of user-defined types is
partially why REZI exists.

REZI does not perform any automatic inference of a user-defined type's encoding
such as what the `json` library is capable of. User-defined types that do not
implement BinaryMarshaler are not supported for encoding, and those that do not
implement BinaryUnmarshaler are not supported for decoding.

Within the `MarshalBinary` method, you can encode the data in whichever format
you wish, though these examples will have that function use REZI to encode the
members of the types. The contents of the slice that MarshalBinary returns are
completely opaque to REZI, which will consider only the slice's length. Do note
that this means that returning a nil slice or an empty but initialized slice
will both be interpreted the same by REZI and will not result in different
encodings.

```golang
// Person is an example of a user-defined type that REZI can encode and decode.
type Person struct {
    Name string
    Number int
}

func (p Person) MarshalBinary() ([]byte, error) {
    var enc []byte

    var err error
    var reziBytes []byte

    reziBytes, err = rezi.Enc(p.Name)
    if err != nil {
        return nil, fmt.Errorf("name: %w", err)
    }
    enc = append(enc, reziBytes...)

    reziBytes, err = rezi.Enc(p.Number)
    if err != nil {
        return nil, fmt.Errorf("number: %w", err)
    }
    enc = append(enc, reziBytes...)

    return enc, nil
}
```

It's always good practice to check the error value returned by Enc, but it is
worth noting that for certain values (generally, ones whose type is built-in or
consists only of built-in types), Enc will never return an error. If you know
that a value cannot possibly return an error under normal circumstances (see the
Godocs for `Enc()` to check which types that is true for), you can use `MustEnc`
to return only the bytes, which can be useful when encoding several values in
sequence directly into `append()` calls.

```golang
// this variant of MarshalBinary calls MustEnc to encode values that are
// built-in types.
func (p Person) MarshalBinary() ([]byte, error) {
    var enc []byte

    enc = append(enc, rezi.MustEnc(p.Name)...)
    enc = append(enc, rezi.MustEnc(p.Number)...)

    return enc, nil
}
```

Decoding of user-defined types is handled with the UnmarshalBinary method. The
bytes that were returned by MarshalBinary while decoding are picked up by REZI
and passed into UnmarshalBinary. Note that unlike the MarshalBinary method,
which must be defined with a value receiver for the type, REZI requires the
UnmarshalBinary to be defined with a pointer receiver.

```golang
// UnmarshalBinary takes in bytes and decodes them into a new Person object,
// which it assigns as the value of its receiver.
func (p *Person) UnmarshalBinary(data []byte) error {
    var n int
    var err error

    var decoded Person

    // decode name member
    n, err = rezi.Dec(data, &decoded.Name)
    if err != nil {
        return fmt.Errorf("name: %w", err)
    }
    data = data[n:]

    // decode number member
    n, err = rezi.Dec(data, &decoded.Number)
    if err != nil {
        return fmt.Errorf("number: %w", err)
    }
    data = data[n:]

    // everyfin was successfully decoded! assign the result as the value of this
    // Person.
    *p = decoded

    return nil
}
```

When a type has both of these methods defined, it can be encoded and decoded
with Enc and Dec just like any other type:

```golang
import (
    "fmt"

    "github.com/dekarrin/rezi"
)

...

p := Person{Name: "Terezi", Number: 612}

data, err := rezi.Enc(p)
if err != nil {
    panic(err)
}

var decoded Person

_, err := rezi.Dec(data, &decoded)
if err != nil {
    panic(err)
}

fmt.Println(decoded.Name) // "Terezi"
fmt.Println(decoded.Number) // 612
```

### Compression

REZI does not currently support compression itself but its results can be
compressed. Automatic compression will come in a future release.
