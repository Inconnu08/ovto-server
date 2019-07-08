package handlers
//
//import (
//	"authdemo/internal/service"
//	"bytes"
//	"database/sql"
//	"fmt"
//	"github.com/hako/branca"
//	"io/ioutil"
//	"net/http"
//	"net/http/httptest"
//	"net/url"
//	"strconv"
//	"testing"
//)
//
//type testLoginInput struct {
//	Email    string
//	Password string
//}
//
//func SetupTestCase(t *testing.T) {
//	var (
//		port      = "3000"
//		originStr = "http://localhost:" + port
//		dbURL     = "postgresql://root@127.0.0.1:26257/authdemo?sslmode=disable"
//		tokenKey  = "supersecretkeyyoushouldnotcommit"
//	)
//
//	origin, err := url.Parse(originStr)
//	if err != nil || !origin.IsAbs() {
//		t.Fatalf("invalid origin url: %v\n", err)
//		return
//	}
//
//	db, err := sql.Open("pgx", dbURL)
//	if err != nil {
//		t.Fatalf("cound not open db connection %v\n", err)
//		return
//	}
//
//	defer db.Close()
//
//	if err = db.Ping(); err != nil {
//		t.Fatalf("could not ping to db %v\n", err)
//		return
//	}
//
//	codec := branca.NewBranca(tokenKey)
//	codec.SetTTL(uint32(service.TokenLifeSpan.Seconds()))
//
//	s := service.New(db, codec, *origin)
//	h := New(s)
//	addr := fmt.Sprintf(":%s", port)
//	if err := http.ListenAndServe(addr, h); err != nil {
//		t.Fatalf("could not start server: %v\n", err)
//	}
//}
//
//func TestLoginHandler(t *testing.T) {
//	tt := []struct {
//		name       string
//		value      testLoginInput
//		expected   string
//		statusCode int
//	}{
//		{name: "Valid credentials", value: testLoginInput{Email: "jonsnow@got.com", Password: "gameofthrones!"}, expected: "fullname", statusCode: 200},
//	}
//
//	for _, tc := range tt {
//		t.Run(tc.name, func(t *testing.T) {
//			req, err := http.NewRequest("GET", "localhost:3000/api/login?v="+tc.value, nil)
//			if err != nil {
//				t.Fatalf("could not create request: %v", err)
//			}
//			rec := httptest.NewRecorder()
//			handler := SetupTestCase(t)
//			handler.login(rec, req)
//
//			res := rec.Result()
//			defer res.Body.Close()
//
//			b, err := ioutil.ReadAll(res.Body)
//			if err != nil {
//				t.Fatalf("could not read response: %v", err)
//			}
//
//			if tc.err != "" {
//				// do something
//				if res.StatusCode != http.StatusBadRequest {
//					t.Errorf("expected status Bad Request; got %v", res.StatusCode)
//				}
//				if msg := string(bytes.TrimSpace(b)); msg != tc.err {
//					t.Errorf("expected message %q; got %q", tc.err, msg)
//				}
//				return
//			}
//
//			if res.StatusCode != http.StatusOK {
//				t.Errorf("expected status OK; got %v", res.Status)
//			}
//
//			d, err := strconv.Atoi(string(bytes.TrimSpace(b)))
//			if err != nil {
//				t.Fatalf("expected an integer; got %s", b)
//			}
//			if d != tc.double {
//				t.Fatalf("expected double to be %v; got %v", tc.double, d)
//			}
//		})
//	}
//}
//
//func TestRouting(t *testing.T) {
//	srv := httptest.NewServer(handler())
//	defer srv.Close()
//
//	res, err := http.Get(fmt.Sprintf("%s/double?v=2", srv.URL))
//	if err != nil {
//		t.Fatalf("could not send GET request: %v", err)
//	}
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK {
//		t.Errorf("expected status OK; got %v", res.Status)
//	}
//
//	b, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		t.Fatalf("could not read response: %v", err)
//	}
//
//	d, err := strconv.Atoi(string(bytes.TrimSpace(b)))
//	if err != nil {
//		t.Fatalf("expected an integer; got %s", b)
//	}
//	if d != 4 {
//		t.Fatalf("expected double to be 4; got %v", d)
//	}
//}
