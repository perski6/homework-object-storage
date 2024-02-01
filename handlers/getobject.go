package handlers

import (
	"context"
	"errors"
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	object, err := minio.GetObject(ctx, id)
	if err != nil {
		logger.Error("error getObject", err)

		if errors.Is(err, storage.InstanceNotAccessible) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(object)
}
