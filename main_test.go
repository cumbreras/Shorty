package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cumbreras/shortener/ent"
	"github.com/cumbreras/shortener/ent/enttest"
	"github.com/cumbreras/shortener/model"
	"github.com/cumbreras/shortener/repository"
	"github.com/cumbreras/shortener/service"
	"github.com/google/uuid"

	"github.com/cumbreras/shortener/server"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

// Helpers for this spec
type shortenURLFixture struct {
	URL        string
	Code       uuid.UUID
	StatusCode int
}

func performRequest(method, path string, requestPayload []byte, server *server.Server) (*httptest.ResponseRecorder, *http.Request) {
	fmt.Println(string(requestPayload))
	req := httptest.NewRequest(method, path, bytes.NewBuffer(requestPayload))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return w, req
}

func setupServer(t *testing.T) (*server.Server, *ent.Client) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "Shortener",
		Level: hclog.LevelFromString("DEBUG"),
	})
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	rp := repository.New(client, logger)
	svc := service.New(rp, logger)
	srv := server.New(mux.NewRouter(), logger, svc)
	return srv, client
}

func storeSeed(ctx context.Context, t *testing.T, fixture *shortenURLFixture, dbClient *ent.Client) {
	uid, err := uuid.Parse(fixture.Code.String())
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.ShortenURL.Create().SetCode(uid).SetURL(fixture.URL).Save(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestGetHandle(t *testing.T) {
	t.Run("When the resource exist", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusMovedPermanently}
		srv, dbClient := setupServer(t)
		storeSeed(context.Background(), t, fixture, dbClient)

		rec, _ := performRequest(http.MethodGet, "/"+fixture.Code.String(), nil, srv)

		res := rec.Result()
		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected status %d ; got %d", fixture.StatusCode, res.StatusCode)
		}
	})

	t.Run("When the resource does not exist", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusNotFound}
		srv, _ := setupServer(t)

		rec, _ := performRequest(http.MethodGet, "/"+fixture.Code.String(), nil, srv)

		res := rec.Result()
		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected status %d ; got %d", fixture.StatusCode, res.StatusCode)
		}
	})
}

func TestCreateHandler(t *testing.T) {
	t.Run("When the parameters are correct", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusCreated}
		requestPayload := []byte(`{"url": "https://news.ycombinator.com"}`)
		srv, _ := setupServer(t)
		rec, _ := performRequest(http.MethodPost, "/", requestPayload, srv)

		res := rec.Result()
		responseBody := model.New()
		if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
			t.Error(err)
		}

		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected status %d; got %d", fixture.StatusCode, res.StatusCode)
		}

		if responseBody.URL != fixture.URL {
			t.Errorf("expected %s got %s", fixture.URL, responseBody.URL)
		}
	})

	t.Run("When the parameters are missing", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusConflict}
		requestPayload := []byte(`{"url": ""}`)
		srv, _ := setupServer(t)
		rec, _ := performRequest(http.MethodPost, "/", requestPayload, srv)

		res := rec.Result()
		responseBody := model.New()
		if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
			t.Error(err)
		}

		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected status %d; got %d", fixture.StatusCode, res.StatusCode)
		}
	})

	t.Run("When the URL parameter is malformed", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusConflict}
		requestPayload := []byte(`{"url": ":\\//aa.ww.http"}`)
		srv, _ := setupServer(t)
		rec, _ := performRequest(http.MethodPost, "/", requestPayload, srv)

		res := rec.Result()
		responseBody := model.New()
		if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
			t.Error(err)
		}

		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected status %d; got %d", fixture.StatusCode, res.StatusCode)
		}
	})
}

func TestDeleteHandler(t *testing.T) {
	t.Run("When the resource still active", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusNoContent}
		srv, dbClient := setupServer(t)
		storeSeed(context.Background(), t, fixture, dbClient)
		rec, _ := performRequest(http.MethodDelete, "/"+fixture.Code.String(), nil, srv)

		res := rec.Result()
		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected %d; got %d", fixture.StatusCode, res.StatusCode)
		}
	})

	t.Run("When the resource is not active anymore", func(t *testing.T) {
		fixture := &shortenURLFixture{URL: "https://news.ycombinator.com", Code: uuid.New(), StatusCode: http.StatusNotFound}
		srv, _ := setupServer(t)
		rec, _ := performRequest(http.MethodDelete, "/"+fixture.Code.String(), nil, srv)

		res := rec.Result()
		if res.StatusCode != fixture.StatusCode {
			t.Errorf("expected %d; got %d", fixture.StatusCode, res.StatusCode)
		}
	})
}
