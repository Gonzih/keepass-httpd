package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tobischo/gokeepasslib"
)

type responseEntry struct {
	UserName string `json:"username"`
	Title    string `json:"title"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Status string `json:"status"`
}

func marshalEntry(entry *gokeepasslib.Entry) ([]byte, error) {
	response := responseEntry{
		UserName: GetUserName(entry),
		Title:    entry.GetTitle(),
		Password: entry.GetPassword(),
		URL:      GetURL(entry)}

	return json.Marshal(&response)
}

func findEntry(values map[string]string) (*gokeepasslib.Entry, error) {
	sharedGroupLock.RLock()
	defer sharedGroupLock.RUnlock()
	return findInGroupByValues(&sharedGroup, values)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := r.FormValue("username")
	title := r.FormValue("title")
	url := r.FormValue("url")

	if len(username) == 0 && len(title) == 0 && len(url) == 0 {
		respondWithError(w, errors.New("from username/title/url at least one parameter is required"), http.StatusBadRequest)
		return
	}

	searchValues := make(map[string]string)

	if len(username) > 0 {
		searchValues["UserName"] = username
	}
	if len(title) > 0 {
		searchValues["Title"] = title
	}
	if len(url) > 0 {
		searchValues["URL"] = url
	}

	entry, err := findEntry(searchValues)

	if err != nil {
		respondWithError(w, err, http.StatusNotFound)
		return
	}

	json, err := marshalEntry(entry)

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func ReloadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := loadDB(r.FormValue("password"))

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	response := successResponse{Status: "success"}
	json, err := json.Marshal(&response)

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
