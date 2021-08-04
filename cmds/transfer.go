package cmds

import (
	"bytes"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/seal"
	"github.com/spikeekips/mitum/util"

	currency "github.com/soonkuk/mitum-data/currency"
)

type TransferCommand struct {
	*BaseCommand
	OperationFlags
	Sender   AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Receiver AddressFlag    `arg:"" name:"receiver" help:"receiver address" required:"true"`
	Currency CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	Big      BigFlag        `arg:"" name:"big" help:"big to send" required:"true"`
	Seal     FileLoad       `help:"seal" optional:""`
	sender   base.Address
	receiver base.Address
}

func NewTransferCommand() TransferCommand {
	return TransferCommand{
		BaseCommand: NewBaseCommand("transfer-operation"),
	}
}

func (cmd *TransferCommand) Run(version util.Version) error {
	if err := cmd.Initialize(cmd, version); err != nil {
		return xerrors.Errorf("failed to initialize command: %w", err)
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	sl, err := loadSealAndAddOperation(
		cmd.Seal.Bytes(),
		cmd.Privatekey,
		cmd.NetworkID.NetworkID(),
		op,
	)
	if err != nil {
		return err
	}
	cmd.pretty(cmd.Pretty, sl)

	return nil
}

func (cmd *TransferCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if sender, err := cmd.Sender.Encode(jenc); err != nil {
		return xerrors.Errorf("invalid sender format, %q: %w", cmd.Sender.String(), err)
	} else if receiver, err := cmd.Receiver.Encode(jenc); err != nil {
		return xerrors.Errorf("invalid sender format, %q: %w", cmd.Sender.String(), err)
	} else {
		cmd.sender = sender
		cmd.receiver = receiver
	}

	return nil
}

func (cmd *TransferCommand) createOperation() (operation.Operation, error) { // nolint:dupl
	i, err := loadOperations(cmd.Seal.Bytes(), cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}

	var items []currency.TransfersItem
	for j := range i {
		if t, ok := i[j].(currency.Transfers); ok {
			items = t.Fact().(currency.TransfersFact).Items()
		}
	}

	am := currency.NewAmount(cmd.Big.Big, cmd.Currency.CID)
	if err = am.IsValid(nil); err != nil {
		return nil, err
	}

	item := currency.NewTransfersItemSingleAmount(cmd.receiver, am)
	if err = item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := currency.NewTransfersFact([]byte(cmd.Token), cmd.sender, items)

	var fs []operation.FactSign
	sig, err := operation.NewFactSignature(cmd.Privatekey, fact, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}
	fs = append(fs, operation.NewBaseFactSign(cmd.Privatekey.Publickey(), sig))

	op, err := currency.NewTransfers(fact, fs, cmd.Memo)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transfers operation: %w", err)
	}
	return op, nil
}

func loadOperations(b []byte, networkID base.NetworkID) ([]operation.Operation, error) {
	if len(bytes.TrimSpace(b)) < 1 {
		return nil, nil
	}

	var sl seal.Seal
	if s, err := loadSeal(b, networkID); err != nil {
		return nil, err
	} else if so, ok := s.(operation.Seal); !ok {
		return nil, xerrors.Errorf("seal is not operation.Seal, %T", s)
	} else if _, ok := so.(operation.SealUpdater); !ok {
		return nil, xerrors.Errorf("seal is not operation.SealUpdater, %T", s)
	} else {
		sl = so
	}

	return sl.(operation.Seal).Operations(), nil
}
