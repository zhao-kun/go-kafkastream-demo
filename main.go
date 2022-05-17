package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/burdiyan/kafkautil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	goka "github.com/lovoo/goka"
	"github.com/lovoo/goka/storage"
)

var (
	brokerStr  = ""
	topic      = ""
	serverAddr = ""
)

func init() {
	flag.StringVar(&brokerStr, "brokers", "127.0.0.1:9092", "brokers' address of the kafka, default is 127.0.0.1:9092")
	flag.StringVar(&topic, "topic", "labels", `topic of the kafka, topic should create with  cleanup.policy=compact
	kafka-topics.sh --topic labels --create  \
--replication-factor 1 --partitions 2  \
--config cleanup.policy=compact \
 --bootstrap-server :9092
	`)
	flag.StringVar(&serverAddr, "server", "127.0.0.1:9527", "web server endpoint, default is 127.0.0.1:9527")
}

func main() {
	flag.Parse()
	brokers := strings.Split(brokerStr, ",")
	view, err := goka.NewView(brokers, goka.Table(topic), newCoder(), goka.WithViewStorageBuilder(storage.MemoryBuilder()), goka.WithViewHasher(kafkautil.MurmurHasher))
	if err != nil {
		log.Printf("create goka.view from topic %s at broker %s:\n\t%s", topic, brokerStr, err)
		os.Exit(-1)
	}
	ctx := context.Background()
	go startView(ctx, view)
	startWeb(view)

	c := make(chan struct{})
	d := <-c
	log.Printf("programming exit %+v", d)
}

func startView(ctx context.Context, view *goka.View) {
	for {
		if err := view.Run(ctx); err != nil {
			log.Printf("view run error: %s", err)
		}
	}
}

func adaptQueryLabels(view *goka.View) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		i, err := view.Iterator()
		if err != nil {
			log.Printf("can't iterator ktable error:%s\n", err)
			responseError(w, http.StatusInternalServerError, fmt.Sprintf("can't interator view, error:%s", err))
			return
		}

		labels := []*Label{}
		for i.Next() {
			value, err := i.Value()
			if err != nil {
				responseError(w, http.StatusInternalServerError, fmt.Sprintf("get value from iterator error, error:%s", err))
				return
			}
			log.Printf("iterator key is :%s, value is %+v\n", i.Key(), value)
			if value == nil {
				continue
			}
			label := value.(*Label)
			if label == nil {
				responseError(w, http.StatusInternalServerError, fmt.Sprintf("expect value is a pointer of the Label struct, but is %+v", reflect.TypeOf(value)))
				return
			}
			labels = append(labels, label)
		}
		b, err := json.Marshal(labels)
		if err != nil {
			responseError(w, http.StatusInternalServerError, fmt.Sprintf("marshal labels error:%s", err))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(b)
		return
	}
}
func adaptQueryLabelsByID(view *goka.View) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if !view.Recovered() {
			log.Printf("recover isn't true\n")
		}
		result, err := view.Get(id)
		if result == nil || err != nil {
			log.Printf("get key [%s] error:%s\n", id, err)
			responseError(w, http.StatusNotFound, "id: ["+id+"] was not found in the view")
			return
		}
		label := result.(*Label)
		if label == nil {
			responseError(w, http.StatusInternalServerError, fmt.Sprintf("expect value is a pointer of the Label struct, but is %+v", reflect.TypeOf(result)))
			return
		}
		b, err := json.Marshal(label)
		if err != nil {
			responseError(w, http.StatusInternalServerError, fmt.Sprintf("marshal labels error:%s", err))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(b)
		return
	}
}

func responseError(w http.ResponseWriter, status int, desc string) {
	w.WriteHeader(status)
	w.Write([]byte(desc))
}

func startWeb(view *goka.View) {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.StripSlashes)

	r.Route("/api/v1/labels", func(r chi.Router) {
		r.Get("/", adaptQueryLabels(view))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", adaptQueryLabelsByID(view))
		})
	})

	server, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("[Web] Failed to start the http server: %s", err)
	}
	log.Printf("[Web] HTTP server is listening on %s\n", serverAddr)

	// Start the http server
	go func() {
		if err := http.Serve(server, r); err != nil {
			log.Printf("[Web] HTTP server error: %s", err)
		}
	}()
}
