package call_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/istomyang/call/call-sdk-golang"
	jsoniter "github.com/json-iterator/go"
)

func Example() {
	RunInMain()

	// Output:
	//
}

func RunInMain() {
	var (
		ctx    = context.Background()
		name   = "service-name"
		number string
	)
	number, _ = os.Hostname()

	call.Init(ctx, name, number, 3, ":3001")

	RunServer()

	// blocking run.
	call.Run()
}

func RunServer() {
	call.Handle(func(req []byte) (res []byte, err error) {
		var (
			data Data
		)
		if err = jsoniter.Unmarshal(req, &data); err != nil {
			return
		}

		// You can use a upper layer to do route dispatching.

		res = []byte("ok")
		err = nil

		return
	})
}

func RunClient() {
	var data = Data{
		Method:  http.MethodGet,
		URL:     "resource_group/resource/3",
		Params:  nil,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    nil,
	}
	b, _ := jsoniter.Marshal(data)

	call.Send(call.Request{
		ServiceName:   "user-service",
		ServiceNumber: nil,                 // if empty, use LB.
		LB:            call.LB_ROUND_ROBIN, // invalid when has ServiceNumber
		WriteOp:       true,
		Data:          b,
	})
}

type Data struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Params  map[string]string `json:"params"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type A struct {
	A *string
}

func Test_1(t *testing.T) {
	// var s = "{\"a\":null}"
	// var a A
	// _ = json.Unmarshal([]byte(s), &a)
	// log.Println(a)

	var a2 = A{A: nil}
	var s1, _ = json.Marshal(&a2)
	log.Println(string(s1))
}
