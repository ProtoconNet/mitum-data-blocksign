package digest

import (
	"github.com/soonkuk/mitum-data/blocksign"
	"github.com/soonkuk/mitum-data/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	mongodbstorage "github.com/spikeekips/mitum/storage/mongodb"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"golang.org/x/xerrors"
)

type AccountDoc struct {
	mongodbstorage.BaseDoc
	address string
	height  base.Height
}

func NewAccountDoc(rs AccountValue, enc encoder.Encoder) (AccountDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, rs, enc)
	if err != nil {
		return AccountDoc{}, err
	}

	return AccountDoc{
		BaseDoc: b,
		address: currency.StateAddressKeyPrefix(rs.ac.Address()),
		height:  rs.height,
	}, nil
}

func (doc AccountDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["address"] = doc.address
	m["height"] = doc.height

	return bsonenc.Marshal(m)
}

/*
func NewDocumentDoc(rs DocumentValue, enc encoder.Encoder) (AccountDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, rs, enc)
	if err != nil {
		return AccountDoc{}, err
	}

	return AccountDoc{
		BaseDoc: b,
		address: currency.StateAddressKeyPrefix(rs.ac.Address()),
		height:  rs.height,
	}, nil
}
*/

type BalanceDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	am currency.Amount
}

// NewBalanceDoc gets the State of Amount
func NewBalanceDoc(st state.State, enc encoder.Encoder) (BalanceDoc, error) {
	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return BalanceDoc{}, xerrors.Errorf("BalanceDoc needs Amount state: %w", err)
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return BalanceDoc{}, err
	}

	return BalanceDoc{
		BaseDoc: b,
		st:      st,
		am:      am,
	}, nil
}

func (doc BalanceDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}
	address := doc.st.Key()[:len(doc.st.Key())-len(currency.StateKeyBalanceSuffix)-len(doc.am.Currency())-1]
	m["address"] = address
	m["currency"] = doc.am.Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type DocumentDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	fh blocksign.FileHash
	id blocksign.DocId
}

// NewDocumentDoc gets the State of DocumentData
func NewDocumentDoc(st state.State, enc encoder.Encoder) (DocumentDoc, error) {

	var doc blocksign.DocumentData
	if i, err := blocksign.StateDocumentDataValue(st); err != nil {
		return DocumentDoc{}, xerrors.Errorf("DocumentDoc needs DocumentData state: %w", err)
	} else {
		doc = i
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DocumentDoc{}, err
	}
	return DocumentDoc{
		BaseDoc: b,
		st:      st,
		fh:      doc.FileHash(),
		id:      doc.DocumentId(),
	}, nil
}

func (doc DocumentDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}
	address := doc.st.Key()[:len(doc.st.Key())-len(blocksign.StateKeyDocumentDataSuffix)-len(doc.id.String())-1]
	m["address"] = address
	m["documentid"] = doc.id.Index()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
