package book

import (
	"encoding/json"
	"net/http"
)

func InitBookHandler() {
	handler := &BookHandler{}
	http.Handle("/books", handler)
}

type BookHandler struct{}

func (bh *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(bh.getAllBooks())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (bh *BookHandler) getAllBooks() []*Book {
	return []*Book{
		{
			ID:   1,
			Name: "Go",
		},
		{
			ID:   2,
			Name: "Python",
		},
	}
}
