package handler

import (
	"net/http"

	"github.com/matryer/way"

	"ovto/internal/service"
)

type handler struct {
	*service.Service
}

func New(s *service.Service) http.Handler {
	h := &handler{s}

	userApi := way.NewRouter()
	userApi.HandleFunc("POST", "/login", h.userLogin)
	userApi.HandleFunc("POST", "/facebook", h.facebookAuth)
	userApi.HandleFunc("GET", "/auth_user", h.authUser)
	userApi.HandleFunc("POST", "/users", h.createUser)
	userApi.HandleFunc("PUT", "/users", h.updateUser)
	userApi.HandleFunc("DELETE", "/users", h.deleteUser)
	userApi.HandleFunc("PUT", "/auth_user/dp", h.updateDisplayPicture)

	foodProviderApi := way.NewRouter()
	foodProviderApi.HandleFunc("POST", "/users", h.createFoodProvider)
	foodProviderApi.HandleFunc("POST", "/login", h.foodProviderLogin)
	foodProviderApi.HandleFunc("GET", "/auth_fp", h.authFp)

	fs := http.FileServer(&spaFileSystem{http.Dir("web/static")})
	//if inLocalhost {
	//	fs = withoutCache(fs)
	//}

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(userApi)))
	r.Handle("*", "/api/fp...", http.StripPrefix("/api/fp", h.withFpAuth(foodProviderApi)))
	r.Handle("GET", "/...", fs)

	return r
}
