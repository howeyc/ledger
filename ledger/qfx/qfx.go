package qfx

import (
	"encoding/xml"
	"io"
)

// QFX/OFX XML structures (simplified for bank statement transactions)

type OFX struct {
	BankMsgsRsV1 BankMsgsRsV1 `xml:"BANKMSGSRSV1"`
}

type BankMsgsRsV1 struct {
	StmtTrnRs StmtTrnRs `xml:"STMTTRNRS"`
}

type StmtTrnRs struct {
	StmtRs StmtRs `xml:"STMTRS"`
}

type StmtRs struct {
	BankTranList BankTranList `xml:"BANKTRANLIST"`
}

type BankTranList struct {
	StmtTrn []StmtTrn `xml:"STMTTRN"`
}

type StmtTrn struct {
	TrnType  string `xml:"TRNTYPE"`
	DtPosted string `xml:"DTPOSTED"`
	TrnAmt   string `xml:"TRNAMT"`
	FitID    string `xml:"FITID"`
	Memo     string `xml:"MEMO"`
}

// ParseQFX parses a QFX/OFX XML document and returns the list of statement
// transactions contained in the first bank statement response.
func ParseQFX(reader io.Reader) ([]StmtTrn, error) {
	var ofx OFX
	if err := xml.NewDecoder(reader).Decode(&ofx); err != nil {
		return nil, err
	}

	return ofx.BankMsgsRsV1.StmtTrnRs.StmtRs.BankTranList.StmtTrn, nil
}
