package cmd

import (
	"log"
	"os"
	"time"

	"github.com/howeyc/ledger"
	"github.com/pelletier/go-toml"
)

type accountOp struct {
	Name                 string  `toml:"name"`
	Operation            string  `toml:"operation"` // +, -
	MultiplicationFactor float64 `toml:"factor"`
	SubAccount           string  `toml:"other_account"` // *, /
}

type calculatedAccount struct {
	Name              string      `toml:"name"`
	UseAbs            bool        `toml:"use_abs"`
	AccountOperations []accountOp `toml:"account_operation"`
}

type reportConfig struct {
	Name                   string
	Chart                  string
	RangeBalanceType       ledger.RangeType `toml:"range_balance_type"`
	RangeBalanceSkipZero   bool             `toml:"range_balance_skip_zero"`
	DateRange              string           `toml:"date_range"`
	DateFreq               string           `toml:"date_freq"`
	Accounts               []string
	ExcludeAccountTrans    []string            `toml:"exclude_account_trans"`
	ExcludeAccountsSummary []string            `toml:"exclude_account_summary"`
	CalculatedAccounts     []calculatedAccount `toml:"calculated_account"`
}

type reportConfigStruct struct {
	Reports []reportConfig `toml:"report"`
}

var reportConfigData reportConfigStruct

type quickviewAccountConfig struct {
	Name      string
	ShortName string `toml:"short_name"`
}

type quickviewConfigStruct struct {
	Accounts []quickviewAccountConfig `toml:"account"`
}

var quickviewConfigData quickviewConfigStruct

type stockConfig struct {
	Name         string
	SecurityType string `toml:"security_type"`
	Section      string
	Ticker       string
	Account      string
	Shares       float64
}

type stockInfo struct {
	Name    string
	Section string
	Type    string
	Ticker  string
	Account string
	Shares  float64

	Price                 float64
	PriceChangeDay        float64
	PriceChangePctDay     float64
	PriceChangeOverall    float64
	PriceChangePctOverall float64

	Cost            float64
	MarketValue     float64
	GainLossDay     float64
	GainLossOverall float64

	Weight float64

	AnnualDividends float64
	AnnualYield     float64
}

type portfolioStruct struct {
	Name string

	ShowDividends bool `toml:"show_dividends"`
	ShowWeight    bool `toml:"show_weight"`

	Stocks []stockConfig `toml:"stock"`
}

type portfolioConfigStruct struct {
	Portfolios []portfolioStruct `toml:"portfolio"`
	IEXToken   string            `toml:"iex_token"`
	AVToken    string            `toml:"av_token"`
}

var portfolioConfigData portfolioConfigStruct

type pageData struct {
	Reports      []reportConfig
	Transactions []*ledger.Transaction
	Accounts     []*ledger.Account
	Stocks       []stockInfo
	Portfolios   []portfolioStruct
	AccountNames []string
}

func configLoaders(dur time.Duration) {
	if len(reportConfigFileName) > 0 {
		go func() {
			for {
				var rLoadData reportConfigStruct
				ifile, ierr := os.Open(reportConfigFileName)
				if ierr != nil {
					log.Println(ierr)
				}
				tdec := toml.NewDecoder(ifile)
				err := tdec.Decode(&rLoadData)
				if err != nil {
					log.Println(err)
				}
				ifile.Close()
				reportConfigData = rLoadData
				time.Sleep(dur)
			}
		}()
	}

	if len(quickviewConfigFileName) > 0 {
		go func() {
			for {
				var sLoadData quickviewConfigStruct
				ifile, ierr := os.Open(quickviewConfigFileName)
				if ierr != nil {
					log.Println(ierr)
				}
				tdec := toml.NewDecoder(ifile)
				err := tdec.Decode(&sLoadData)
				if err != nil {
					log.Println(err)
				}
				ifile.Close()
				quickviewConfigData = sLoadData
				time.Sleep(dur)
			}
		}()
	}

	if len(stockConfigFileName) > 0 {
		go func() {
			for {
				var sLoadData portfolioConfigStruct
				ifile, ierr := os.Open(stockConfigFileName)
				if ierr != nil {
					log.Println(ierr)
				}
				tdec := toml.NewDecoder(ifile)
				err := tdec.Decode(&sLoadData)
				if err != nil {
					log.Println(err)
				}
				ifile.Close()
				portfolioConfigData = sLoadData
				time.Sleep(dur)
			}
		}()
	}

}
