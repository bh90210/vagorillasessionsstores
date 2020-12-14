package dgraph

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test for BadgerStore
func TestBadgerStore(t *testing.T) {
	originalPath := "/"
	store, err := NewBadgerStore("~/test1")
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	store.Options.Path = originalPath
	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	store.Options.Path = "/foo"
	if session.Options.Path != originalPath {
		t.Fatalf("bad session path: got %q, want %q", session.Options.Path, originalPath)
	}
}

// Test delete badger store with max-age: -1
func TestBadgerStoreDelete(t *testing.T) {
	store, err := NewBadgerStore("~/test2", []byte("some key"))
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save session", err)
	}

	session.Options.MaxAge = -1
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to delete session", err)
	}
}

// Test delete badger store with max-age: 0
func TestBadgerStoreDelete2(t *testing.T) {
	store, err := NewBadgerStore("~/test3", []byte("some key"))
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save session", err)
	}

	session.Options.MaxAge = 0
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to delete session", err)
	}
}
