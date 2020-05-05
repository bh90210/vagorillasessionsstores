// Package badgerstore is a Gorilla sessions.Store implementation for BadgerDB
package badgerstore

import (
	"net/http"
	"os"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// BadgerStore stores sessions using BadgerDB
type BadgerStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options

	db *badger.DB
}

// NewBadgerStore returns a new BadgerStore.
//
// Filesystem directory where Badger is located. It will be created if it doesn't exist.
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
		path = os.TempDir()
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

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *BadgerStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// return GetRegistry(r).Get(s, name)
	return nil, nil
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
		err = securecookie.DecodeMulti(name, c.Value, &session.Values,
			s.Codecs...)
		if err == nil {
			session.IsNew = false
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *BadgerStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
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
