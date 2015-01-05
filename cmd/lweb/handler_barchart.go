package main

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/howeyc/ledger/pkg/ledger"

	"github.com/gorilla/mux"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
)

func BarChartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountNames := vars["accountNames"]
	accounts := strings.Split(accountNames, ",")
	startString := vars["startDate"]
	endString := vars["endDate"]

	ledgerFileReader := bytes.NewReader(ledgerBuffer.Bytes())

	trans, terr := ledger.ParseLedger(ledgerFileReader)
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}

	barc := chart.BarChart{Title: "Bar Chart"}
	barc.XRange.Label = "Fiscal Year"
	barc.YRange.Label = "USD"
	barc.ShowVal = 1

	parsedStartDate, _ := time.Parse("2006-01-02", startString)
	parsedEndDate, _ := time.Parse("2006-01-02", endString)

	categories := []string{}
	pointData := make(map[string][]chart.Point)
	xIdx := 0
	for startDate := parsedStartDate; startDate.Before(parsedEndDate); startDate = startDate.AddDate(1, 0, 0) {
		endDate := startDate.AddDate(1, 0, 1)
		timeStartIndex, timeEndIndex := 0, 0
		for idx := 0; idx < len(trans); idx++ {
			if trans[idx].Date.After(startDate) {
				timeStartIndex = idx
				break
			}
		}
		for idx := len(trans) - 1; idx >= 0; idx-- {
			if trans[idx].Date.Before(endDate) {
				timeEndIndex = idx
				break
			}
		}
		ytrans := trans[timeStartIndex : timeEndIndex+1]

		balances := ledger.GetBalances(ytrans, accounts)

		for _, accName := range accounts {
			for _, bal := range balances {
				if accName == bal.Name {
					value, _ := bal.Balance.Float64()
					value = math.Abs(value)
					point := chart.Point{X: float64(xIdx), Y: value}
					if curData, found := pointData[accName]; found {
						curData = append(curData, point)
						pointData[accName] = curData
					} else {
						pointData[accName] = []chart.Point{point}
					}
				}
			}
		}
		categories = append(categories, fmt.Sprintf("%d", startDate.Year()))
		xIdx++
	}
	barc.XRange.Category = categories

	for aIdx, accName := range accounts {
		style := chart.AutoStyle(aIdx, true)
		barc.AddData(accName, pointData[accName], style)
	}

	igr := imgg.New(1600, 900, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	barc.Plot(igr)

	// Set type to png
	png.Encode(w, igr.Image)
}
