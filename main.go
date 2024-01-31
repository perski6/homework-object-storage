package main

import (
	"hash/fnv"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/perski6/homework-object-storage/consistentHash"
	"github.com/perski6/homework-object-storage/handlers"
	"github.com/perski6/homework-object-storage/services/eventwatcher"
	"github.com/perski6/homework-object-storage/services/storage"
)

func main() {
	ch := consistentHash.New[*minio.Client](hasher32a{})
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	watcher := eventwatcher.New(logger, ch)
	watcher.DiscoverInstances()
	go watcher.Watch()

	minioStorage := storage.New(ch)
	router := mux.NewRouter()

	router.HandleFunc("/object/{id:[a-zA-Z0-9]{1,32}}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetObject(w, r, *minioStorage, logger)
	}).Methods("GET")

	router.HandleFunc("/object/{id:[a-zA-Z0-9]{1,32}}", func(w http.ResponseWriter, r *http.Request) {
		handlers.PutObject(w, r, *minioStorage, logger)
	}).Methods("PUT")

	http.Handle("/", router)
	http.ListenAndServe(":3000", nil)
}

type hasher32a struct {
}

func (h hasher32a) Hash(key string) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return int(hash.Sum32())
}
