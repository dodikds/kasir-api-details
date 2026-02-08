package handlers

import (
	"database/sql"
	"encoding/json"
	"kasir-api/models"
	"kasir-api/services"
	"net/http"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// multiple item apa aja, quantity nya
func (h *TransactionHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Checkout(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.service.Checkout(req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func GetReportHariIni(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var report models.ReportHariIni

		// Query total revenue & transaksi
		totalQuery := `
		SELECT
		  COALESCE(SUM(td.subtotal), 0),
		  COUNT(DISTINCT t.id)
		FROM transactions t
		JOIN transaction_details td
		  ON t.id = td.transaction_id
		WHERE t.created_at >= (CURRENT_DATE AT TIME ZONE 'Asia/Jakarta') AT TIME ZONE 'UTC'
		  AND t.created_at <  ((CURRENT_DATE + 1) AT TIME ZONE 'Asia/Jakarta') AT TIME ZONE 'UTC'
		`

		err := db.QueryRow(totalQuery).
			Scan(&report.TotalRevenue, &report.TotalTransaksi)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Query produk terlaris
		bestProductQuery := `
		SELECT
		  p.name,
		  SUM(td.quantity)
		FROM transaction_details td
		JOIN transactions t ON t.id = td.transaction_id
		JOIN products p ON p.id = td.product_id
		WHERE t.created_at >= (CURRENT_DATE AT TIME ZONE 'Asia/Jakarta') AT TIME ZONE 'UTC'
		  AND t.created_at <  ((CURRENT_DATE + 1) AT TIME ZONE 'Asia/Jakarta') AT TIME ZONE 'UTC'
		GROUP BY p.id, p.name
		ORDER BY SUM(td.quantity) DESC
		LIMIT 1
		`

		var produk_terlaris models.ProdukTerlaris
		err = db.QueryRow(bestProductQuery).
			Scan(&produk_terlaris.Nama, &produk_terlaris.QtyTerjual)

		if err == sql.ErrNoRows {
			report.ProdukTerlaris = nil
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			report.ProdukTerlaris = &produk_terlaris
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	}
}
