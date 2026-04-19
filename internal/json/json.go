package json

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func Write(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	json.NewEncoder(writer).Encode(data)
}

func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
