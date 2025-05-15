package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	errNoAuth        = errors.New("no Authorization header")
	errInvalidClaims = errors.New("invalid token claims")
)

func (s Server) handleGetSwipes(w http.ResponseWriter, r *http.Request) {
	userID, err := getJWTUserID(r)
	if err != nil {
		s.log.WarnContext(r.Context(), "no auth user", slog.Any("error", err))
		errStatusCode(w, http.StatusUnauthorized)

		return
	}

	var dto geoData

	err = json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		s.log.DebugContext(r.Context(), "could not unmarshal geoData", slog.Any("error", err))

		errStatusCode(w, http.StatusBadRequest)

		return
	}

	err = s.repo.UpsertGeoData(r.Context(), userID, dto.Long, dto.Lat)
	if err != nil {
		s.log.ErrorContext(r.Context(), "failed to upsert geo data", slog.Any("error", err))
		errStatusCode(w, http.StatusInternalServerError)

		return
	}
}

func (s Server) handleHealthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getJWTUserID(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil, errNoAuth
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse jwt: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errInvalidClaims
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get subject from jwt: %w", err)
	}

	return uuid.Parse(sub)
}

type geoData struct {
	Long float64 `json:"longitude"`
	Lat  float64 `json:"latitude"`
}

func errStatusCode(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
