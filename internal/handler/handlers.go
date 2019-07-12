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

	api := way.NewRouter()
	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("POST", "/facebook", h.facebookAuth)
	api.HandleFunc("GET", "/auth_user", h.authUser)
	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("PUT", "/users", h.updateUser)
	api.HandleFunc("PUT", "/auth_user/dp", h.updateDisplayPicture)

	fs := http.FileServer(&spaFileSystem{http.Dir("web/static")})
	//if inLocalhost {
	//	fs = withoutCache(fs)
	//}

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(api)))
	r.Handle("GET", "/...", fs)

	return r
}
