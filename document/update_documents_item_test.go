//go:build test
// +build test

package document

import (
	"testing"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/stretchr/testify/suite"
)

type testUpdateDocumentsItemImpl struct {
	baseTest
}

func (t *testUpdateDocumentsItemImpl) TestNewUpdateDocumentsItem() {
	bsDocID := "1sdi"
	bcUserDocID := "1cui"
	bcLandDocID := "1cli"
	bcVotingDocID := "1cvi"
	bcHistoryDocID := "1chi"
	// create BSDocData
	bsDocData, ownerAccount, signerAccount := newBSDocData("filehash", bsDocID, account{})
	// create BCUserData
	bcUserData, _, stat := newBCUserData(bcUserDocID, *ownerAccount)
	// create BCLandData
	bcLandData, _, renterAccount := newBCLandData(bcLandDocID, *ownerAccount)
	// create BCVotingData
	bcVotingData, _, bossAccount := newBCVotingData(bcVotingDocID, *ownerAccount)
	// create BCHistoryData
	bcHistoryData, _, cityAdminAccount := newBCHistoryData(bcHistoryDocID, *ownerAccount)
	// currency id
	cid := currency.CurrencyID("SHOWME")
	// create document item
	bsDocDataItem := NewUpdateDocumentsItemImpl(*bsDocData, cid)
	bcUserDatatItem := NewUpdateDocumentsItemImpl(*bcUserData, cid)
	bcLandDatatItem := NewUpdateDocumentsItemImpl(*bcLandData, cid)
	bcVotingDatatItem := NewUpdateDocumentsItemImpl(*bcVotingData, cid)
	bcHistoryDatatItem := NewUpdateDocumentsItemImpl(*bcHistoryData, cid)

	// compare filedata from updated item's BSDocData with original filedata
	doc0, _ := bsDocDataItem.doc.(BSDocData)
	t.Equal(MustNewBSDocID(bsDocID), doc0.info.id)
	t.Equal(BSDocDataType, doc0.info.docType)
	t.Equal(currency.NewBig(100), doc0.size)
	t.Equal(FileHash("filehash"), doc0.fileHash)
	t.Equal(MustNewDocSign(ownerAccount.Address, "signcode0", true), doc0.creator)
	t.Equal("title", doc0.title)
	t.Equal(MustNewDocSign(signerAccount.Address, "signcode1", false), doc0.signers[0])

	// compare filedata from updated item's BCUserData with original filedata
	doc1, _ := bcUserDatatItem.doc.(BCUserData)
	t.Equal(MustNewBCUserDocID(bcUserDocID), doc1.info.id)
	t.Equal(BCUserDataType, doc1.info.docType)
	t.Equal(uint(10), doc1.gold)
	t.Equal(uint(10), doc1.bankgold)
	t.Equal(stat, doc1.statistics)

	// compare filedata from updated BCLandData's fact with original filedata
	doc2, _ := bcLandDatatItem.doc.(BCLandData)
	t.Equal(MustNewBCLandDocID(bcLandDocID), doc2.info.id)
	t.Equal(BCLandDataType, doc2.info.docType)
	t.Equal(renterAccount.Address, doc2.account)
	t.Equal("address", doc2.address)
	t.Equal("area", doc2.area)
	t.Equal(uint(10), doc2.periodday)
	t.Equal("rentdate", doc2.rentdate)
	t.Equal("renter", doc2.renter)

	// compare filedata from updated BCVotingData's fact with original filedata
	doc3, _ := bcVotingDatatItem.doc.(BCVotingData)
	t.Equal(MustNewBCVotingDocID(bcVotingDocID), doc3.info.id)
	t.Equal(BCVotingDataType, doc3.info.docType)
	t.Equal(bossAccount.Address, doc3.account)
	t.Equal("bossname", doc3.bossname)
	t.Equal([]VotingCandidate{MustNewVotingCandidate(bossAccount.Address, "nickname", "manifest", uint(10))}, doc3.candidates)

	// compare filedata from updated BCHistoryData's fact with original filedata
	doc4, _ := bcHistoryDatatItem.doc.(BCHistoryData)
	t.Equal(MustNewBCHistoryDocID(bcHistoryDocID), doc4.info.id)
	t.Equal(BCHistoryDataType, doc4.info.docType)
	t.Equal(cityAdminAccount.Address, doc4.account)
	t.Equal("application", doc4.application)
	t.Equal("date", doc4.date)
	t.Equal("name", doc4.name)
	t.Equal("usage", doc4.usage)
}

