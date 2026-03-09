// Package solax provides a client for the SolaX Cloud User API v2.
package solax

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const DefaultBaseURL = "https://global.solaxcloud.com"

// InverterStatus represents the operational status of an inverter.
type InverterStatus int

// Inverter status codes per appendix 8.1.
const (
	StatusWaitingForOperation    InverterStatus = 100
	StatusSelfTest               InverterStatus = 101
	StatusNormal                 InverterStatus = 102
	StatusRecoverableFault       InverterStatus = 103
	StatusPermanentFault         InverterStatus = 104
	StatusFirmwareUpgrade        InverterStatus = 105
	StatusEPSDetection           InverterStatus = 106
	StatusOffGrid                InverterStatus = 107
	StatusSelfTestItalian        InverterStatus = 108
	StatusSleepMode              InverterStatus = 109
	StatusStandbyMode            InverterStatus = 110
	StatusPVWakeUpBattery        InverterStatus = 111
	StatusGenDetection           InverterStatus = 112
	StatusGenMode                InverterStatus = 113
	StatusFastShutdownStandby    InverterStatus = 114
	StatusVPPMode                InverterStatus = 130
	StatusTOUSelfUse             InverterStatus = 131
	StatusTOUCharging            InverterStatus = 132
	StatusTOUDischarging         InverterStatus = 133
	StatusTOUBatteryOff          InverterStatus = 134
	StatusTOUPeakShaving         InverterStatus = 135
	StatusNormalGenOperation     InverterStatus = 136
	StatusBatteryExpansion       InverterStatus = 137
	StatusOnGridBatteryHeating   InverterStatus = 138
	StatusEPSBatteryHeating      InverterStatus = 139
	StatusNormalModeR1           InverterStatus = 141
	StatusNormalModeR2           InverterStatus = 142
	StatusNormalModeR3           InverterStatus = 143
	StatusNormalModeR4           InverterStatus = 144
	StatusNormalModeR5           InverterStatus = 145
	StatusNormalModeR6           InverterStatus = 146
	StatusNormalModeR7           InverterStatus = 147
	StatusNormalModeSS           InverterStatus = 148
	StatusSelfUse                InverterStatus = 150
	StatusForceTimeUse           InverterStatus = 151
	StatusBackUpMode             InverterStatus = 152
	StatusFeedinPriority         InverterStatus = 153
	StatusDemandMode             InverterStatus = 154
	StatusConstPowerMode         InverterStatus = 155
	StatusOpenAdrMode            InverterStatus = 160
)

var inverterStatusNames = map[InverterStatus]string{
	StatusWaitingForOperation:  "Waiting for operation",
	StatusSelfTest:             "Self-test",
	StatusNormal:               "Normal",
	StatusRecoverableFault:     "Recoverable fault",
	StatusPermanentFault:       "Permanent fault",
	StatusFirmwareUpgrade:      "Firmware upgrade",
	StatusEPSDetection:         "EPS detection",
	StatusOffGrid:              "Off-grid",
	StatusSelfTestItalian:      "Self-test mode (Italian safety regulations)",
	StatusSleepMode:            "Sleep mode",
	StatusStandbyMode:          "Standby mode",
	StatusPVWakeUpBattery:      "Photovoltaic wake-up battery mode",
	StatusGenDetection:         "Generator detection mode",
	StatusGenMode:              "Generator mode",
	StatusFastShutdownStandby:  "Fast shutdown standby mode",
	StatusVPPMode:              "VPP mode",
	StatusTOUSelfUse:           "TOU-Self use",
	StatusTOUCharging:          "TOU-Charging",
	StatusTOUDischarging:       "TOU-Discharging",
	StatusTOUBatteryOff:        "TOU-Battery off",
	StatusTOUPeakShaving:       "TOU-Peak Shaving",
	StatusNormalGenOperation:   "Normal generator operation mode",
	StatusBatteryExpansion:     "Battery expansion mode",
	StatusOnGridBatteryHeating: "On-grid battery heating mode",
	StatusEPSBatteryHeating:    "EPS battery heating mode",
	StatusNormalModeR1:         "NormalMode(R-1)",
	StatusNormalModeR2:         "NormalMode(R-2)",
	StatusNormalModeR3:         "NormalMode(R-3)",
	StatusNormalModeR4:         "NormalMode(R-4)",
	StatusNormalModeR5:         "NormalMode(R-5)",
	StatusNormalModeR6:         "NormalMode(R-6)",
	StatusNormalModeR7:         "NormalMode(R-7)",
	StatusNormalModeSS:         "NormalMode(SS)",
	StatusSelfUse:              "Self Use",
	StatusForceTimeUse:         "Force Time Use",
	StatusBackUpMode:           "Back Up Mode",
	StatusFeedinPriority:       "Feedin Priority",
	StatusDemandMode:           "Demand Mode",
	StatusConstPowerMode:       "ConstPowr Mode",
	StatusOpenAdrMode:          "OpenAdr Mode",
}

