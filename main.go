package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type Mapper interface {
	Range(func(key, value interface{}) (more bool))
	Load(key interface{}) (value interface{}, loaded bool)
	Store(key, value interface{})
	Delete(key interface{})
}

type Cache struct {
	Mapper
}

type Content struct {
	Type []string
	bytes.Buffer
}

func (c *Content) ReadFrom(r io.ReadCloser) (n int64, err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	return c.Buffer.ReadFrom(r)
}

func (c *Cache) Keys(w http.ResponseWriter, r *http.Request) {
	c.Mapper.Range(func(key, value interface{}) bool {
		_, err := fmt.Fprintln(w, key)
		return err == nil
	})
}

func (c *Cache) Get(w http.ResponseWriter, r *http.Request) {
	b, ok := c.Mapper.Load(r.URL.Path[1:])
	if ok {
		switch b := b.(type) {
		case Content:
			for _, h := range b.Type {
				w.Header().Add("Content-Type", h)
			}
			_, _ = b.WriteTo(w)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (c *Cache) Put(w http.ResponseWriter, r *http.Request) {
	b := Content{Type: r.Header["Content-Type"]}
	_, err := b.ReadFrom(r.Body)
	if err == nil {
		c.Mapper.Store(r.URL.Path[1:], b)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (c *Cache) Delete(w http.ResponseWriter, r *http.Request) {
	c.Mapper.Delete(r.URL.Path[1:])
}

func main() {
	a := Cache{Mapper: &sync.Map{}}
	h := mux.NewRouter()
	h.HandleFunc("/", a.Keys).Methods(http.MethodGet)
	h.HandleFunc("/{key}", a.Get).Methods(http.MethodGet)
	h.HandleFunc("/{key}", a.Put).Methods(http.MethodPut)
	h.HandleFunc("/{key}", a.Delete).Methods(http.MethodDelete)
	h.Use(logMiddleware)
	h.Use(jwtMiddleware)
	err := http.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}

type jwtHandle struct {
	h http.Handler
	k [32]byte
}

func (j *jwtHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete:
		if a, ok := r.Header["Authorization"]; ok && len(a) > 0 && strings.HasPrefix(a[0], "Bearer") {
			if t, err := jwt.Parse(strings.TrimSpace(a[0][6:]), func(t *jwt.Token) (interface{}, error) {
				switch t.Method.(type) {
				case *jwt.SigningMethodHMAC:
					return j.k[:], nil
				}
				return nil, jwt.ErrSignatureInvalid
			}); err == nil && t.Valid {
				break
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	default:
		if v, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
		}).SignedString(j.k[:]); err == nil {
			w.Header().Add("Authorization", "Bearer "+v)
			break
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j.h.ServeHTTP(w, r)
}

func jwtMiddleware(h http.Handler) http.Handler {
	return &jwtHandle{h: h, k: [32]byte{'T', 'O', 'P', 'S', 'E', 'C', 'R', 'E', 'T'}}
}

type logHandle struct {
	h http.Handler
	w http.ResponseWriter
	c int
	n int
	u int
	b io.ReadCloser
}

func (l *logHandle) Read(p []byte) (n int, err error) {
	n, err = l.b.Read(p)
	l.u += n
	return
}

func (l *logHandle) Close() error {
	return l.b.Close()
}

func (l *logHandle) Header() http.Header {
	return l.w.Header()
}

func (l *logHandle) Write(b []byte) (n int, err error) {
	n, err = l.w.Write(b)
	l.n += n
	return
}

func (l *logHandle) WriteHeader(c int) {
	l.c = c
	l.w.WriteHeader(c)
}

func (l *logHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	l.w = w
	l.b = r.Body
	r.Body = l
	l.h.ServeHTTP(l, r)
	log.Println(r.Method, r.URL, r.Proto, l.c, l.u, l.n, time.Now().Sub(t))
}

func logMiddleware(h http.Handler) http.Handler {
	return &logHandle{h: h, c: http.StatusOK}
}
