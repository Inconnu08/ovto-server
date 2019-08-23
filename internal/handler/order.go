package handler

import (
	"encoding/json"
	"github.com/matryer/way"
	"mime"
	"net/http"
	"ovto/internal/service"
)

type OrderInput struct {
	CId    int64
	Status int64
	Items  map[string]int64
}

func (h *handler) ordersStream(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		respondErr(w, errStreamingUnsupported)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	oo, err := h.OrdersStream(ctx, rID)
	if err == service.ErrInvalidRestaurantId {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	header := w.Header()
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Content-Type", "text/event-stream; charset=utf-8")

	for o := range oo {
		writeSSE(w, o)
		f.Flush()
	}
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in OrderInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	err := h.CreateOrder(ctx, rID, in.CId, in.Status, in.Items)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound || err == service.ErrInvalidRestaurantId {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createUserOrder(w http.ResponseWriter, r *http.Request) {
	var in OrderInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	err := h.CreateUserOrder(ctx, rID, in.Status, in.Items)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound || err == service.ErrInvalidRestaurantId {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) getOrders(w http.ResponseWriter, r *http.Request) {
	if a, _, err := mime.ParseMediaType(r.Header.Get("Accept")); err == nil && a == "text/event-stream" {
		h.ordersStream(w, r)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	o, err := h.GetOrders(ctx, rID)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound || err == service.ErrInvalidRestaurantId {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, o, http.StatusOK)
}