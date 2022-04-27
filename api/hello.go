package api

import (
	"net/http"
)

func HelloEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HelloEndpoint"))
}
