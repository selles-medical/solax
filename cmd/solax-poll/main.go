package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/selles-medical/solax"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: solax-poll <wifi-sn>\n")
		os.Exit(1)
	}
	wifiSn := os.Args[1]

	token := os.Getenv("SOLAX_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "error: SOLAX_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	baseURL := os.Getenv("SOLAX_BASE_URL")
	if baseURL == "" {
		baseURL = solax.DefaultBaseURL
	}

	client := solax.NewClient(baseURL, token)

	data, err := client.GetRealtimeData(context.Background(), wifiSn)
	if err != nil {
		var apiErr *solax.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API error [%d]: %s\n", apiErr.Code, apiErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("SolaX Inverter Realtime Data\n")
	fmt.Printf("============================\n")
	fmt.Printf("Inverter SN:     %s\n", data.InverterSN)
	fmt.Printf("Wi-Fi SN:        %s\n", data.SN)
	fmt.Printf("Status:          %s\n", data.InverterStatus)
	fmt.Printf("Inverter Type:   %s\n", data.InverterType)
	fmt.Printf("Upload Time:     %s\n", data.UploadTime)
	fmt.Printf("UTC DateTime:    %s\n", data.UtcDateTime)
	fmt.Println()

	fmt.Printf("AC Power:        %s W\n", fmtFloat(data.ACPower))
	fmt.Printf("Yield Today:     %s kWh\n", fmtFloat(data.YieldToday))
	fmt.Printf("Yield Total:     %s kWh\n", fmtFloat(data.YieldTotal))
	fmt.Println()

	fmt.Printf("Feed-in Power:   %s W\n", fmtFloat(data.FeedinPower))
	fmt.Printf("Feed-in Energy:  %s kWh\n", fmtFloat(data.FeedinEnergy))
	fmt.Printf("Consume Energy:  %s kWh\n", fmtFloat(data.ConsumeEnergy))
	fmt.Printf("Feed-in M2:      %s W\n", fmtFloat(data.FeedinPowerM2))
	fmt.Println()

	fmt.Printf("Battery SOC:     %s %%\n", fmtFloat(data.SOC))
	fmt.Printf("Battery Power:   %s W\n", fmtFloat(data.BatPower))
	fmt.Printf("Battery Status:  %s\n", fmtStr(data.BatStatus))
	fmt.Println()

	fmt.Printf("PV1 Power:       %s W\n", fmtFloat(data.PowerDC1))
	fmt.Printf("PV2 Power:       %s W\n", fmtFloat(data.PowerDC2))
	fmt.Printf("PV3 Power:       %s W\n", fmtFloat(data.PowerDC3))
	fmt.Printf("PV4 Power:       %s W\n", fmtFloat(data.PowerDC4))
	fmt.Println()

	fmt.Printf("EPS A Phase:     %s W\n", fmtFloat(data.Peps1))
	fmt.Printf("EPS B Phase:     %s W\n", fmtFloat(data.Peps2))
	fmt.Printf("EPS C Phase:     %s W\n", fmtFloat(data.Peps3))
}

func fmtFloat(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", *v)
}

func fmtStr(v *string) string {
	if v == nil {
		return "n/a"
	}
	return *v
}
