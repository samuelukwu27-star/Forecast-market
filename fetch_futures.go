// fetch_futures.go — Fetch real E-mini futures daily bars from Massive
package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"time"

	massive "github.com/massive-com/client-go/v2/rest"
	"github.com/massive-com/client-go/v2/rest/models"
)

type Bar struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

func main() {
	apiKey := os.Getenv("MASSIVE_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ MASSIVE_API_KEY not set")
	}

	c := massive.New(apiKey)
	ctx := context.Background()

	// Define front-month contracts covering 2024–2025
	contracts := map[string][]string{
		"NQ": {"CME.NQ.H24", "CME.NQ.M24", "CME.NQ.U24", "CME.NQ.Z24", "CME.NQ.H25", "CME.NQ.M25"},
		"ES": {"CME.ES.H24", "CME.ES.M24", "CME.ES.U24", "CME.ES.Z24", "CME.ES.H25", "CME.ES.M25"},
	}

	for asset, tickers := range contracts {
		file, err := os.Create(asset + "_2024_2025.csv")
		if err != nil {
			log.Fatalf("Cannot create %s.csv: %v", asset, err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Write([]string{"date", "open", "high", "low", "close", "volume"})
		defer writer.Flush()

		allBars := []Bar{}

		for _, ticker := range tickers {
			params := &models.GetAggsv2Params{
				Ticker:     ticker,
				Multiplier: 1,
				Timespan:   models.Day,
				From:       models.Nanos(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				To:         models.Nanos(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
			}

			iter := c.GetAggsv2(ctx, params)
			for iter.Next() {
				a := iter.Item()
				t := time.Unix(0, int64(a.Timestamp)*int64(time.Millisecond))
				dateStr := t.Format("2006-01-02")
				allBars = append(allBars, Bar{
					Date:   dateStr,
					Open:   a.Open,
					High:   a.High,
					Low:    a.Low,
					Close:  a.Close,
					Volume: a.Volume,
				})
			}
			if err := iter.Err(); err != nil {
				log.Printf("Warning: %s %v", ticker, err)
			}
		}

		// Sort and dedupe by date (keep latest contract for roll days)
		seen := make(map[string]bool)
		for i := len(allBars) - 1; i >= 0; i-- {
			bar := allBars[i]
			if !seen[bar.Date] {
				seen[bar.Date] = true
				writer.Write([]string{
					bar.Date,
					f2s(bar.Open), f2s(bar.High), f2s(bar.Low), f2s(bar.Close),
					i64s(bar.Volume),
				})
			}
		}

		log.Printf("✅ %s: %d daily bars saved", asset, len(seen))
	}
}

func f2s(f float64) string  { return fmt.Sprintf("%.2f", f) }
func i64s(i int64) string   { return fmt.Sprintf("%d", i) }
import "fmt"
