package camt

import (
	"encoding/xml"
	"io"
)

// XML structures for CAMT.053 format
type Document struct {
	XMLName       xml.Name      `xml:"Document"`
	BkToCstmrStmt BkToCstmrStmt `xml:"BkToCstmrStmt"`
}

type BkToCstmrStmt struct {
	Stmt Stmt `xml:"Stmt"`
}

type Stmt struct {
	Acct Acct   `xml:"Acct"`
	Ntry []Ntry `xml:"Ntry"`
}

type Acct struct {
	Id   Id     `xml:"Id"`
	Ccy  string `xml:"Ccy"`
	Ownr Ownr   `xml:"Ownr"`
}

type Id struct {
	IBAN string `xml:"IBAN"`
}

type Ownr struct {
	Nm string `xml:"Nm"`
}

type Ntry struct {
	Amt          Amount    `xml:"Amt"`
	CdtDbtInd    string    `xml:"CdtDbtInd"`
	BookgDt      BookgDt   `xml:"BookgDt"`
	BkTxCd       BkTxCd    `xml:"BkTxCd"`
	NtryRef      string    `xml:"NtryRef"`
	AddtlNtryInf string    `xml:"AddtlNtryInf"`
	NtryDtls     *NtryDtls `xml:"NtryDtls"`
}

type Amount struct {
	Value string `xml:",chardata"`
	Ccy   string `xml:"Ccy,attr"`
}

type BookgDt struct {
	DtTm string `xml:"DtTm"`
}

type BkTxCd struct {
	Prtry Prtry `xml:"Prtry"`
}

type Prtry struct {
	Cd string `xml:"Cd"`
}

type NtryDtls struct {
	TxDtls TxDtls `xml:"TxDtls"`
}

type TxDtls struct {
	RltdPties RltdPties `xml:"RltdPties"`
}

type RltdPties struct {
	Cdtr *Cdtr `xml:"Cdtr"`
}

type Cdtr struct {
	Pty Pty `xml:"Pty"`
}

type Pty struct {
	Nm string `xml:"Nm"`
}

func ParseCamt(reader io.Reader) ([]Ntry, error) {
	var doc Document
	if err := xml.NewDecoder(reader).Decode(&doc); err != nil {
		return nil, err
	}

	return doc.BkToCstmrStmt.Stmt.Ntry, nil
}
