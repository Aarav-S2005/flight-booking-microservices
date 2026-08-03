package utility

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ConvertJSONToStruct(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return nil
}

func ConvertStructToJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
