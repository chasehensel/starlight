package rpc

import (
	"awans.org/aft/internal/bus"
	"awans.org/aft/internal/db"
	"awans.org/aft/internal/server/lib"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

type RPCRequest struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

type RPCResponse struct {
	Response interface{} `json:"response"`
}

type RPCHandler struct {
	bus *bus.EventBus
	db  db.DB
}

func (rh RPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	var rr RPCRequest
	buf, _ := ioutil.ReadAll(r.Body)
	err = jsoniter.Unmarshal(buf, &rr)
	if err != nil {
		return
	}

	rh.bus.Publish(lib.ParseRequest{Request: rr})

	rwtx := rh.db.NewRWTx()
	RPCOut, err := eval(rr.Name, rr.Data, rwtx)
	if err != nil {
		return
	}
	rwtx.Commit()
	response := RPCResponse{Response: RPCOut}

	bytes, _ := jsoniter.Marshal(&response)
	_, _ = w.Write(bytes)
	w.WriteHeader(http.StatusOK)
	return
}
