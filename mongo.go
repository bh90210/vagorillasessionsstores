// Package vagorillasessionsstores is a Gorilla sessions.Store implementation for MongoDB
package vagorillasessionsstores

import (
	"context"
	"encoding/base32"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoStore returns a new Mongo backed store.
//
// var cred options.Credential
//
// cred.AuthSource = "YourAuthSource"
// cred.Username = "YourUserName"
// cred.Password = "YourPassword"
//
// clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")).SetAuth(cred)
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
func NewMongoStore(opts *options.ClientOptions, keyPairs ...[]byte) (*MongoStore, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	collection := client.Database("sessions").Collection("store")

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

// NewMongoStoreWithOpts is intended for advanced configuration of Mongo's client.
// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
// func NewMongoStoreWithOpts(client *mongo.Client, keyPairs ...[]byte) (*Store, error) {
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	err := client.Connect(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	collection := client.Database("sessions").Collection("store")

// 	store := &Store{
// 		Codecs: securecookie.CodecsFromPairs(keyPairs...),
// 		Options: &sessions.Options{
// 			Path:   "/",
// 			MaxAge: 86400 * 30,
// 		},
// 		db: collection,
// 	}

// 	store.MaxAge(store.Options.MaxAge)
// 	return store, nil
// }

// MongoStore stores sessions using MongoDB
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

type SessionEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	SessionID string             `bson:"sessionid,omitempty"`
	Value     string             `bson:"value,omitempty"`
}

func (s *MongoStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}

	filt := SessionEntry{
		SessionID: session.ID,
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	opts := options.Update().SetUpsert(true)
	_, err = s.db.UpdateOne(
		ctx,
		filt,
		bson.D{
			{"$set", bson.D{{"value", encoded}, {"sessionid", session.ID}}},
		},
		opts,
	)

	return err
}

func (s *MongoStore) load(session *sessions.Session) error {
	var result SessionEntry

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	entry := SessionEntry{
		SessionID: session.ID,
	}
	res := s.db.FindOne(ctx, entry)
	err := res.Decode(&result)
	if err != nil {
		return (err)
	}

	return securecookie.DecodeMulti(session.Name(), string(result.Value), &session.Values, s.Codecs...)
}

func (s *MongoStore) erase(session *sessions.Session) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	entry := SessionEntry{
		SessionID: session.ID,
	}
	_, err := s.db.DeleteOne(ctx, entry)

	return err
}
