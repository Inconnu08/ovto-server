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
	userApi.HandleFunc("POST", "/:restaurant_id/order", h.createUserOrder)

	foodProviderApi := way.NewRouter()
	foodProviderApi.HandleFunc("POST", "/users", h.createFoodProvider)
	foodProviderApi.HandleFunc("POST", "/login", h.foodProviderLogin)
	foodProviderApi.HandleFunc("GET", "/auth_fp", h.authFp)
	foodProviderApi.HandleFunc("GET", "/restaurants", h.getRestaurants)
	foodProviderApi.HandleFunc("POST", "/restaurants/:restaurant_id/role", h.createRole)

	restaurantApi := way.NewRouter()
	restaurantApi.HandleFunc("POST", "/", h.createRestaurant)
	restaurantApi.HandleFunc("PUT", "/:restaurant_id", h.updateRestaurant)
	restaurantApi.HandleFunc("PUT", "/:restaurant_id/dp", h.updateRestaurantDisplayPicture)
	restaurantApi.HandleFunc("PUT", "/:restaurant_id/cover", h.updateRestaurantCoverPicture)
	restaurantApi.HandleFunc("POST", "/:restaurant_id/gallery", h.createRestaurantGalleryPicture)
	restaurantApi.HandleFunc("GET", "/:restaurant_id/gallery", h.getRestaurantGallery)
	restaurantApi.HandleFunc("DELETE", "/:restaurant_id/gallery/:image", h.deleteRestaurantGalleryPicture)
	restaurantApi.HandleFunc("POST", "/:restaurant_id/offers", h.createRestaurantOffersPicture)
	restaurantApi.HandleFunc("DELETE", "/:restaurant_id/offers/:image", h.deleteRestaurantOffersPicture)
	restaurantApi.HandleFunc("POST", "/:restaurant_id/category", h.createCategory)
	restaurantApi.HandleFunc("GET", "/:restaurant_id/category", h.getCategoriesByRestaurant)
	restaurantApi.HandleFunc("POST", "/:restaurant_id/menu", h.createItem)
	restaurantApi.HandleFunc("GET", "/:restaurant_id/menu", h.getMenuForFp)
	restaurantApi.HandleFunc("GET", "/:restaurant_id/orders", h.getOrders)

	fs := http.FileServer(&spaFileSystem{http.Dir("web/static")})
	//if inLocalhost {
	//	fs = withoutCache(fs)
	//}

	r := way.NewRouter()
	r.Handle("*", "/api/fp...", http.StripPrefix("/api/fp", h.withFpAuth(foodProviderApi)))
	r.Handle("*", "/api/restaurants...", http.StripPrefix("/api/restaurants", h.withFpAuth(restaurantApi)))
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(userApi)))
	r.Handle("GET", "/...", fs)

	return r
}