func (s InverterStatus) String() string {
	if name, ok := inverterStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", int(s))
}

// RealtimeData holds the real-time data returned by the SolaX Cloud API.
// Float fields use *float64 because the API returns null for unsupported values.
// InverterStatus is returned as a string by the API (e.g. "102") and decoded
// into the InverterStatus named type.
type RealtimeData struct {
	InverterSN     string         `json:"inverterSN"`
	SN             string         `json:"sn"`
	ACPower        *float64       `json:"acpower"`
	YieldToday     *float64       `json:"yieldtoday"`
	YieldTotal     *float64       `json:"yieldtotal"`
	FeedinPower    *float64       `json:"feedinpower"`
	FeedinEnergy   *float64       `json:"feedinenergy"`
	ConsumeEnergy  *float64       `json:"consumeenergy"`
	FeedinPowerM2  *float64       `json:"feedinpowerM2"`
	SOC            *float64       `json:"soc"`
	Peps1          *float64       `json:"peps1"`
	Peps2          *float64       `json:"peps2"`
	Peps3          *float64       `json:"peps3"`
	InverterType   string         `json:"inverterType"`
	InverterStatus InverterStatus `json:"-"`
	UploadTime     string         `json:"uploadTime"`
	BatPower       *float64       `json:"batPower"`
	PowerDC1       *float64       `json:"powerdc1"`
	PowerDC2       *float64       `json:"powerdc2"`
	PowerDC3       *float64       `json:"powerdc3"`
	PowerDC4       *float64       `json:"powerdc4"`
	BatStatus      *string        `json:"batStatus"`
	UtcDateTime    string         `json:"utcDateTime"`
}

// UnmarshalJSON implements custom unmarshalling to handle inverterStatus
// being returned as a string (e.g. "102") by the API.
func (r *RealtimeData) UnmarshalJSON(data []byte) error {
	type Alias RealtimeData
	aux := &struct {
		*Alias
		RawInverterStatus string `json:"inverterStatus"`
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.RawInverterStatus != "" {
		v, err := strconv.Atoi(aux.RawInverterStatus)
		if err != nil {
			return fmt.Errorf("parsing inverterStatus %q: %w", aux.RawInverterStatus, err)
		}
		r.InverterStatus = InverterStatus(v)
	}
	return nil
}

// APIError represents an error returned by the SolaX Cloud API.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("solax api error %d: %s", e.Code, e.Message)
}

// Error codes per appendix 8.2.
var apiErrorMessages = map[int]string{
	1001: "Interface Unauthorized",
	1002: "Parameter validation failed",
	1003: "Data Unauthorized",
	1004: "Duplicate data",
	2001: "Operation failed",
	2002: "Data not found",
}

// Client is a SolaX Cloud API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new SolaX Cloud API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: http.DefaultClient,
	}
}

type apiRequest struct {
	WifiSN string `json:"wifiSn"`
}

type apiResponse struct {
	Success   bool          `json:"success"`
	Exception string        `json:"exception"`
	Code      int           `json:"code"`
	Result    *RealtimeData `json:"result"`
}

// GetRealtimeData fetches real-time inverter data for the given wifi serial number.
func (c *Client) GetRealtimeData(ctx context.Context, wifiSn string) (*RealtimeData, error) {
	body, err := json.Marshal(apiRequest{WifiSN: wifiSn})
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	url := c.BaseURL + "/api/v2/dataAccess/realtimeInfo/get"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tokenId", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !apiResp.Success || apiResp.Code != 0 {
		msg := apiResp.Exception
		if msg == "" {
			if known, ok := apiErrorMessages[apiResp.Code]; ok {
				msg = known
			} else {
				msg = "unknown error"
			}
		}
		return nil, &APIError{Code: apiResp.Code, Message: msg}
	}

	if apiResp.Result == nil {
		return nil, fmt.Errorf("api returned success but no result data")
	}

	return apiResp.Result, nil
}
