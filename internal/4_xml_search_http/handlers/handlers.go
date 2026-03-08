package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/service"
)

type Handler struct {
	service service.Servicer
	timeout time.Duration
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.SearchServer(w, req)
}

func NewHandler(service service.Servicer, timeout time.Duration) *Handler {
	return &Handler{
		service: service,
		timeout: timeout,
	}
}

func (h *Handler) SearchServer(w http.ResponseWriter, r *http.Request) {
	AccessToken := r.Header[http.CanonicalHeaderKey("AccessToken")]
	if len(AccessToken) == 0 || AccessToken[0] == "" {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	query := r.FormValue("query")

	orderField := r.FormValue("order_field")
	orderByStr := r.FormValue("order_by")
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil {
		http.Error(w, "incorrect conversion order_by to int", http.StatusBadRequest)
		return
	}

	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		http.Error(w, "incorrect conversion limit to int", http.StatusBadRequest)
		return
	}
	offsetStr := r.FormValue("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		http.Error(w, "incorrect conversion offset to int", http.StatusBadRequest)
		return
	}

	type resultType struct {
		data []models.User
		err  *service.ServiceError
	}
	resultCh := make(chan resultType)

	go func() {
		data, err := h.service.Users(query, orderField, orderBy, offset, limit)
		resultCh <- resultType{data: data, err: err}
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "Timeout", http.StatusRequestTimeout)
	case result := <-resultCh:
		if result.err != nil {
			switch result.err.TypeError {
			case service.InternalError:
				http.Error(w, result.err.Error(), http.StatusInternalServerError)
				return

			case service.BadRequest:
				w.WriteHeader(http.StatusBadRequest)
				response := models.SearchErrorResponse{
					Error: result.err.Error(),
				}
				bytes, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "Unable to marshal response to json", http.StatusInternalServerError)
					return
				}

				_, _ = w.Write(bytes)
				return
			}
		}

		bytes, err := json.Marshal(result.data)
		if err != nil {
			http.Error(w, "Unable to marshal users to json", http.StatusInternalServerError)
		}
		_, _ = w.Write(bytes)
	}
}
