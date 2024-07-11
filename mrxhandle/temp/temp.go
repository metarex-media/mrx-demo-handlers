// temp takes a temperature in a known format and converts it from celsius to fahrenheit
package temp

import (
	"fmt"

	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/apicall"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/mrxlog"
)

const (
	MRXID  = "MRX.123.456.789.vwx"
	fahrID = "MRX.DEV.DEV.DEV.DEV"
)

type TempsOut struct {
	Temperature float64 `json:"temperature_celsius" yaml:"temperature_celsius"`
	Feels       float64 `json:"feels_like_celsius" yaml:"feels_like_celsius"`
	MinTemp     float64 `json:"temperature_min_celsius" yaml:"temperature_min_celsius"`
	MaxTemp     float64 `json:"temperature_max_celsius" yaml:"temperature_max_celsius"`
}

type TempsFahren struct {
	Temperature float64 `json:"temperature_fahrenheit" yaml:"temperature_fahrenheit"`
	Feels       float64 `json:"feels_like_fahrenheit" yaml:"feels_like_fahrenheit"`
	MinTemp     float64 `json:"temperature_min_fahrenheit" yaml:"temperature_min_fahrenheit"`
	MaxTemp     float64 `json:"temperature_max_fahrenheit" yaml:"temperature_max_fahrenheit"`
}

const (
	CentigradeToFahrenheit = "CentigradeToF"
)

func Transform(MRX *mrxlog.MRXHistory, input []any, API, APISpec, Action string) (any, error) {

	// res := apicall.ApiExtract[MRXDataFormat](input, API, APISpec)
	// extract the celsius values from the API
	MRX.LogWarn(fmt.Sprintf("Transforming using %v", API))
	results, err := apicall.ApiExtract[TempsOut](input, API, APISpec)
	if err != nil {

		// LOG error
		return nil, err
	}
	MRX = MRX.PushChild(mrxlog.MRXHistory{MrxID: MRXID})
	MRX.LogDebug(fmt.Sprintf("transformed to Known datatype %v", MRXID))
	//*MRX = newData

	// convert celsius to fahrenheit
	// and return as a new struct
	fahrenList := make([]TempsFahren, len(results))
	// (1°C × 9/5) + 32

	//	outputFormat := mrxlog.ErrMesage{MrxID: fahrID, Parent: &newData}
	for i, c := range results {
		fahrenList[i] = TempsFahren{Temperature: celToF(c.Temperature), Feels: celToF(c.Feels), MinTemp: celToF(c.MinTemp), MaxTemp: celToF(c.MaxTemp)}
	}

	MRX = MRX.PushChild(mrxlog.MRXHistory{MrxID: fahrID})
	MRX.LogDebug("transforming to fahrenheit data")

	return fahrenList, nil
}

// celToF is the formula for celsius to fahrenheit
func celToF(c float64) float64 {
	return (c * 1.8) + 32
}
