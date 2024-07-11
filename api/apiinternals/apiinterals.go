package apiinternals

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v3"
)

func Run() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/3dTransform", transform3d)
	e.POST("/tempTransform", transformTemp)
	e.POST("/mpegh", transformMPEGH)
	e.POST("/dolby", transformMPEGHtoDolby)
	e.POST("/sphere", transformSphere)
	e.POST("/xyzToSphere", trasnformXYZtoSphere)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// transform the 3d json into a generic 3d one which is used by everyone
func transform3d(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusBadRequest, message{"no data received"})
	}
	var target in3d
	err := json.Unmarshal(bodyBytes, &target)
	// company ticker

	if err != nil {

		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	out := out3d(target)
	return c.JSONPretty(http.StatusOK, out, "    ")
}

// transform the 3d json into a generic 3d one which is used by everyone
func transformMPEGH(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusBadRequest, message{"no data received"})
	}
	var target out3d
	err := json.Unmarshal(bodyBytes, &target)
	// company ticker

	if err != nil {

		return c.JSON(http.StatusBadRequest, message{err.Error()})
	}

	switch {
	case target.X > 100:
		return c.JSON(http.StatusBadRequest, message{fmt.Sprintf("the x value of %v is greater than 100", target.X)})
	case target.Y > 100:
		return c.JSON(http.StatusBadRequest, message{fmt.Sprintf("the x value of %v is greater than 100", target.Y)})
	case target.Z > 100:
		return c.JSON(http.StatusBadRequest, message{fmt.Sprintf("the x value of %v is greater than 100", target.Z)})
	}

	out := outMpegh{X: float64(target.X) / 100, Y: float64(target.Y) / 100, Z: float64(target.Z) / 100}
	return c.JSONPretty(http.StatusOK, out, "    ")
}

// transform the 3d json into a generic 3d one which is used by everyone
func transformMPEGHtoDolby(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusBadRequest, message{"no data received"})
	}
	var target outMpegh
	err := json.Unmarshal(bodyBytes, &target)
	// company ticker

	if err != nil {

		return c.JSON(http.StatusBadRequest, message{err.Error()})
	}

	out := in3d{X: (target.X * 100), Y: (target.Y * 100), Z: (target.Z * 100)}
	return c.JSONPretty(http.StatusOK, out, "    ")
}

type sphericalCoords struct {
	Azimuth   float64 `yaml:"azimuth"`
	R         float64 `yaml:"distance"`
	Elevation float64 `yaml:"elevation"`
}

func transformSphere(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusBadRequest, message{"no data received"})
	}
	var sc sphericalCoords
	err := yaml.Unmarshal(bodyBytes, &sc)
	if err != nil {

		return c.JSON(http.StatusBadRequest, message{err.Error()})
	}

	var out out3d
	out.X = sc.R * math.Sin((math.Pi*sc.Elevation)/180) * math.Cos((math.Pi*sc.Azimuth)/180)
	out.Y = sc.R * math.Sin((math.Pi*sc.Elevation)/180) * math.Sin((math.Pi*sc.Azimuth)/180)
	out.Z = sc.R * math.Cos((math.Pi*sc.Elevation)/180)

	return c.JSONPretty(http.StatusOK, out, "    ")
}

func trasnformXYZtoSphere(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusBadRequest, message{"no data received"})
	}
	var input out3d
	err := json.Unmarshal(bodyBytes, &input)
	if err != nil {

		return c.JSON(http.StatusBadRequest, message{err.Error()})
	}

	var out sphericalCoords
	r := math.Sqrt(input.X*input.X + input.Y*input.Y + input.Z*input.Z)
	out.R = r
	out.Elevation = math.Acos(input.Z/r) * (180 / math.Pi)
	out.Azimuth = math.Acos(input.X/math.Sqrt(input.X*input.X+input.Y*input.Y)) * math.Copysign(1, input.Y) * (180 / math.Pi)

	blob, _ := yaml.Marshal(out)
	return c.Blob(http.StatusOK, "application/yaml", blob)
	// return c.JSONPretty(http.StatusOK, out, "    ")
}

type message struct {
	Message string `json:"message"`
}

type in3d struct {
	X float64 `json:"x1"`
	Y float64 `json:"y1"`
	Z float64 `json:"z1"`
}

type out3d struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type outMpegh struct {
	X float64 `json:"xNormalised"`
	Y float64 `json:"yNormalised"`
	Z float64 `json:"zNormalised"`
}

// transform the temperature json into a generic one which is used by everyone
func transformTemp(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)

	} else {
		return c.JSON(http.StatusOK, true)
	}
	var target tempsIn
	err := json.Unmarshal(bodyBytes, &target)
	// company ticker

	if err != nil {

		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	out := tempsOut(target)
	return c.JSONPretty(http.StatusOK, out, "    ")
}

type tempsIn struct {
	Temperature float64 `json:"temperature"`
	Feels       float64 `json:"feels_like"`
	MinTemp     float64 `json:"temperature_min"`
	MaxTemp     float64 `json:"temperature_max"`
}

type tempsOut struct {
	Temperature float64 `json:"temperature_celsius"`
	Feels       float64 `json:"feels_like_celsius"`
	MinTemp     float64 `json:"temperature_min_celsius"`
	MaxTemp     float64 `json:"temperature_max_celsius"`
}
