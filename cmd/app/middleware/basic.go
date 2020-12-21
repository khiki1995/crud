package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func Basic(isManagerExist func (login, password string) bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			auth := strings.SplitN(request.Header.Get("Authorization"), " ", 2)

			if len(auth) != 2 || auth[0] != "Basic" {
				http.Error(writer, "authorization failed", http.StatusUnauthorized)
				return
			}

			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) != 2 || !isManagerExist(pair[0], pair[1]) {
				http.Error(writer, "authorization failed", http.StatusUnauthorized)
				return
			}

			handler.ServeHTTP(writer, request)
		})
	}
}