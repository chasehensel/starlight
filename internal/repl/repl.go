package repl

import (
	"awans.org/aft/internal/bus"
	"awans.org/aft/internal/db"
	"awans.org/aft/internal/server/lib"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

type REPLRequest struct {
	Data string `json:"data"`
}

type REPLResponse struct {
	Data interface{} `json:"data"`
}

type REPLHandler struct {
	bus *bus.EventBus
	db  db.DB
}

func (rh REPLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	var rr REPLRequest
	buf, _ := ioutil.ReadAll(r.Body)
	err = jsoniter.Unmarshal(buf, &rr)
	if err != nil {
		return
	}

	rh.bus.Publish(lib.ParseRequest{Request: rr})

	rwtx := rh.db.NewRWTx()
	replOut := eval(rr.Data, rwtx)
	rwtx.Commit()

	response := REPLResponse{Data: replOut}
	// write out the response
	bytes, _ := jsoniter.Marshal(&response)
	_, _ = w.Write(bytes)
	w.WriteHeader(http.StatusOK)
	return
}
