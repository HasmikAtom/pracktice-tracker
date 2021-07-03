package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"

	api "github.com/HasmikAtom/tracker/v1"
)

var (
	swaggerFilePath   = "./v1/service.swagger.json"
	swaggerFolderPath = "./swaggerui"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL unset")
	}
	dbConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := pgxpool.ConnectConfig(ctx, dbConfig)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	db := newDatabase(conn)

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   false,
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "X-Atom-User":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	server := newApiServer(db)
	if err := api.RegisterApiHandlerServer(ctx, mux, server); err != nil {
		log.Fatal(err)
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/", mux)

	fs := http.FileServer(http.Dir(swaggerFolderPath))
	rootMux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))

	rootMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerFilePath)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Println("running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, rootMux))
}
