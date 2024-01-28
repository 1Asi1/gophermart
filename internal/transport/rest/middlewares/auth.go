package middlewares

import (
	"net/http"

	"github.com/1Asi1/gophermart/internal/service"
)

func Authorization(next http.HandlerFunc, service service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		id, err := service.CheckAccess(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add("ID", id)
		next.ServeHTTP(w, r)
	}
}
