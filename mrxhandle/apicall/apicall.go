package apicall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"gopkg.in/yaml.v3"
)

// API Extract takes a generic input, converts it to the JSON bytes and makes the API call
// It validates the inputs and outputs against the OpenAPI specification of that API
func ApiExtract[T any](toTransform []any, API, APISpec string) ([]T, error) {

	// Load the OpenAPI Spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(APISpec)
	if err != nil {
		return nil, err
	}

	extracted := make([]T, len(toTransform))

	for i, tt := range toTransform {

		// convert to jsonbytes then make the call
		in, _ := json.Marshal(tt)
		resp, err := http.Post(API, "application/json", bytes.NewReader(in))
		if err != nil {
			return nil, err
		}

		// API Schema Fluff
		// checking for errors afterwards
		// generate a copy of the request body that was made to check against
		ctx := context.Background()
		httpReq, err := http.NewRequest(http.MethodPost, "http://localhost:1323/3dTransform", bytes.NewReader(in))
		if err != nil {
			return nil, err
		}
		// add the real header
		httpReq.Header.Add("Content-Type", "application/json")
		router, err := gorillamux.NewRouter(doc)
		if err != nil {
			return nil, err
		}
		route, pathParams, err := router.FindRoute(httpReq)
		if err != nil {
			return nil, err
		}
		// Validate request
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    httpReq,
			PathParams: pathParams,
			Route:      route,
		}
		err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
		if err != nil {
			return nil, err
		}

		// get the result and return it as the required golang struct
		resBody, _ := io.ReadAll(resp.Body)
		err = yaml.Unmarshal(resBody, &extracted[i])

		if err != nil {
			return nil, err
		}

	}

	return extracted, nil
}

// API Extract takes bytes and makes the API call returning the bytes
// It validates the inputs and outputs against the OpenAPI specification of that API
func ApiExtractBytes(toTransform [][]byte, API, APISpec, dataFormat string) ([][]byte, error) {

	// Load the OpenAPI Spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(APISpec)
	if err != nil {
		return nil, err
	}

	extracted := make([][]byte, len(toTransform))

	for i, in := range toTransform {

		// convert to jsonbytes then make the call

		resp, err := http.Post(API, "application/json", bytes.NewReader(in))
		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}

		// API Schema Fluff
		// checking for errors afterwards
		// generate a copy of the request body that was made to check against
		ctx := context.Background()
		httpReq, err := http.NewRequest(http.MethodPost, API, bytes.NewReader(in))
		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}
		// add the real header
		httpReq.Header.Add("Content-Type", dataFormat)
		router, err := gorillamux.NewRouter(doc)
		if err != nil {
			return nil, err
		}
		route, pathParams, err := router.FindRoute(httpReq)
		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}
		// Validate request
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    httpReq,
			PathParams: pathParams,
			Route:      route,
		}
		err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}

		// get the result and return it as the required golang struct
		resBody, err := io.ReadAll(resp.Body)
		extracted[i] = resBody

		// Validate response using OpenAPI
		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 resp.StatusCode,
			Header:                 resp.Header,
		}
		responseValidationInput.SetBodyBytes(resBody)
		err = openapi3filter.ValidateResponse(ctx, responseValidationInput)
		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}

		if resp.StatusCode != http.StatusOK {
			var e errMessage
			json.Unmarshal(resBody, &e)
			return nil, fmt.Errorf(" error at data point %v : %v", i, e.Message)
		}

		if err != nil {
			return nil, fmt.Errorf(" error at data point %v : %v", i, err.Error())
		}

	}

	return extracted, nil
}

type errMessage struct {
	Message string `json:"message"`
}
