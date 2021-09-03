package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

const maximumOperation = 1024

// Logger defines the contract that a logger must implements
type logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
}

type Batch struct {
	*batch
}

type batch struct {
	basePath string
	logger   logger
	trace    func(context.Context, http.Header)
}

func New(basePath string, logger logger) *Batch {
	b := &batch{basePath: basePath, logger: logger}
	return &Batch{b}
}

func (b *Batch) With(tracer func(context.Context, http.Header)) *Batch {
	b.trace = tracer
	return b
}

func (b *Batch) DecodeBatchRequest(respWriter http.ResponseWriter, request *http.Request) (BatchPayload, error) {
	payload := &BatchPayload{}
	err := decodeJSONBody(request, payload)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(respWriter, mr.msg, mr.status)
		} else {
			b.errorw(err.Error())
			http.Error(respWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

	return *payload, nil
}

func (b *Batch) Process(ctx context.Context, operations BatchPayload) (BatchPayload, error) {

	result := BatchPayload{}

	if len(operations.Operations) > maximumOperation {
		return result, fmt.Errorf("maximum number of operations exceed")
	}

	errorCount := 0
	errorTolerance := math.MaxUint32
	if operations.FailOnErrors != nil {
		errorTolerance = int(*operations.FailOnErrors)
	}

	for _, anOperation := range operations.Operations {
		response, err := b.doOperation(ctx, anOperation)
		if err != nil || response.Status.CodeInt > 299 {
			errorCount += 1
			if errorTolerance <= errorCount {
				return result, err
			}
		}
		result.Operations = append(result.Operations, response)
	}

	return result, nil
}

func (b *Batch) doOperation(ctx context.Context, operation Operation) (Operation, error) {
	result := Operation{}
	path := b.basePath + operation.Path

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, operation.Method, path, nil)
	if err != nil {
		return result, err
	}

	// Prepare the request
	for _, aHeader := range operation.Headers {
		req.Header.Add(aHeader.Name, aHeader.Value)
	}
	if operation.Body != nil {
		rawBody, err := json.Marshal(operation.Body)
		if err != nil {
			//TODO
		}
		req.Body = io.NopCloser(bytes.NewReader(rawBody))
	}

	// Perform the request
	if b.trace != nil {
		b.trace(ctx, req.Header)
	}
	b.infow("Sending request", "path", path, "method", operation.Method)
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	// Prepare the response
	result.Status.CodeInt = resp.StatusCode
	result.Status.Code = fmt.Sprint(resp.StatusCode)
	for k, v := range resp.Header {
		result.Headers = append(result.Headers, Header{Name: k, Value: strings.Join(v, ",")})
	}

	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		//Partial response
		return result, err
	}

	rawResponse := new(interface{})
	if err := json.Unmarshal(responseBody, rawResponse); err != nil {
		return result, err
	}
	result.Body = rawResponse

	return result, nil
}

func (b *Batch) infow(msg string, keysAndValues ...interface{}) {
	if b.logger != nil {
		b.logger.Infow(msg, keysAndValues...)
	}
}
func (b *Batch) errorw(msg string, keysAndValues ...interface{}) {
	if b.logger != nil {
		b.logger.Errorw(msg, keysAndValues...)
	}
}
func (b *Batch) warnw(msg string, keysAndValues ...interface{}) {
	if b.logger != nil {
		b.logger.Warnw(msg, keysAndValues...)
	}
}

// b := batch.New("http://localhost/dsrs/batch", logger)
// b.With(tracer).Process(context.Background(), payload)
