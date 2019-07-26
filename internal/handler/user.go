package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ovto/internal/service"
)

type createUserInput struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserInput struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type ChangePasswordInput struct {
	OldPassword   string `json:"old_password"`
	NewPassword   string `json:"new_password"`
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var in createUserInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateUser(r.Context(), in.Email, in.Fullname, in.Password)
	if err == service.ErrInvalidEmail || err == service.ErrInvalidFullname {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrEmailTaken {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var in updateUserInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.UpdateUser(r.Context(), in.Address, in.Phone)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.WriteHeader(http.StatusOK)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	err := h.DeleteUser(r.Context())
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) updateDisplayPicture(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxAvatarBytes)
	defer r.Body.Close()
	avatarURL, err := h.UpdateDisplayPicture(r.Context(), r.Body)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrUnsupportedPictureFormat {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	fmt.Fprint(w, avatarURL)
}

func (h *handler) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	var in ChangePasswordInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.ChangeUserPassword(r.Context(), in.OldPassword, in.NewPassword)
	if err == service.ErrUnauthenticated || err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.WriteHeader(http.StatusOK)
}