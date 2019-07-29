package handler

import (
	"encoding/json"
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

func (h *handler) createRestaurant(w http.ResponseWriter, r *http.Request) {
	var in createRestaurantInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateFoodProvider(r.Context(), )
	if err == service.ErrInvalidEmail || err == service.ErrInvalidFullname || err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrEmailTaken || err == service.ErrPhoneNumberTaken {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
