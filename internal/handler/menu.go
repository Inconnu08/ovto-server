package handler

import (
	"net/http"

	"github.com/matryer/way"

	"ovto/internal/service"
)

type CategoryInput struct {
	Name         string `json:"name"`
	Availability bool   `json:"availability"`
}

type ItemInput struct {
	Name         string  `json:"name,omitempty"`
	Category     string  `json:"category,omitempty"`
	Description  string  `json:"description,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Availability bool    `json:"availability,omitempty"`
}

func (h *handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in CategoryInput

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	err := h.CreateCategory(ctx, rID, in.Name, in.Availability)
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

	respond(w, nil, http.StatusOK)
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

	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

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

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, nil, http.StatusOK)
}

func (h *handler) getMenuForFp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rID := way.Param(ctx, "restaurant_id")

	restaurants, err := h.GetMenuForFp(ctx, rID)
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

	respond(w, restaurants, http.StatusOK)
}
