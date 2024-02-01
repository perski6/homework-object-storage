package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/perski6/homework-object-storage/config"
	"github.com/perski6/homework-object-storage/services/storage"
)

func PutObject(w http.ResponseWriter, r *http.Request, minio storage.Service, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.App.Timeout)*time.Millisecond)
	defer cancel()
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("error reading body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = minio.PutObject(ctx, id, body)
	if err != nil {
		logger.Error("error putObject", err)

		if errors.Is(err, storage.InstanceNotAccessible) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Object created")
}
