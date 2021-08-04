package digest

import (
	"github.com/spikeekips/mitum/base/state"
	mongodbstorage "github.com/spikeekips/mitum/storage/mongodb"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/xerrors"
)

func loadOperationHash(decoder func(interface{}) error) (valuehash.Hash, error) {
	var doc struct {
		FH valuehash.Bytes `bson:"fact"`
	}

	if err := decoder(&doc); err != nil {
		return nil, err
	}
	return doc.FH, nil
}

func loadOperation(decoder func(interface{}) error, encs *encoder.Encoders) (OperationValue, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return OperationValue{}, err
	}

	if _, hinter, err := mongodbstorage.LoadDataFromDoc(b, encs); err != nil {
		return OperationValue{}, err
	} else if va, ok := hinter.(OperationValue); !ok {
		return OperationValue{}, xerrors.Errorf("not OperationValue: %T", hinter)
	} else {
		return va, nil
	}
}

func loadAccountValue(decoder func(interface{}) error, encs *encoder.Encoders) (AccountValue, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return AccountValue{}, err
	}

	if _, hinter, err := mongodbstorage.LoadDataFromDoc(b, encs); err != nil {
		return AccountValue{}, err
	} else if rs, ok := hinter.(AccountValue); !ok {
		return AccountValue{}, xerrors.Errorf("not AccountValue: %T", hinter)
	} else {
		return rs, nil
	}
}

func loadBalance(decoder func(interface{}) error, encs *encoder.Encoders) (state.State, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return nil, err
	}

	if _, hinter, err := mongodbstorage.LoadDataFromDoc(b, encs); err != nil {
		return nil, err
	} else if st, ok := hinter.(state.State); !ok {
		return nil, xerrors.Errorf("not currency.Big: %T", hinter)
	} else {
		return st, nil
	}
}

/*
func loadFileData(decoder func(interface{}) error, encs *encoder.Encoders) (state.State, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return nil, err
	}

	if _, hinter, err := mongodbstorage.LoadDataFromDoc(b, encs); err != nil {
		return nil, err
	} else if st, ok := hinter.(state.State); !ok {
		return nil, xerrors.Errorf("filedata load error when cast hinter to state.State : %T", hinter)
	} else {
		return st, nil
	}
}

func loadDocumentValue(decoder func(interface{}) error, encs *encoder.Encoders) (DocumentValue, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return DocumentValue{}, err
	}

	if _, hinter, err := mongodbstorage.LoadDataFromDoc(b, encs); err != nil {
		return DocumentValue{}, err
	} else if rs, ok := hinter.(DocumentValue); !ok {
		return DocumentValue{}, xerrors.Errorf("not DocumentValue: %T", hinter)
	} else {
		return rs, nil
	}
}
*/
