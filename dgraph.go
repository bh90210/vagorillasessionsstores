// Package vagorillasessionsstores is a Gorilla sessions.Store implementation for BadgerDB, MongoDB and Dgraph
package vagorillasessionsstores

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"google.golang.org/grpc"
)

// NewDgraphStore returns a new Dgraph backed store.
//
// A gRPC connection is needed before the store initiates.
// The store implements a Close() function to on SIGTERM.
// TODO: finish documentation
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
func NewDgraphStore(conn *grpc.ClientConn, keyPairs ...[]byte) (*DgraphStore, error) {
	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	store := &DgraphStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		db: dg,
	}

	store.MaxAge(store.Options.MaxAge)
	return store, nil
}

// NewDgraphStoreWithSchema returns a new Dgraph backed store but also initiates it with store's schema.
// 	sessionid: string @index(hash) .
// 	sessionvalue: string .
// 	type Session {
// 		sessionid
// 		sessionvalue
// 	}
//
// A gRPC connection is needed before the store initiates.
// The store implements a Close() function to on SIGTERM.
// TODO: finish documentation
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
func NewDgraphStoreWithSchema(conn *grpc.ClientConn, keyPairs ...[]byte) (*DgraphStore, error) {
	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	op := &api.Operation{}
	op.Schema = `
	sessionid: string @index(hash) . 
	sessionvalue: string . 
	type Session {
		sessionid
		sessionvalue
	}
	`

	ctx := context.Background()
	err := dg.Alter(ctx, op)
	if err != nil {
		log.Fatal(err)
	}

	store := &DgraphStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		db: dg,
	}

	store.MaxAge(store.Options.MaxAge)
	return store, nil
}

// DgraphStore stores sessions using MongoDB
type DgraphStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	db      *dgo.Dgraph
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *DgraphStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (s *DgraphStore) New(r *http.Request, name string) (*sessions.Session, error) {
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
func (s *DgraphStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
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
func (s *DgraphStore) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// Session represents a custom Type in Dgraph
type Session struct {
	Uid          string   `json:"uid,omitempty"`
	DType        []string `json:"dgraph.type,omitempty"`
	SessionID    string   `json:"sessionid,omitempty"`
	SessionValue string   `json:"sessionvalue,omitempty"`
}

func (s *DgraphStore) save(session *sessions.Session) error {
	ctx := context.Background()

	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}

	query := `{
		  q(func: eq(sessionid, "` + session.ID + `")) {
			v as uid
		  }
}`

	mutation := `
	uid(v) <sessionid> "` + session.ID + `" .
	uid(v) <sessionvalue> "` + encoded + `" .
	uid(v) <dgraph.type> "Session" .`

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{
			{
				SetNquads: []byte(mutation),
			},
		},
		CommitNow: true,
	}

	_, err = s.db.NewTxn().Do(ctx, req)

	return err
}

func (s *DgraphStore) load(session *sessions.Session) error {
	ctx := context.Background()

	query := `{
	q(func: eq(sessionid, "` + session.ID + `")) {
	  sessionvalue
	}
}`

	request := &api.Request{
		Query:     query,
		CommitNow: true,
	}

	response, err := s.db.NewTxn().Do(ctx, request)
	if err != nil {
		return err
	}

	var r struct {
		Q []Session `json:"q"`
	}

	err = json.Unmarshal(response.Json, &r)
	if err != nil {
		return err
	}

	if len(r.Q[0].SessionValue) == 0 {
		return errors.New("no key found")
	}

	return securecookie.DecodeMulti(session.Name(), r.Q[0].SessionValue, &session.Values, s.Codecs...)
}

func (s *DgraphStore) erase(session *sessions.Session) error {
	ctx := context.Background()

	query := `{
		  q(func: eq(sessionid, "` + session.ID + `")) {
			v as uid
		  }
}`

	deletion := `uid(v) * * . `

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{
			{
				DelNquads: []byte(deletion),
			},
		},
		CommitNow: true,
	}

	_, err := s.db.NewTxn().Do(ctx, req)

	return err
}
