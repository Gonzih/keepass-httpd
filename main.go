package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/tobischo/gokeepasslib"
)

type responseEntry struct {
	UserName string `json:"username"`
	Title    string `json:"title"`
	Password string `json:"password"`
}

type errorResponseEntry struct {
	Error string `json:"error"`
}

func GetUserName(entry *gokeepasslib.Entry) string {
	return entry.Get("UserName").Value.Content
}

func findInGroupByUserName(group *gokeepasslib.Group, userName string) (*gokeepasslib.Entry, error) {
	for _, entry := range group.Entries {
		if userName == GetUserName(&entry) {
			return &entry, nil
		}
	}

	for _, innerGroup := range group.Groups {
		entry, err := findInGroupByUserName(&innerGroup, userName)

		if err == nil {
			return entry, err
		}
	}

	return nil, errors.New("Entry not found")
}

func marshalEntry(entry *gokeepasslib.Entry) ([]byte, error) {
	response := responseEntry{
		UserName: GetUserName(entry),
		Title:    entry.GetTitle(),
		Password: entry.GetPassword()}

	return json.Marshal(&response)
}

var sharedGroup gokeepasslib.Group

func loadDB() {
	file, _ := os.Open("test.kdbx")
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials("test")
	_ = gokeepasslib.NewDecoder(file).Decode(db)

	db.UnlockProtectedEntries()

	sharedGroup = db.Content.Root.Groups[0]
}

func init() {
	loadDB()
}

func main() {
	router := httprouter.New()

	router.GET("/search", SearchHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func respondWithError(w http.ResponseWriter, err error) {
	response := errorResponseEntry{Error: err.Error()}
	json, err := json.Marshal(&response)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := r.FormValue("username")

	if len(username) == 0 {
		respondWithError(w, errors.New("Username parameter is required"))
		return
	}

	entry, err := findInGroupByUserName(&sharedGroup, username)

	if err != nil {
		respondWithError(w, err)
		return
	}

	json, err := marshalEntry(entry)

	if err != nil {
		respondWithError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
