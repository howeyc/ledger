package main

import (
	"bytes"
	"image/color"
	"image/png"
	"net/http"
	"strings"

	"github.com/howeyc/ledger/pkg/ledger"

	"github.com/gorilla/mux"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
)

func PieChartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountName := vars["accountName"]

	ledgerFileReader := bytes.NewReader(ledgerBuffer.Bytes())

	trans, terr := ledger.ParseLedger(ledgerFileReader)
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}
	balances := ledger.GetBalances(trans, []string{accountName})

	skipCount := 0
	for _, account := range balances {
		if !strings.HasPrefix(account.Name, accountName) {
			skipCount++
		}
		if account.Name == accountName {
			skipCount++
		}
	}

	accNames := make([]string, 0)
	values := make([]float64, 0)
	accStartLen := len(accountName)

	for _, account := range balances[skipCount:] {
		accName := account.Name[accStartLen+1:]
		value, _ := account.Balance.Float64()
		if !strings.Contains(accName, ":") {
			accNames = append(accNames, accName)
			values = append(values, value)
		}
	}

	piec := chart.PieChart{Title: "Pie Chart"}
	piec.FmtVal = chart.AbsoluteValue
	piec.FmtKey = chart.PercentValue
	piec.AddDataPair(accountName, accNames, values)
	piec.Inner = 0.5

	igr := imgg.New(1600, 900, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	piec.Plot(igr)

	// Set type to png
	png.Encode(w, igr.Image)
}
