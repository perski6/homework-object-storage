package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/perski6/homework-object-storage/config"
	"github.com/perski6/homework-object-storage/services/storage"
)

func GetObject(w http.ResponseWriter, r *http.Request, minio storage.Service, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.App.Timeout)*time.Millisecond)
	defer cancel()
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {

	}

	object, err := minio.GetObject(ctx, id)
	if err != nil {

	}
	w.WriteHeader(http.StatusOK)
	w.Write(object)
}
