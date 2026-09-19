package utility

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

func ToUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(ids))

	for _, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}
