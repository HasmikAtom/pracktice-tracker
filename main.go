package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/soheilhy/cmux"
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

	grpcServer := newApiServer(db)
	if err := api.RegisterApiHandlerServer(ctx, mux, grpcServer); err != nil {
		log.Fatal(err)
	}

	httpServer := http.NewServeMux()
	httpServer.Handle("/", mux)

	fs := http.FileServer(http.Dir(swaggerFolderPath))
	httpServer.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))

	httpServer.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerFilePath)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Panic(err)
	}

	m := cmux.New(l)

	grpcl := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)
	httpl := m.Match(cmux.HTTP1Fast())
	go serveGRPC(grpcl, grpcServer)
	go serveHTTP(httpl, httpServer)

	if err := m.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		panic(err)
	}
}
