package api

import (
	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type API struct {
	authService     ports.AuthService
	ordersService   ports.OrdersService
	balanceService  ports.BalanceService
	withdrawService ports.WithdrawService
	logger          common.Logger
}

func NewAPI(
	authService ports.AuthService,
	ordersService ports.OrdersService,
	balanceService ports.BalanceService,
	withdrawService ports.WithdrawService,
	logger common.Logger,
) *API {
	return &API{
		authService:     authService,
		ordersService:   ordersService,
		balanceService:  balanceService,
		withdrawService: withdrawService,
		logger:          logger,
	}
}

func (api *API) Routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(GzipHandler)
	router.Use(middleware.Compress(5, "text/html", "application/json"))
	router.Use(api.CookieAuth([]string{"/api/user/register", "/api/user/login"}))

	router.Post("/api/user/register", api.Register)
	router.Post("/api/user/login", api.Login)
	router.Post("/api/user/orders", api.CreateOrder)
	router.Get("/api/user/orders", api.GetOrders)
	router.Get("/api/user/balance", api.GetBalance)
	router.Post("/api/user/balance/withdraw", api.Withdraw)
	router.Get("/api/user/withdrawals", api.GetWithdrawals)

	return router
}
