package socketgo

import (
	"encoding/json"
	"net/http"
)

func Response(w http.ResponseWriter, data string, status int) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
