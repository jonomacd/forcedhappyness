package dao

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Session is used to load and save session data in the datastore.
type Session struct {
	Date  time.Time
	Value []byte
}

type key struct {
	KeyPairs [][]byte
}

// NewDatastoreSessionStore returns a new DatastoreStore.
//
// The kind argument is the kind name used to store the session data.
// If empty it will use "Session".
//
// See NewCookieStore() for a description of the other parameters.
func NewSessionStore() *DatastoreStore {

	kind := "session"
	k := datastore.NameKey("sessionKey", "key", nil)
	kp := &key{}
	err := ds.Get(context.Background(), k, kp)
	if err == datastore.ErrNoSuchEntity {
		kp.KeyPairs = [][]byte{securecookie.GenerateRandomKey(64)}
		_, err = ds.Put(context.Background(), k, kp)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}

	return &DatastoreStore{
		Codecs: securecookie.CodecsFromPairs(kp.KeyPairs...),
		Options: &sessions.Options{
			Domain: ".iwillbenice.com",
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		kind: kind,
	}
}

// DatastoreStore stores sessions in the App Engine datastore.
type DatastoreStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
	kind    string
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *DatastoreStore) Get(r *http.Request, name string) (*sessions.Session,
	error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *DatastoreStore) New(r *http.Request, name string) (*sessions.Session,
	error) {
	session := sessions.NewSession(s, name)
	session.Options.MaxAge = 60 * 60 * 24 * 365 * 4 // 4 years in seconds
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(r, session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *DatastoreStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		session.ID = string(newToken("", 32))
	}
	if err := s.save(r, session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	options := s.Options
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, options))
	return nil
}

// save writes encoded session.Values to datastore.
func (s *DatastoreStore) save(r *http.Request,
	session *sessions.Session) error {
	if len(session.Values) == 0 {
		// Don't need to write anything.
		return nil
	}
	serialized, err := serialize(session.Values)
	if err != nil {
		return err
	}

	k := datastore.NameKey(s.kind, session.ID, nil)
	k, err = ds.Put(context.Background(), k, &Session{
		Date:  time.Now(),
		Value: serialized,
	})
	if err != nil {
		return fmt.Errorf("Could not put session %s: %v", session.ID, err)
	}
	return nil
}

// load gets a value from datastore and decodes its content into
// session.Values.
func (s *DatastoreStore) load(r *http.Request,
	session *sessions.Session) error {

	k := datastore.NameKey(s.kind, session.ID, nil)
	entity := Session{}
	if err := ds.Get(context.Background(), k, &entity); err != nil {
		return fmt.Errorf("Could not get session %s: %v", session.ID, err)
	}
	if err := deserialize(entity.Value, &session.Values); err != nil {
		return err
	}
	return nil
}

// Serialization --------------------------------------------------------------

// serialize encodes a value using gob.
func serialize(src interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserialize decodes a value using gob.
func deserialize(src []byte, dst interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(src))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func newToken(prefix string, length int) string {
	token := make([]byte, length)
	for ii := 0; ii < length; ii++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			// This should never happen
			panic(err)
		}
		token[ii] = chars[n.Uint64()]
	}

	return prefix + string(token)
}
