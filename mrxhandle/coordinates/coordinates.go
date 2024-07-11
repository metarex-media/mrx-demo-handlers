// package coordinates contains the Metarex Ids and transformations
// of coordinate based metadata
package coordinates

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/apicall"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/mrxlog"
)

const (
	// XYZ ID
	MRXID = "MRX.123.456.789.pqr"
	// Velocity ID
	MRXIDVel = "MRX.123.456.789.yza"
)

// The input data format we know how to handle
type MRXDataFormat struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// An output of normalised data
type MRXNormalDataFormat struct {
	XNormal float64 `json:"xNormalised"`
	YNormal float64 `json:"yNormalised"`
	ZNormal float64 `json:"zNormalised"`
}

// an output of the data as velocity
type MRXVelocity struct {
	VX      float64 `json:"XVelocity"`
	VY      float64 `json:"YVelocity"`
	VZ      float64 `json:"ZVelocity"`
	VNormal float64 `json:"Velocity"`
}

// an output of acceleration
type MRXAcceleration struct {
	AX      float64 `json:"XAcceleration "`
	AY      float64 `json:"YAcceleration "`
	AZ      float64 `json:"ZAcceleration "`
	ANormal float64 `json:"Acceleration "`
}

const (
	// normalise the XYZ data
	Normalise = "Normalise"
	// Calculate the velocity instead
	Velocity = "Velocity"
)

// Handle velocity transforms velocity to acceleration
func HandleVelocity(MRX *mrxlog.MRXHistory, input []any, API, APISpec, Action string) (any, error) {

	vels := make([]MRXVelocity, len(input))

	// convert input to velocities
	for i, v := range input {
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &vels[i])
	}

	acc := make([]MRXAcceleration, len(input))

	// first one always has a velocity of 0
	acc[0] = MRXAcceleration{}
	MRX = MRX.PushChild(mrxlog.MRXHistory{MrxID: MRXIDVel})

	for i, v := range vels[1:] {

		vx := v.VX - vels[i].VX
		vy := v.VY - vels[i].VY
		vz := v.VZ - vels[i].VZ
		vn := v.VNormal - vels[i].VNormal
		acc[i+1] = MRXAcceleration{AX: vx, AY: vy, AZ: vz, ANormal: vn}
	}
	MRX.LogDebug(fmt.Sprintf("transformed to %s", "MRX.DEV.DEV.DEV.DEV"))

	return acc, nil
}

// Transform XYZ changes the XYZ to normalised coordiantes
// or to velocity
func TransformXYZ(MRX *mrxlog.MRXHistory, input []any, API, APISpec, Action string) (any, error) {

	MRX.LogWarn(fmt.Sprintf("Transforming using %v", API))
	// Extract into the known data form
	vals, err := apicall.ApiExtract[MRXDataFormat](input, API, APISpec)

	if err != nil {
		return nil, err
	}
	MRX = MRX.PushChild(*mrxlog.NewMRX(MRXID))

	switch Action {
	case Velocity:
		return VelocityCalc(MRX, vals)
	default:
		return normalise(MRX, vals)
	}

	//return normalised, nil
}

func normalise(MRX *mrxlog.MRXHistory, input []MRXDataFormat) ([]MRXNormalDataFormat, error) {
	maxX, maxY, maxZ := 0.0, 0.0, 0.0

	for _, v := range input {
		if math.Abs(v.X) > maxX {
			maxX = float64(v.X)
		}

		if math.Abs(v.Y) > maxY {
			maxY = float64(v.Y)
		}

		if math.Abs(v.Z) > maxZ {
			maxZ = float64(v.Z)
		}
	}

	normalised := make([]MRXNormalDataFormat, len(input))

	for i, v := range input {
		normalised[i] = MRXNormalDataFormat{XNormal: (v.X / maxX), YNormal: (v.Y / maxY), ZNormal: (v.Z / maxZ)}
	}
	MRX = MRX.PushChild(*mrxlog.NewMRX("MRX.DEV.DEV.DEV.DEV"))
	MRX.LogDebug(fmt.Sprintf("transformed to %s", "MRX.DEV.DEV.DEV.DEV"))
	return normalised, nil
}

/*
VelocityCalc finds velocity with the following formula

3 velocity vectors (vx, vy, vz) and after only do the rms averaging (v_avg = sqrt(vx² + vy² + vz²))
*/
func VelocityCalc(MRX *mrxlog.MRXHistory, input []MRXDataFormat) ([]MRXVelocity, error) {

	MRX.PushChild(*mrxlog.NewMRX(MRXIDVel))
	MRX.LogDebug(fmt.Sprintf("transformed to %s", MRXIDVel))
	vel := make([]MRXVelocity, len(input))

	// first one always has a velocity of 0
	vel[0] = MRXVelocity{}

	for i, v := range input[1:] {

		vx := v.X - input[i].X
		vy := v.Y - input[i].Y
		vz := v.Z - input[i].Z
		vn := math.Sqrt(vx*vx + vy*vy + vz*vz)
		vel[i+1] = MRXVelocity{VX: vx, VY: vy, VZ: vz, VNormal: vn}
	}

	return vel, nil
}
