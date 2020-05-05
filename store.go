// Package badgerstore is a Gorilla sessions.Store implementation for BadgerDB
package badgerstore

import (
	"encoding/base32"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// NewBadgerStore returns a new BadgerStore.
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
func NewBadgerStore(path string, keyPairs ...[]byte) (*BadgerStore, error) {
	if path == "" {
		path = filepath.Join(os.TempDir(), "badger")
	}

	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	store := &BadgerStore{
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

// NewBadgerStoreWithOpts is intended for advanced configuration of Badger.
// Create a new variable `opts := badger.Options{}` and set on it the desired settings.
// For more information please see Badger's documentation https://github.com/dgraph-io/badger
func NewBadgerStoreWithOpts(opts badger.Options, keyPairs ...[]byte) (*BadgerStore, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	store := &BadgerStore{
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
type BadgerStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	db      *badger.DB
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *BadgerStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (s *BadgerStore) New(r *http.Request, name string) (*sessions.Session, error) {
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
func (s *BadgerStore) Save(r *http.Request, w http.ResponseWriter,
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
func (s *BadgerStore) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *BadgerStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("session_"+session.ID), []byte(encoded))
		return err
	})
}

func (s *BadgerStore) load(session *sessions.Session) error {
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

func (s *BadgerStore) erase(session *sessions.Session) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte("session_" + session.ID))
		return err
	})
	if err != nil {
		return err
	}

	return err
}
