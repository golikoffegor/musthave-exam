package router

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/golikoffegor/musthave-exam/internal/handler"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
)

// * `POST /api/user/register` — регистрация пользователя;
// * `POST /api/user/login` — аутентификация пользователя;
// * `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
// * `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// * `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
// * `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// * `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.
func InitRouter(handler handler.Handler) chi.Router {
	r := chi.NewRouter()
	// r.Use(packmiddleware.Logger)

	// r.Use(middleware.Logger)
	// r.Use(packmiddleware.ContentTypeSet("application/json"))
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Recoverer)
	// r.Use(packmiddleware.GzipMiddleware)

	r.Head("/", func(rw http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "Content-Type: application/json")
	})
	// app.HandleRequest()
	// r.Get("/", handler.MetricsHandlerFunc)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.RegisterHandler)
		r.Post("/login", handler.LoginUserHandler)
		r.Post("/orders", handler.AddOrderHandler)
		r.Get("/orders", handler.GetOrdersHandler)
		r.Get("/balance", handler.GetBalanceHandler)
		r.Get("/withdrawals", handler.GetWithdrawalsHandler)
		r.Post("/balance/withdraw", handler.WithdrawHandler)
	})

	FileServer(r, "/docs", http.Dir("./docs"))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.yaml"), // Ссылка на ваш swagger.json
	))

	r.Get("/metrics", prometheusHandler())

	return r
}

func prometheusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/docs/", http.FileServer(root)).ServeHTTP(w, r)
	})
}
