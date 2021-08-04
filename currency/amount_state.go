package currency

import (
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	AmountStateType = hint.Type("mitum-currency-amount-state")
	AmountStateHint = hint.NewHint(AmountStateType, "v0.0.1")
)

type AmountState struct {
	state.State
	cid CurrencyID
	add Big
	fee Big
}

func NewAmountState(st state.State, cid CurrencyID) AmountState {
	if sst, ok := st.(AmountState); ok {
		return sst
	}

	return AmountState{
		State: st,
		cid:   cid,
		add:   ZeroBig,
		fee:   ZeroBig,
	}
}

func (AmountState) Hint() hint.Hint {
	return AmountStateHint
}

func (st AmountState) IsValid(b []byte) error {
	if err := st.State.IsValid(b); err != nil {
		return err
	}

	if !st.fee.OverNil() {
		return xerrors.Errorf("invalid fee; under zero, %v", st.fee)
	}

	return nil
}

func (st AmountState) Bytes() []byte {
	return util.ConcatBytesSlice(
		st.State.Bytes(),
		st.fee.Bytes(),
	)
}

func (st AmountState) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(st.Bytes())
}

func (st AmountState) Merge(b state.State) (state.State, error) {
	var am Amount
	if b, err := StateBalanceValue(b); err != nil {
		if !xerrors.Is(err, util.NotFoundError) {
			return nil, err
		}
		am = NewZeroAmount(st.cid)
	} else {
		am = b
	}
	// 수수료 처리를 위해서 AmountState의 fee에 수수료는 더해주고
	// state.state의 value에 들어갈 amount에 AmountState의 add값을 더해주는데
	// add 값은 +가 될 수도 있고 -가 될 수도 있다.
	// +면 가감이 되고 -면 차감이 된다.
	return SetStateBalanceValue(
		st.AddFee(b.(AmountState).fee),
		am.WithBig(am.Big().Add(st.add)),
	)
}

func (st AmountState) Currency() CurrencyID {
	return st.cid
}

func (st AmountState) Fee() Big {
	return st.fee
}

func (st AmountState) AddFee(fee Big) AmountState {
	st.fee = st.fee.Add(fee)

	return st
}

func (st AmountState) Add(a Big) AmountState {
	st.add = st.add.Add(a)

	return st
}

func (st AmountState) Sub(a Big) AmountState {
	st.add = st.add.Sub(a)

	return st
}

func (st AmountState) SetValue(v state.Value) (state.State, error) {
	s, err := st.State.SetValue(v)
	if err != nil {
		return nil, err
	}
	st.State = s

	return st, nil
}

func (st AmountState) SetHash(h valuehash.Hash) (state.State, error) {
	s, err := st.State.SetHash(h)
	if err != nil {
		return nil, err
	}
	st.State = s

	return st, nil
}

func (st AmountState) SetHeight(h base.Height) state.State {
	st.State = st.State.SetHeight(h)

	return st
}

func (st AmountState) SetPreviousHeight(h base.Height) (state.State, error) {
	s, err := st.State.SetPreviousHeight(h)
	if err != nil {
		return nil, err
	}
	st.State = s

	return st, nil
}

func (st AmountState) SetOperation(ops []valuehash.Hash) state.State {
	st.State = st.State.SetOperation(ops)

	return st
}

func (st AmountState) Clear() state.State {
	st.State = st.State.Clear()

	st.add = ZeroBig
	st.fee = ZeroBig

	return st
}
