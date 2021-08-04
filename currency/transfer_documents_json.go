package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/valuehash"
)

type TransferDocumentsFactJSONPacker struct {
	jsonenc.HintedHead
	H  valuehash.Hash          `json:"hash"`
	TK []byte                  `json:"token"`
	SD base.Address            `json:"sender"`
	IT []TransferDocumentsItem `json:"items"`
}

func (fact TransferDocumentsFact) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(TransferDocumentsFactJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fact.Hint()),
		H:          fact.h,
		TK:         fact.token,
		SD:         fact.sender,
		IT:         fact.items,
	})
}

func (fact *TransferDocumentsFact) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufact struct {
		H  valuehash.Bytes     `json:"hash"`
		TK []byte              `json:"token"`
		SD base.AddressDecoder `json:"sender"`
		IT []json.RawMessage   `json:"items"`
	}
	if err := jsonenc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	its := make([][]byte, len(ufact.IT))
	for i := range ufact.IT {
		its[i] = ufact.IT[i]
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, its)
}

func (op TransferDocuments) MarshalJSON() ([]byte, error) {
	m := op.BaseOperation.JSONM()
	m["memo"] = op.Memo

	return jsonenc.Marshal(m)
}

func (op *TransferDocuments) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo operation.BaseOperation
	if err := ubo.UnpackJSON(b, enc); err != nil {
		return err
	}

	*op = TransferDocuments{BaseOperation: ubo}

	var um MemoJSONUnpacker
	if err := enc.Unmarshal(b, &um); err != nil {
		return err
	} else {
		op.Memo = um.Memo
	}

	return nil
}
