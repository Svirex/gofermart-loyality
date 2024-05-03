package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/api/orders/{number}", func(w http.ResponseWriter, r *http.Request) {
		number := chi.URLParam(r, "number")
		if number == "18" {
			s := `
			{
				"order": "18",
				"status": "INVALID"
			}
			`
			w.Write([]byte(s))
		} else if number == "26" {
			s := `
			{
				"order": "26",
				"status": "PROCESSING"
			}
			`
			w.Write([]byte(s))
		} else if number == "59" {
			s := `No more than 2 requests per minute allowed`
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(s))
		} else if number == "217" {
			w.WriteHeader(http.StatusNoContent)
		} else if number == "1" || number == "2" || number == "3" {
			s := fmt.Sprintf(`
			{
				"order": "%s",
				"status": "INVALID"
			}
			`, number)
			w.Write([]byte(s))
		} else if number == "67" {
			time.Sleep(2 * time.Second) // чтобы checkAccrualService.dbLoader прочитал значение из БД и заново закинул такой же номер в канал
			s := fmt.Sprintf(`
			{
				"order": "%s",
				"status": "PROCESSED",
				"accrual": 500
			}
			`, number)
			w.Write([]byte(s))
		} else {
			s := fmt.Sprintf(`
		{
			"order": "%s",
			"status": "INVALID"
		}
		`, number)
			w.Write([]byte(s))
		}

	})
	http.ListenAndServe(":3000", r)
}
