package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"bakery-scheduler/internal/entity"
	"bakery-scheduler/pkg/response"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware memeriksa header Authorization: Bearer <token>.
// Token dibuat saat login (HS256, berlaku 12 jam).
func (s *Server) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.Parse(raw, func(*jwt.Token) (interface{}, error) { return s.JWTSecret, nil },
			jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if raw == "" || err != nil {
			resp := response.Response{}
			resp.SetError(errors.New("perlu login"), http.StatusUnauthorized)
			resp.RenderJSON(w, r)
			return
		}

		// simpan claims (sub, name, email) ke context, dibaca handler (mis. penanggung jawab order)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctxVal := entity.ContextValue{M: map[string]interface{}{"sub": claims["sub"], "name": claims["name"], "email": claims["email"]}}
			r = r.WithContext(context.WithValue(r.Context(), entity.ContextKey("claims"), ctxVal))
		}

		next.ServeHTTP(w, r)
	})
}
