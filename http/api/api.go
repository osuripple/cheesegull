package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type _base struct {
	Ok      bool
	Message string `json:",omitempty"`
}

// An Error oCcured. aeo sounded like the name of a JavaScript framework,
// or something in Neapolitan.
var aec = _base{
	Message: "An error occurred.",
}

// j writes JSON to the response writer
func j(w http.ResponseWriter, code int, obj interface{}) {
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		fmt.Println(err)
	}
}
