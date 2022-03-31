package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/hamoz/uxsdp-facebook/common"
	fb "github.com/hamoz/uxsdp-facebook/fb/handler"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rapidUrl := os.Getenv("RAPID_URL")
	var port int64
	if port, err = strconv.ParseInt(os.Getenv("port"), 10, 64); err != nil {
		port = 8119
	}

	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	appSecret := os.Getenv("FB_APP_SECRET")
	verifyToken := os.Getenv("FB_VERIFY_TOKEN")
	rapridProApi := &common.RapidProApi{Url: rapidUrl}
	facebook := fb.NewHandler(rapridProApi, verifyToken, appSecret, accessToken)
	r := mux.NewRouter()
	rapidRouter := r.PathPrefix("/rp").Subrouter()
	rapidRouter.Path("/{ChannelType}/{ChannelId}/receive").HandlerFunc(facebook.HandleIncoming)
	rapidRouter.Path("/{ChannelType}/{AppId}/send").HandlerFunc(facebook.HandleOutgoing)
	http.Handle("/", r)
	errs := make(chan error, 2)
	go func() {
		log.Printf("Listening on port %d", port)
		errs <- http.ListenAndServe(":"+strconv.Itoa(int(port)), nil)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("terminated %s", <-errs)
}
