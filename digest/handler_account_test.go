// +build mongodb

package digest

import (
	"fmt"
	"io"
	"sort"
	"testing"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/localtime"
	"github.com/stretchr/testify/suite"
)

type testHandlerAccount struct {
	baseTestHandlers
}

func (t *testHandlerAccount) TestAccount() {
	st, _ := t.Database()

	ac := t.newAccount()
	height := base.Height(33)

	am := currency.MustNewAmount(t.randomBig(), t.cid)

	va, _ := t.insertAccount(st, height, ac, am)

	handlers := t.handlers(st, DummyCache{})

	self, err := handlers.router.Get(HandlerPathAccount).URLPath("address", ac.Address().String())
	t.NoError(err)

	blockLink, err := handlers.router.Get(HandlerPathBlockByHeight).URLPath("height", va.Height().String())
	t.NoError(err)
	previousBlockLink, err := handlers.router.Get(HandlerPathBlockByHeight).URLPath("height", va.PreviousHeight().String())
	t.NoError(err)

	w := t.requestOK(handlers, "GET", self.Path, nil)

	b, err := io.ReadAll(w.Result().Body)
	t.NoError(err)

	hal := t.loadHal(b)

	t.Equal(self.String(), hal.Links()["self"].Href())
	t.Equal(blockLink.Path, hal.Links()["block"].Href())
	t.Equal(previousBlockLink.Path, hal.Links()["previous_block"].Href())

	hinter, err := t.JSONEnc.Decode(hal.RawInterface())
	t.NoError(err)
	uva, ok := hinter.(AccountValue)
	t.True(ok)

	t.compareAccountValue(va, uva)
}

func (t *testHandlerAccount) TestAccountNotFound() {
	st, _ := t.Database()

	handlers := t.handlers(st, DummyCache{})

	unknown, err := currency.NewAddress(util.UUID().String())
	t.NoError(err)

	self, err := handlers.router.Get(HandlerPathAccount).URLPath("address", unknown.String())
	t.NoError(err)

	w := t.request404(handlers, "GET", self.Path, nil)

	b, err := io.ReadAll(w.Result().Body)
	t.NoError(err)

	var problem Problem
	t.NoError(jsonenc.Unmarshal(b, &problem))
	t.Contains(problem.Error(), "not found")
}

func (t *testHandlerAccount) TestAccountOperations() {
	st, _ := t.Database()

	sender := currency.MustAddress(util.UUID().String())

	var offsets []string
	offsetByHashes := map[string]string{}
	hashesByOffset := map[string]string{}

	for i := 0; i < 10; i++ {
		height := base.Height(i % 3)
		index := uint64(i)
		tf := t.newTransfer(sender, currency.MustAddress(util.UUID().String()))
		doc, err := NewOperationDoc(tf, t.BSONEnc, height, localtime.UTCNow(), true, nil, index)
		t.NoError(err)
		_ = t.insertDoc(st, defaultColNameOperation, doc)

		fh := tf.Fact().Hash().String()
		offset := buildOffset(height, index)
		offsets = append(offsets, offset)
		offsetByHashes[fh] = offset
		hashesByOffset[offset] = fh
	}

	var hashes []string
	sort.Strings(offsets)
	for _, o := range offsets {
		hashes = append(hashes, hashesByOffset[o])
	}

	var limit int64 = 3
	handlers := t.handlers(st, DummyCache{})
	_ = handlers.SetLimiter(func(string) int64 {
		return limit
	})

	self, err := handlers.router.Get(HandlerPathAccountOperations).URLPath("address", sender.String())
	t.NoError(err)

	next, err := handlers.router.Get(HandlerPathAccountOperations).URLPath("address", sender.String())
	t.NoError(err)
	next.RawQuery = stringOffsetQuery(offsetByHashes[hashes[limit-1]])

	w := t.requestOK(handlers, "GET", self.Path, nil)

	b, err := io.ReadAll(w.Result().Body)
	t.NoError(err)

	hal := t.loadHal(b)

	// NOTE check self link
	t.Equal(self.String(), hal.Links()["self"].Href())

	// NOTE check next link
	t.Equal(next.String(), hal.Links()["next"].Href())

	var em []BaseHal
	t.NoError(jsonenc.Unmarshal(hal.RawInterface(), &em))
	t.Equal(int(limit), len(em))
}

