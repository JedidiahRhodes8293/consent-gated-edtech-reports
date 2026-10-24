package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

var deadlines = []Deadline{
	{Course: "algebra", Learner: "learner-17", DueDate: "2026-11-03", Complete: false},
	{Course: "biology", Learner: "learner-17", DueDate: "2026-11-07", Complete: true},
}

func main() {
	client, err := NewConsentClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		granted, checkErr := client.Check(r.Context(), "learner-17", "analytics")
		if checkErr != nil {
			http.Error(w, checkErr.Error(), http.StatusBadGateway)
			return
		}
		result, decisionErr := ReportingDecision(granted, deadlines)
		if decisionErr != nil {
			http.Error(w, decisionErr.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	http.HandleFunc("/consent/grant", consentHandler(client.Grant, "analytics"))
	http.HandleFunc("/consent/revoke", consentHandler(client.Revoke, "analytics"))
	log.Println("edtech consent service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func consentHandler(change func(context.Context, string, string) error, category string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := change(r.Context(), "learner-17", category); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
