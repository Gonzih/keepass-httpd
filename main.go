package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tobischo/gokeepasslib"
)

func initViper() {
	pflag.String("keepass-password", "", "KeepassDB password")
	pflag.String("keepass-file", "", "KeepassDB file path")
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

	log.Fatal(http.ListenAndServe(":8080", router))
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

func findEntry(values map[string]string) (*gokeepasslib.Entry, error) {
	sharedGroupLock.RLock()
	defer sharedGroupLock.RUnlock()
	return findInGroupByValues(&sharedGroup, values)
}
