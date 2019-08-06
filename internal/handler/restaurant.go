package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ovto/internal/service"
)

type createRestaurantInput struct {
	Title       string `json:"title"`
	About       string `json:"about,omitempty"`
	Phone       string `json:"phone"`
	Location    string `json:"location"`
	City        string `json:"city"`
	Area        string `json:"area"`
	Country     string `json:"country"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
	Referral    string `json:"referral,omitempty"`
}

type updateRestaurantInput struct {
	Id			string `json:"id"`
	About       string `json:"about,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Location    string `json:"location,omitempty"`
	City        string `json:"city,omitempty"`
	Area        string `json:"area,omitempty"`
	Country     string `json:"country,omitempty"`
}

func (h *handler) createRestaurant(w http.ResponseWriter, r *http.Request) {
	var in createRestaurantInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateRestaurant(r.Context(), in.Title, in.About, in.Phone, in.Location, in.City, in.Area, in.Country, in.OpeningTime, in.ClosingTime)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrTitleTaken {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) updateRestaurant(w http.ResponseWriter, r *http.Request) {
	var in updateRestaurantInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.UpdateRestaurant(r.Context(), in.Id, in.About, in.Phone, in.Location, in.City, in.Area)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrTitleTaken {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) getRestaurants(w http.ResponseWriter, r *http.Request) {
	var in createRestaurantInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	restaurants, err := h.GetRestaurantsByFp(r.Context())
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

func (h *handler) updateRestaurantDisplayPicture(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxImageBytes)
	defer r.Body.Close()
	imageURL, err := h.UpdateRestaurantDisplayPicture(r.Context(), r.Body)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrUnsupportedImageFormat {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	fmt.Fprint(w, imageURL)
}