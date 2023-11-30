# REZI

![Tests Status Badge](https://github.com/dekarrin/rezi/actions/workflows/tests.yml/badge.svg?branch=main&event=push)
[![Go Reference](https://pkg.go.dev/badge/github.com/dekarrin/rezi/v2.svg)](https://pkg.go.dev/github.com/dekarrin/rezi/v2)

The Rarefied Encoding (Compressible) for Interchange (REZI) library performs
binary marshaling of data to REZI-format bytes. It can encode and decode most
simple built-in Go types and structs that contain those types to bytes. It
further allows customization of decoding and encoding of user-defined types via
`encoding.BinaryMarshaler` or `encoding.TextMarshaler`.

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
go get -u github.com/dekarrin/rezi/v2@latest
```

And import the REZI library to use it:

```golang
import "github.com/dekarrin/rezi/v2"
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

    "github.com/dekarrin/rezi/v2"
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

    "github.com/dekarrin/rezi/v2"
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

#### Readers And Writers

You can also use REZI by creating a Reader or Writer and calling their Dec or
Enc methods respectively. This lets you read and write values directly to and
from streams of bytes.

```golang

// on the write side, get an io.Writer someWriter you want to write REZI-encoded
// data to, and write out with Enc.

w, err := rezi.NewWriter(someWriter, nil)
if err != nil {
    panic(err)
}

w.Enc(413)
w.Enc("NEPETA")
w.Enc(true)

// don't forget to call Flush or Close when done!
w.Flush()

// on the read side, get an io.Reader someReader you want to read REZI-encoded
// data from, and read it with Dec.

r, err := rezi.NewReader(someReader, nil)
if err != nil {
    panic(err)
}

var number int
var name string
var isGood bool

r.Dec(&number)
r.Dec(&name)
r.Dec(&isGood)

fmt.Println(number) // 413
fmt.Println(name)   // "NEPETA"
fmt.Println(isGood) // true
```

Output from Writer can be read in earlier versions of REZI as well with
non-Reader calls, as long as a nil or a Version 1 Format is used at startup,
without compression enabled. This does not extend to data types that didn't
exist in that version, however.

Readers created with a nil or Version 1 Format with compression disabled are
able to read data written by any prior version of REZI.

### Supported Types

REZI supports all built-in basic types. Additionally, any type that implements
`encoding.BinaryMarshaler` can be encoded, and any type that implements
`encoding.BinaryUnmarshaler` with a pointer receiver can be decoded.
Additionally, any type that implements `encoding.TextMarshaler` can be encoded,
and any type that implements `encoding.TextUnmarshaler` can be decoded. If a
type implements both sets of functions, REZI will prefer the binary marshaling
functions.

REZI supports slice, array, and map types whose values are of any supported type
(including those whose values are themselves slice, array, or map values). Maps
must additionally have keys of type `string`, `bool`, one of the built-in
integer types, or one of the built-in float types.

REZI can also handle encoding and decoding pointers to any supported type, with
any level of indirection.

On top of all of the above, REZI automatically supports any type whose
underlying type is supported, as well as any struct whose exported fields are
all of supported types.

#### Struct Support

Much like the `json` package, REZI can encode and decode most simple structs
out of the box without requiring any further customization. Simple in this case
means that all of its exported fields are a supported type. As only the exported
fields are encoded and decoded to bytes, it's okay if an unexported field is of
an unsupported type.

```golang

// AnimalInfo is fully supported; all fields will be encoded and decoded.
type AnimalInfo struct {
    Name string
    Taxonomy []string
    AverageAge int
}

// Animal is *not* supported; despite field Info being of a supported type,
// Counter is a channel, which is unsupported.
type Animal struct {
    Info AnimalInfo
    Counter chan int
}

// HiddenCounterAnimal is supported. Even though field counter is of an
// unsupported type, it is unexported and so it will be ignored.
type HiddenCounterAnimal struct {
    Info AnimalInfo
    counter chan int
}
```

Unexported fields of a struct are ignored when encoding and decoding. Any struct
that has unexported fields will keep their original values if a pointer to that
struct is passed in to be decoded.

```golang
type Player struct {
    Name string
    Classpect string

    echeladder string
}

john := Player{Name: "John Egbert", Classpect: "Heir of Breath", echeladder: "Plucky Tot"}

// the encoded bytes will only contain Name and Classpect; echeladder is not
// exported so it is ignored
data, err := rezi.Enc(john)
if err != nil {
    panic(err)
}

// now we will decode the bytes to two structs, one with the unexported member
// pre-set
var playerWithoutEcheladder Player
var playerWithEcheladder Player = Player{echeladder: "Plucky Tot"}

_, err = rezi.Dec(data, &playerWithoutEcheladder)
if err != nil {
    panic(err)
}
_, err = rezi.Dec(data, &playerWithEcheladder)
if err != nil {
    panic(err)
}

fmt.Println(playerWithoutEcheladder.echeladder) // ""
fmt.Println(playerWithEcheladder.echeladder) // "Plucky Tot"
```

Embedded structs within structs are supported if the embedded struct type is
exported; this is because it will be turned into a field with the same name as
the embedded type, and if it is exported, the field name will correspondingly
be exported. Likewise, embedded structs whose type is unexported will be ignored
during encoding and will not be encoded to.

```golang
type InternalRecord struct {
    ID int
    Location string
}

type secret struct {
    BigSecret string
}

// All fields of Employee will be marshaled and unmarshaled to; InternalRecord
// is exported
type Employee struct {
    InternalRecord
    Name string
}

// Only Name will be encoded and decoded to; secret is an unexported type.
type KeyData struct {
    secret
    Name string
}
```

If any of the above limitations are a concern, you can customize the encoding of
user-defined types by implementing one of the marshaler types
`encoding.BinaryMarshaler` or `encoding.TextMarshaler` (and their corresponding
unmarshler interfaces for decoding) as described in the next section.

### Customizing Encoding And Decoding

REZI supports encoding any custom type that implements
`encoding.BinaryMarshaler`, and it supports decoding any custom type that
implements `encoding.BinaryUnmarshaler` with a pointer receiver. In fact, the
lack of built-in facilities in Go for binary encoding of user-defined types is
partially why REZI exists.

REZI can perform automatic inference of a user-defined struct type's encoding,
similar to what the `json` library is capable of. User-defined types that
do not implement BinaryMarshaler or TextMarshaler are supported for encoding if
their underlying type is one supported by REZI, or if it is a struct type, if
all of its exported fields are supported, and vice-versa for decoding.

Within the `MarshalBinary` method, you can customize encoding the data to
whichever format you wish, though these examples will have that function use
REZI to encode the members of the types. The contents of the slice that
MarshalBinary returns are completely opaque to REZI, which will consider only
the slice's length. Do note that this means that returning a nil slice or an
empty but initialized slice will both be interpreted the same by REZI and will
not result in different encodings.

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

Custom decoding of user-defined types is handled with the UnmarshalBinary
method. The bytes that were returned by MarshalBinary while decoding are picked
up by REZI and passed into UnmarshalBinary. Note that unlike the MarshalBinary
method, which must be defined with a value receiver for the type, REZI requires
the UnmarshalBinary to be defined with a pointer receiver.

```golang
// UnmarshalBinary takes in bytes and decodes them into a new Person object,
// which it assigns as the value of its receiver.
func (p *Person) UnmarshalBinary(data []byte) error {
    var n int
    var offset int
    var err error

    var decoded Person

    // decode name member
    n, err = rezi.Dec(data[offset:], &decoded.Name)
    if err != nil {
        return fmt.Errorf("name: %w", err)
    }
    offset += n

    // decode number member
    n, err = rezi.Dec(data[offset:], &decoded.Number)
    if err != nil {
        return fmt.Errorf("number: %w", err)
    }
    offset += n

    // everyfin was successfully decoded! assign the result as the value of this
    // Person.
    *p = decoded

    return nil
}
```

REZI decoding supports reporting byte offsets where an error occurred in the
supplied data. In order to support this in user-defined types, Wrapf can be used
to wrap an error returned from REZI functions and give the offset into the data
that the REZI function was called on. This offset will be combined with any
inside the REZI error to give the complete offset:

```golang
// a typical use of Wrapf within an UnmarshalBinary method:

n, err = rezi.Dec(data[offset:], &decoded.Name)
if err != nil {
    // Always make sure to use %s or %v in Wrapf, never %w!
    return rezi.Wrapf(offset, "name: %s", err)
}
offset += n

// Additionally, first arg to the format string must always be the error
// returned from the REZI function.
```

When a type has both the `UnmarshalBinary` and `MarshalBinary` methods defined,
it can be encoded and decoded with Enc and Dec just like any other type:

```golang
import (
    "fmt"

    "github.com/dekarrin/rezi/v2"
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

REZI supports compression via the use of Reader and Writer. When one is created,
instead of giving a nil value for the Format it accepts, pass in a Format with
Compression set to true.

```golang

w, err := rezi.NewWriter(someWriter, &rezi.Format{Compression: true})
if err != nil {
    panic(err)
}

w.Enc(413)

// don't forget to call Flush or Close when done
w.Flush()

// on the read side, get an io.Reader someReader you want to read REZI-encoded
// data from, and pass it to NewReader along with a Format that enables
// Compression.

r, err := rezi.NewReader(someReader, &rezi.Format{Compression: true})
if err != nil {
    panic(err)
}

var number int

r.Dec(&number)

fmt.Println(number) // 413
```