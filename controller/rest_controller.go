package controller

import (
	"net/http"
	"testtask/service"
)

func HandleHome(page string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, page)
}

func HandleApi(m *service.PriceManager, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fsyms := r.URL.Query()["fsyms"]
	tsyms := r.URL.Query()["tsyms"]

	if fsyms == nil || tsyms == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := m.GetPrices(fsyms, tsyms)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(result)
}
