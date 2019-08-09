package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"ovto/internal/service"
)

type loginInput struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type thirdpartyAuthInput struct {
	AccessToken    string `json:"access_token"`
}

func (h *handler) authUser(w http.ResponseWriter, r *http.Request) {
	u, err := h.AuthUser(r.Context())
	if err == service.ErrUnauthenticated {
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

	respond(w, u, http.StatusOK)
}

func (h *handler) authFp(w http.ResponseWriter, r *http.Request) {
	u, err := h.AuthFp(r.Context())
	if err == service.ErrUnauthenticated {
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

	respond(w, u, http.StatusOK)
}

func (h *handler) userLogin(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := h.UserLogin(r.Context(), in.Phone, in.Password)
	if err == service.ErrUnimplemented {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	if err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, out, http.StatusOK)
}

func (h *handler) foodProviderLogin(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := h.FoodProviderLogin(r.Context(), in.Phone, in.Password)
	if err == service.ErrUnimplemented {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	if err == service.ErrInvalidPhone {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, out, http.StatusOK)
}

func (h *handler) facebookAuth(w http.ResponseWriter, r *http.Request) {
	var in thirdpartyAuthInput

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken := "EAAGKzLH7udYBAPEMOE3UjixPJjsbVdV8gUYKEVkMJXYOvw8VSRAlHJZAKbU38qhtNV95RPh5ZBvZCBi4l7eDsak1Heby8eWVNVkG4t2Ev0rjlVu3TJbUVavnaYRLlYwjuuEu4v8zRxFNFWhZA8lvzoxZCZAgZCPHZBrw3LPuQ2Kr2FPH6IXtKOAIyMVCBHBbCSHrt60QUkyTZBQEIIOwbvf2NF9eB0NrdLBCuH4NhlGZAB4AZDZD"
	resp, err := http.Get("https://graph.facebook.com/v3.3/me?fields=id%2Cname%2Cemail%2Cbirthday%2Cpicture&access_token=" + accessToken)
	if err != nil {
		log.Fatalln(err)
	}

	var fbin service.FacebookAuthOutput
	resBytes, _ := ioutil.ReadAll(resp.Body)
	defer r.Body.Close()

	rs := string(resBytes)
	log.Println(rs)
	if strings.Contains(rs, "Error") {
		http.Error(w, rs, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(resBytes, &fbin); err != nil {
		log.Println("FAILED: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("RESPONSE struct:	", fbin.Email, "\nDP: ", fbin.Picture.Data.Url)

	out, err := h.FacebookAuth(r.Context(), fbin)
	if err == service.ErrUnimplemented {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	if err == service.ErrInvalidEmail {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondErr(w, err)
		return
	}

	respond(w, out, http.StatusOK)
}

func (h *handler) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if !strings.HasPrefix(a, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		token := a[7:]
		uid, err := h.AuthUserID(token)
		log.Println(token, "=", uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, service.KeyAuthUserID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) withFpAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if !strings.HasPrefix(a, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		token := a[7:]
		uid, err := h.AuthFpID(token)
		log.Println(token, "=", uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, service.KeyAuthFoodProviderID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) withAmbassadorAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if !strings.HasPrefix(a, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		token := a[7:]
		uid, err := h.AuthAmbassadorID(token)
		log.Println(token, "=", uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, service.KeyAuthFoodProviderID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
