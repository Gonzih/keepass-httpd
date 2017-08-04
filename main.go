package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initViper() {
	pflag.String("keepass-password", "", "KeepassDB password")
	pflag.String("keepass-file", "", "KeepassDB file path")
	pflag.Int("http-port", 8080, "Port to listen on")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

func init() {
	initViper()

	err := loadDB()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := httprouter.New()

	router.GET("/search", SearchHandler)
	router.POST("/reload", ReloadHandler)

	addr := fmt.Sprintf(":%d", viper.GetInt("http-port"))
	log.Fatal(http.ListenAndServe(addr, router))
}

func respondWithError(w http.ResponseWriter, err error, status int) {
	response := errorResponse{Error: err.Error()}
	json, err := json.Marshal(&response)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
