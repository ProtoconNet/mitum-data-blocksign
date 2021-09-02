package digest

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (va OperationValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(va.Hint()),
		bson.M{
			"op":           va.op,
			"height":       va.height,
			"confirmed_at": va.confirmedAt,
			"in_state":     va.inState,
			"reason":       va.reason,
			"index":        va.index,
		},
	))
}

type OperationValueBSONUnpacker struct {
	OP bson.Raw    `bson:"op"`
	HT base.Height `bson:"height"`
	CT time.Time   `bson:"confirmed_at"`
	IN bool        `bson:"in_state"`
	RS bson.Raw    `bson:"reason"`
	ID uint64      `bson:"index"`
}

func (va *OperationValue) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uva OperationValueBSONUnpacker
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	if hinter, err := enc.Decode(uva.OP); err != nil {
		return err
	} else if op, ok := hinter.(operation.Operation); !ok {
		return errors.Errorf("not operation.Operation: %T", hinter)
	} else {
		va.op = op
	}

	i, err := operation.DecodeReasonError(uva.RS, enc)
	if err != nil {
		return err
	}
	va.reason = i

	va.height = uva.HT
	va.confirmedAt = uva.CT
	va.inState = uva.IN
	va.index = uva.ID

	return nil
}
