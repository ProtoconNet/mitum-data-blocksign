package currency

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/logging"
)

var (
	AddressType = hint.Type("mca")
	AddressHint = hint.NewHint(AddressType, "v0.0.1")
)

var EmptyAddress = Address("")

type Address string

func NewAddress(name string) (Address, error) {
	ca := Address(name)

	return ca, ca.IsValid(nil)
}

func NewAddressFromKeys(keys Keys) (Address, error) {
	if err := keys.IsValid(nil); err != nil {
		return EmptyAddress, err
	}

	return NewAddress(keys.Hash().String())
}

func (ca Address) Raw() string {
	return string(ca)
}

func (ca Address) String() string {
	return hint.NewHintedString(ca.Hint(), string(ca)).String()
}

func (Address) Hint() hint.Hint {
	return AddressHint
}

func (ca Address) IsValid([]byte) error {
	if s := strings.TrimSpace(ca.String()); len(s) < 1 {
		return xerrors.Errorf("empty address")
	}

	return nil
}

func (ca Address) Equal(a base.Address) bool {
	if ca.Hint().Type() != a.Hint().Type() {
		return false
	}

	return ca == a.(Address)
}

func (ca Address) Bytes() []byte {
	return []byte(ca.String())
}

func (ca Address) MarshalText() ([]byte, error) {
	return []byte(ca.String()), nil
}

func (ca *Address) UnmarshalText(b []byte) error {
	a, err := NewAddress(string(b))
	if err != nil {
		return err
	}
	*ca = a

	return nil
}

func (ca Address) MarshalLog(key string, e logging.Emitter, _ bool) logging.Emitter {
	return e.Str(key, ca.String())
}

type Addresses interface {
	Addresses() ([]base.Address, error)
}
