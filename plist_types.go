package plist

import (
	"hash/crc32"
	"strconv"
	"time"
)

// magic value used in the non-binary encoding of UIDs
// (stored as a dictionary mapping CF$UID->integer)
const cfUIDMagic = "CF$UID"

type cfValue interface {
	typeName() string
	hash() interface{}
}

type cfDictionary struct {
	keys   []string
	values []cfValue
}

func (*cfDictionary) typeName() string {
	return "dictionary"
}

func (p *cfDictionary) hash() interface{} {
	return p
}

func (p *cfDictionary) maybeUID(lax bool) cfValue {
	if len(p.keys) == 1 && p.keys[0] == "CF$UID" && len(p.values) == 1 {
		pval := p.values[0]
		if integer, ok := pval.(*cfNumber); ok {
			return cfUID(integer.value)
		}
		// Openstep only has cfString. Act like the unmarshaller a bit.
		if lax {
			if str, ok := pval.(cfString); ok {
				if i, err := strconv.ParseUint(string(str), 10, 64); err == nil {
					return cfUID(i)
				}
			}
		}
	}
	return p
}

type cfArray struct {
	values []cfValue
}

func (*cfArray) typeName() string {
	return "array"
}

func (p *cfArray) hash() interface{} {
	return p
}

type cfString string

func (cfString) typeName() string {
	return "string"
}

func (p cfString) hash() interface{} {
	return string(p)
}

type cfNumber struct {
	signed bool
	value  uint64
}

func (*cfNumber) typeName() string {
	return "integer"
}

func (p *cfNumber) hash() interface{} {
	if p.signed {
		return int64(p.value)
	}
	return p.value
}

type cfReal struct {
	wide  bool
	value float64
}

func (cfReal) typeName() string {
	return "real"
}

func (p *cfReal) hash() interface{} {
	if p.wide {
		return p.value
	}
	return float32(p.value)
}

type cfBoolean bool

func (cfBoolean) typeName() string {
	return "boolean"
}

func (p cfBoolean) hash() interface{} {
	return bool(p)
}

type cfUID UID

func (cfUID) typeName() string {
	return "UID"
}

func (p cfUID) hash() interface{} {
	return p
}

func (p cfUID) toDict() *cfDictionary {
	return &cfDictionary{
		keys: []string{cfUIDMagic},
		values: []cfValue{&cfNumber{
			signed: false,
			value:  uint64(p),
		}},
	}
}

type cfData []byte

func (cfData) typeName() string {
	return "data"
}

func (p cfData) hash() interface{} {
	// Data are uniqued by their checksums.
	// Todo: Look at calculating this only once and storing it somewhere;
	// crc32 is fairly quick, however.
	return crc32.ChecksumIEEE([]byte(p))
}

type cfDate time.Time

func (cfDate) typeName() string {
	return "date"
}

func (p cfDate) hash() interface{} {
	return time.Time(p)
}
