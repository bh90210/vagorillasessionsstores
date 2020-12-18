package main

import (
	"log"
	"net/http"
	"time"

	stores "github.com/bh90210/vagorillasessionsstores"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	store, err := stores.NewDgraphStore(conn, []byte("SESSION_KEY"))
	if err != nil {
		log.Fatal("eee", err)
	}

	store.Options.HttpOnly = false
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteStrictMode
	store.Options.MaxAge = 5

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "mssgng-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if session.Values["test"] == nil {
			session.Values["test"] = true
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		if session.Values["test"] == true {
			w.Write([]byte("OK"))
			return
		}

	}).Methods(http.MethodGet)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8085",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
