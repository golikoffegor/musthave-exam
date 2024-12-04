package router

import (
	"musthave-exam/internal/handler"
	"net/http"

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
		r.Post("/login", handler.LoginHandler)
		r.Post("/orders", handler.AddOrderHandler)
		r.Get("/orders", handler.GetOrdersHandler)
		r.Get("/balance", handler.GetBalanceHandler)
		r.Get("/withdrawals", handler.GetWithdrawalsHandler)
		r.Post("/balance/withdraw", handler.WithdrawHandler)
	})

	// r.Get("/swagger/*", httpSwagger.Handler(
	// 	httpSwagger.URL("./doc.json"), // Ссылка на ваш swagger.json
	// ))

	// log.Fatal(http.ListenAndServe(flags.Parse(), r))
	return r
}
