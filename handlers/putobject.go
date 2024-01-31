package handlers

import (
	"context"
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

	}

	body, err := io.ReadAll(r.Body)
	if err != nil {

	}

	err = minio.PutObject(ctx, id, body)
	if err != nil {

	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Object created")
}
