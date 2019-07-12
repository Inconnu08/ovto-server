package handler

import (
	"encoding/json"
	"net/http"
	"ovto/internal/service"
)

type createFoodProviderInput struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *handler) createFoodProvider(w http.ResponseWriter, r *http.Request) {
	var in createFoodProviderInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateFoodProvider(r.Context(), in.Email, in.Fullname, in.Phone, in.Password)
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
