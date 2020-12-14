// Package mongo is a Gorilla sessions.Store implementation for BadgerDB
package mongo

import (
	"context"
	"encoding/base32"
	"fmt"
	"net/http"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoStore returns a new Mongo backed store.
//
// Path represents a filesystem directory where Badger is located. It will be created if it doesn't exist.
// Badger's DefaultOptions are used badger.DefaultOptions(path).
// For use with custom options see NewBadgerStoreWithOpts()
//
// Keys are defined in pairs to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes.
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
func NewMongoStore(uri string, keyPairs ...[]byte) (*MongoStore, error) {
	// if path == "" {
	// 	path = filepath.Join(os.TempDir(), "mongo")
	// }

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	collection := client.Database("testing").Collection("numbers")
	// db, err := badger.Open(badger.DefaultOptions(path))
	// if err != nil {
	// 	return nil, err
	// }

	store := &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		db: collection,
	}

	store.MaxAge(store.Options.MaxAge)
	return store, nil
}

// NewMongoStoreWithOpts is intended for advanced configuration of Badger.
// Create a new variable `opts := badger.Options{}` and set on it the desired settings.
// For more information please see Badger's documentation https://github.com/dgraph-io/badger
func NewMongoStoreWithOpts(opts badger.Options, keyPairs ...[]byte) (*MongoStore, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	store := &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		db: db,
	}

	store.MaxAge(store.Options.MaxAge)
	return store, nil
}

// BadgerStore stores sessions using BadgerDB
type MongoStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	db      *mongo.Collection
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *MongoStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (s *MongoStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID,
			s.Codecs...)
		err = s.load(session)
		if err == nil {
			session.IsNew = false
		}
	}
	return session, err
}

// Save adds a single session to the response.
//
// If the Options.MaxAge of the session is <= 0 then the session file will be
// deleted from the store path. With this process it enforces the properly
// session cookie handling so no need to trust in the cookie management in the
// web browser.
func (s *MongoStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	// Delete if max-age is <= 0
	if s.Options.MaxAge <= 0 {
		if err := s.erase(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting Options.MaxAge
// = -1 for that session.
func (s *MongoStore) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// Edit is a helper function for editing sessions directly from the back-end store without http request from the user.
func (s *MongoStore) Edit(session *sessions.Session) {
	if err := s.save(session); err != nil {
		fmt.Println(err)
	}
}

// Delete is a helper function for deleting sessions directly from the back-end store without http request from the user.
func (s *MongoStore) Delete(session *sessions.Session) {
	if err := s.erase(session); err != nil {
		fmt.Println(err)
	}
}

func (s *MongoStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}
	// return s.db.Update(func(txn *badger.Txn) error {
	// 	err := txn.Set([]byte("session_"+session.ID), []byte(encoded))
	// 	return err
	// })
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// res, err := s.db.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	_, err = s.db.UpdateOne(ctx, bson.M{session.ID: []byte(encoded)}, "")
	return err
}

func (s *MongoStore) load(session *sessions.Session) error {
	var queryResp []byte

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("session_" + session.ID))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			queryResp = append(queryResp, val...)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return securecookie.DecodeMulti(session.Name(), string(queryResp), &session.Values, s.Codecs...)
}

func (s *MongoStore) erase(session *sessions.Session) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte("session_" + session.ID))
		return err
	})
	if err != nil {
		return err
	}

	return err
}