func (t *testUpdateDocumentsItemImpl) TestInvaliDocumentType() {
	bsDocID := "1sdi"
	bcUserDocID := "1cui"
	bcLandDocID := "1cli"
	bcVotingDocID := "1cvi"
	bcHistoryDocID := "1chi"

	// create BSDocData
	bsDocData, ownerAccount, _ := newBSDocData("filehash", bsDocID, account{})
	// create BCUserData
	bcUserData, _, _ := newBCUserData(bcUserDocID, *ownerAccount)
	// create BCLandData
	bcLandData, _, _ := newBCLandData(bcLandDocID, *ownerAccount)
	// create BCVotingData
	bcVotingData, _, _ := newBCVotingData(bcVotingDocID, *ownerAccount)
	// create BCHistoryData
	bcHistoryData, _, _ := newBCHistoryData(bcHistoryDocID, *ownerAccount)
	// set unmatched document type
	bsDocData.info.docType = BCUserDataType
	// set unmatched document type
	bcUserData.info.docType = BCLandDataType
	// set unmatched document type
	bcLandData.info.docType = BCVotingDataType
	// set unmatched document type
	bcVotingData.info.docType = BCHistoryDataType
	// set unmatched document type
	bcHistoryData.info.docType = BSDocDataType
	// currency id
	cid := currency.CurrencyID("SHOWME")
	// create document item
	cd := NewUpdateDocumentsItemImpl(*bsDocData, cid)
	err := cd.IsValid(nil)
	t.Contains(err.Error(), "docInfo not matched with DocumentData Type")
	cd = NewUpdateDocumentsItemImpl(*bcUserData, cid)
	err = cd.IsValid(nil)
	t.Contains(err.Error(), "docInfo not matched with DocumentData Type")
	cd = NewUpdateDocumentsItemImpl(*bcLandData, cid)
	err = cd.IsValid(nil)
	t.Contains(err.Error(), "docInfo not matched with DocumentData Type")
	cd = NewUpdateDocumentsItemImpl(*bcVotingData, cid)
	err = cd.IsValid(nil)
	t.Contains(err.Error(), "docInfo not matched with DocumentData Type")
	cd = NewUpdateDocumentsItemImpl(*bcHistoryData, cid)
	err = cd.IsValid(nil)
	t.Contains(err.Error(), "docInfo not matched with DocumentData Type")
}

func TestUpdateDocumentsItemImpl(t *testing.T) {
	suite.Run(t, new(testUpdateDocumentsItemImpl))
}

func testUpdateDocumentsItemImplEncode(enc encoder.Encoder) suite.TestingSuite {
	t := new(baseTestOperationItemEncode)

	t.enc = enc
	t.newObject = func() interface{} {
		bsDocID := "1sdi"
		bsDocData, _, _ := newBSDocData("filehash", bsDocID, account{})
		cid := currency.CurrencyID("SHOWME")
		cd := NewUpdateDocumentsItemImpl(*bsDocData, cid)

		return cd
	}

	t.compare = func(a, b interface{}) {
		da := a.(UpdateDocumentsItem)
		db := b.(UpdateDocumentsItem)

		t.Equal(da.Hint(), db.Hint())
		t.Equal(da.Doc(), db.Doc())
		t.Equal(da.Currency(), db.Currency())
	}

	return t
}

func TestUpdateDocumentsItemImplEncodeJSON(t *testing.T) {
	suite.Run(t, testUpdateDocumentsItemImplEncode(jsonenc.NewEncoder()))
}

func TestUpdateDocumentsItemImplEncodeBSON(t *testing.T) {
	suite.Run(t, testUpdateDocumentsItemImplEncode(bsonenc.NewEncoder()))
}
