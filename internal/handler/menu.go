package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/matryer/way"

	"ovto/internal/service"
)

type CategoryInput struct {
	Label        string `json:"label"`
	Availability bool   `json:"availability"`
}

type ItemInput struct {
	Name         string  `json:"name"`
	Category     int64   `json:"category"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Availability bool    `json:"availability"`
}

func (h *handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in CategoryInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	err := h.CreateCategory(ctx, rID, in.Label, in.Availability)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrTitleTaken {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err == service.ErrRestaurantNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) getCategoriesByRestaurant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	c, err := h.GetCategoriesByRestaurant(ctx, rID)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, c, http.StatusOK)
}

func (h *handler) createItem(w http.ResponseWriter, r *http.Request) {
	var in ItemInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	fmt.Println(in.Category)
	err := h.CreateItem(ctx, rID, in.Category, in.Name, in.Description, in.Price, in.Availability)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrItemAlreadyExists {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) getMenuForFp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	restaurants, err := h.GetMenuForFp(ctx, rID)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrRestaurantNotFound || err == service.ErrMenuNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, restaurants, http.StatusOK)
}
