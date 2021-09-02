package batch

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeBatchRequest(t *testing.T) {
	batch := New("https://httpbin.org", &testLogger{})

	body := strings.NewReader(`{
		"operations" : [
			{
				"method": "POST",
				"path": "/v1/tests",
				"headers": [
					{
						"name": "Accept",
						"value" : "application/json"
					}
				],
				"bulk_id": "123",
				"body" : {
					"abc" : "value1",
					"def" : {
						"a1" : 1,
						"a2" : "dfsfs"
					},
					"ghi" : [
						{
							"b1" :  2,
							"b2": "asdasd"
						},
						{
							"b1" :  3,
							"b2": "sdgdfgdg"
						}
					]
				}
			},
			{
				"method": "GET",
				"path": "/v1/tests/123-1231-1231",
				"headers": [
					{
						"name": "Accept",
						"value" : "application/json"
					}
				],
				"bulk_id": "123"
			}
		] }`)
	req, err := http.NewRequest("POST", "/batch", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	decodedPayload, err := batch.DecodeBatchRequest(rw, req)
	if err != nil {
		t.Error(err)
	}
	if rw.Code > 299 {
		t.Errorf("Got error was expecting success : %s", rw.Body.String())
	}

	if len(decodedPayload.Operations) != 2 {
		t.Errorf("Expected %d operation, got %d", 2, len(decodedPayload.Operations))
	}

}

func TestProcess(t *testing.T) {

	batch := New("https://httpbin.org", &testLogger{})

	operationsPayload := BatchPayload{}

	basicGet := Operation{}
	basicGet.Method = "GET"
	basicGet.BulkId = "1"
	basicGet.Path = "/json"

	operationsPayload.Operations = append(operationsPayload.Operations, basicGet)

	basicPost := Operation{}
	basicPost.Method = "POST"
	basicPost.BulkId = "2"
	basicPost.Path = "/post"
	body := `{"abc":123"}`
	basicPost.Body = &body

	operationsPayload.Operations = append(operationsPayload.Operations, basicPost)

	results, err := batch.Process(context.Background(), operationsPayload)
	if err != nil {
		t.Error(err)
	}

	if len(results.Operations) != 2 {
		t.Errorf("Expected %d operation result, got %d", 1, len(results.Operations))
	}

	// log.Print(results.Operations[0].Status.Code)
	// log.Print(*results.Operations[0].Body)

	// log.Print(results.Operations[1].Status.Code)
	// log.Print(*results.Operations[1].Body)

}

type testLogger struct{}

func (tl *testLogger) Infow(msg string, keysAndValues ...interface{}) {
	tl.log(msg, "INFO", keysAndValues)
}
func (tl *testLogger) Errorw(msg string, keysAndValues ...interface{}) {
	tl.log(msg, "ERR", keysAndValues)
}
func (tl *testLogger) Warnw(msg string, keysAndValues ...interface{}) {
	tl.log(msg, "WARN", keysAndValues)
}

func (tl *testLogger) log(msg, level string, keysAndValues ...interface{}) {
	var keysAndValuesString string
	for _, kv := range keysAndValues {
		keysAndValuesString += fmt.Sprintf(", %v", kv)
	}
	log.Printf("[%s] %s%s", level, msg, keysAndValuesString)
}