func (t *testHandlerAccount) TestAccountOperationsPaging() {
	st, _ := t.Database()

	sender := currency.MustAddress(util.UUID().String())
	var offsets []string
	hashesByOffset := map[string]string{}

	for i := 0; i < 10; i++ {
		height := base.Height(i % 3)
		index := uint64(i)
		tf := t.newTransfer(sender, currency.MustAddress(util.UUID().String()))
		doc, err := NewOperationDoc(tf, t.BSONEnc, height, localtime.UTCNow(), true, nil, index)
		t.NoError(err)
		_ = t.insertDoc(st, defaultColNameOperation, doc)

		fh := tf.Fact().Hash().String()

		offset := buildOffset(height, index)
		offsets = append(offsets, offset)
		hashesByOffset[offset] = fh
	}

	var limit int64 = 3
	handlers := t.handlers(st, DummyCache{})
	_ = handlers.SetLimiter(func(string) int64 {
		return limit
	})

	{ // no reverse
		var hashes []string
		sort.Strings(offsets)
		for _, o := range offsets {
			hashes = append(hashes, hashesByOffset[o])
		}

		reverse := false
		offset := ""

		self, err := handlers.router.Get(HandlerPathAccountOperations).URLPath("address", sender.String())
		t.NoError(err)
		self.RawQuery = fmt.Sprintf("%s&%s", stringOffsetQuery(offset), stringBoolQuery("reverse", reverse))

		var uhashes []string
		for {
			w := t.requestOK(handlers, "GET", self.String(), nil)

			b, err := io.ReadAll(w.Result().Body)
			t.NoError(err)

			hal := t.loadHal(b)

			var em []BaseHal
			t.NoError(jsonenc.Unmarshal(hal.RawInterface(), &em))
			t.True(int(limit) >= len(em))

			for _, b := range em {
				hinter, err := t.JSONEnc.Decode(b.RawInterface())
				t.NoError(err)
				va := hinter.(OperationValue)

				fh := va.Operation().Fact().Hash().String()
				uhashes = append(uhashes, fh)
			}

			next, err := hal.Links()["next"].URL()
			t.NoError(err)
			self = next

			if int64(len(em)) < limit {
				break
			}
		}

		t.Equal(hashes, uhashes)
	}

	{ // reverse
		var hashes []string
		sort.Sort(sort.Reverse(sort.StringSlice(offsets)))

		for _, o := range offsets {
			hashes = append(hashes, hashesByOffset[o])
		}

		reverse := true
		offset := ""

		self, err := handlers.router.Get(HandlerPathAccountOperations).URLPath("address", sender.String())
		t.NoError(err)
		self.RawQuery = fmt.Sprintf("%s&%s", stringOffsetQuery(offset), stringBoolQuery("reverse", reverse))

		var uhashes []string
		for {
			w := t.requestOK(handlers, "GET", self.String(), nil)

			b, err := io.ReadAll(w.Result().Body)
			t.NoError(err)

			hal := t.loadHal(b)

			var em []BaseHal
			t.NoError(jsonenc.Unmarshal(hal.RawInterface(), &em))
			t.True(int(limit) >= len(em))

			for _, b := range em {
				hinter, err := t.JSONEnc.Decode(b.RawInterface())
				t.NoError(err)
				va := hinter.(OperationValue)

				fh := va.Operation().Fact().Hash().String()
				uhashes = append(uhashes, fh)
			}

			next, err := hal.Links()["next"].URL()
			t.NoError(err)
			self = next

			if int64(len(em)) < limit {
				break
			}
		}

		t.Equal(hashes, uhashes)
	}
}

func (t *testHandlerAccount) TestAccountOperationsPagingOverOffset() {
	st, _ := t.Database()

	sender := currency.MustAddress(util.UUID().String())

	var hashes []string
	for i := 0; i < 10; i++ {
		height := base.Height(i % 3)
		tf := t.newTransfer(sender, currency.MustAddress(util.UUID().String()))
		doc, err := NewOperationDoc(tf, t.BSONEnc, height, localtime.UTCNow(), true, nil, uint64(i))
		t.NoError(err)
		_ = t.insertDoc(st, defaultColNameOperation, doc)

		hashes = append(hashes, tf.Fact().Hash().String())
	}

	sort.Sort(sort.Reverse(sort.StringSlice(hashes)))

	var limit int64 = 3
	handlers := t.handlers(st, DummyCache{})
	_ = handlers.SetLimiter(func(string) int64 {
		return limit
	})

	offset := buildOffset(base.Height(9), uint64(20))

	self, err := handlers.router.Get(HandlerPathAccountOperations).URLPath("address", sender.String())
	self.RawQuery = fmt.Sprintf("%s&", stringOffsetQuery(offset))
	t.NoError(err)

	w := t.request404(handlers, "GET", self.String(), nil)

	b, err := io.ReadAll(w.Result().Body)
	t.NoError(err)

	var problem Problem
	t.NoError(jsonenc.Unmarshal(b, &problem))
	t.Contains(problem.Error(), "operations not found")
}

func TestHandlerAccount(t *testing.T) {
	suite.Run(t, new(testHandlerAccount))
}
