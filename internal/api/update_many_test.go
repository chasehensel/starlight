package api

import (
	"awans.org/aft/internal/bus"
	"awans.org/aft/internal/db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateManyServerParseSimple(t *testing.T) {
	appDB := db.NewTest()
	eventbus := bus.New()
	db.AddSampleModels(appDB)

	jsonString := `{ "firstName":"Andrew", "lastName":"Wansley", "age": 32, "emailAddress":"andrew.wansley@gmail.com"}`
	u := makeRecord(appDB.NewTx(), "user", jsonString)
	cOp := CreateOperation{
		Record: u,
		Nested: []NestedOperation{},
	}
	jsonString2 := `{ "firstName":"Chase", "lastName":"Hensel", "age": 32, "emailAddress":"chase.hensel@gmail.com"}`
	u2 := makeRecord(appDB.NewTx(), "user", jsonString2)
	cOp2 := CreateOperation{
		Record: u2,
		Nested: []NestedOperation{},
	}

	tx := appDB.NewRWTx()
	cOp.Apply(tx)
	cOp2.Apply(tx)
	tx.Commit()

	req, err := http.NewRequest("POST", "/user.updateMany", strings.NewReader(
		`{"data":{
			"firstName":"bob"
		},
		"where": {
			"age": "32"
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{"modelName": "user"})

	cs := UpdateManyHandler{DB: appDB, Bus: eventbus}
	w := httptest.NewRecorder()
	err = cs.ServeHTTP(w, req)
	if err != nil {
		t.Fatal(err)
	}

	var data map[string]interface{}
	result := w.Result()
	bytes, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Error(err)
	}
	json.Unmarshal(bytes, &data)
	fmt.Printf("%s", data)
	objData := data["BatchPayload"].(map[string]interface{})
	assert.Equal(t, 2.0, objData["count"])
}
