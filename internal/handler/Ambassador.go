package handler

import (
	"encoding/json"
	"net/http"

	"ovto/internal/service"
)

type createAmbassadorInput struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Facebook string `json:"facebook"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

func (h *handler) createAmbassador(w http.ResponseWriter, r *http.Request) {
	var in createAmbassadorInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateAmbassador(r.Context(), in.Email, in.Fullname, in.Phone,in.Facebook, in.City, in.Area, in.Address, in.Password)
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
