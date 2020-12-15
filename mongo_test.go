package vagorillasessionsstores

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// // Test for MongoStore
// func TestMongoStore(t *testing.T) {
// 	originalPath := "/"
// 	store, err := NewMongoStore("")
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}

// 	store.Options.Path = originalPath
// 	req, err := http.NewRequest("GET", "http://www.example.com", nil)
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}

// 	session, err := store.New(req, "hello")
// 	if err != nil {
// 		t.Fatal("failed to create session", err)
// 	}

// 	store.Options.Path = "/foo"
// 	if session.Options.Path != originalPath {
// 		t.Fatalf("bad session path: got %q, want %q", session.Options.Path, originalPath)
// 	}
// }

// // Test delete Mongo store with max-age: -1
// func TestMongoStoreDelete(t *testing.T) {
// 	store, err := NewMongoStore("", []byte("some key"))
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}

// 	req, err := http.NewRequest("GET", "http://www.example.com", nil)
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}
// 	w := httptest.NewRecorder()

// 	session, err := store.New(req, "hello")
// 	if err != nil {
// 		t.Fatal("failed to create session", err)
// 	}

// 	err = session.Save(req, w)
// 	if err != nil {
// 		t.Fatal("failed to save session", err)
// 	}

// 	session.Options.MaxAge = -1
// 	err = session.Save(req, w)
// 	if err != nil {
// 		t.Fatal("failed to delete session", err)
// 	}
// }

// // Test delete Mongo store with max-age: 0
// func TestMongoStoreDelete2(t *testing.T) {
// 	store, err := NewMongoStore("", []byte("some key"))
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}

// 	req, err := http.NewRequest("GET", "http://www.example.com", nil)
// 	if err != nil {
// 		t.Fatal("failed to create request", err)
// 	}
// 	w := httptest.NewRecorder()

// 	session, err := store.New(req, "hello")
// 	if err != nil {
// 		t.Fatal("failed to create session", err)
// 	}

// 	err = session.Save(req, w)
// 	if err != nil {
// 		t.Fatal("failed to save session", err)
// 	}

// 	session.Options.MaxAge = 0
// 	err = session.Save(req, w)
// 	if err != nil {
// 		t.Fatal("failed to delete session", err)
// 	}
// }
