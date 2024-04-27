package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/izaakdale/dinghy-agent/api/v1"
	"github.com/izaakdale/ittp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.Dial(os.Getenv("DINGHY_AGENT_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := v1.NewAgentClient(conn)

	mux := ittp.NewServeMux()

	mux.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var req v1.InsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := client.Insert(r.Context(), &req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	mux.Get("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		req := v1.FetchRequest{Key: key}

		resp, err := client.Fetch(r.Context(), &req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Delete("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		req := v1.DeleteRequest{Key: key}

		_, err := client.Delete(r.Context(), &req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), mux)
}
