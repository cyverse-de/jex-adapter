package adapter

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cyverse-de/jex-adapter/db"
	"github.com/cyverse-de/jex-adapter/millicores"
	"github.com/cyverse-de/model/v6"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

func getTestConfig() *viper.Viper {
	cfg := viper.New()
	cfg.Set("condor.log_path", "/tmp/")
	cfg.Set("condor.filter_files", "output-last-stderr")
	cfg.Set("irods.base", "/iplant/home")
	return cfg
}

type TestMessenger struct{}

func (t *TestMessenger) Stop(context context.Context, id string) error {
	return nil
}

func (t *TestMessenger) Launch(context context.Context, job *model.Job) error {
	return nil
}

var a *JEXAdapter
var mock sqlmock.Sqlmock

func initTestAdapter(t *testing.T) (*JEXAdapter, sqlmock.Sqlmock) {
	if a == nil || mock == nil {
		var err error
		var mockconn *sql.DB
		cfg := getTestConfig()
		msger := &TestMessenger{}
		mockconn, mock, err = sqlmock.New()
		if err != nil {
			t.Fatalf("error opening mocked database connection: %s", err)
		}
		dbconn := sqlx.NewDb(mockconn, "postgres")
		dbase := db.New(dbconn)
		detector, err := millicores.New(dbase, 4000.0)
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		a = New(cfg, detector, msger)
	}
	return a, mock
}
func TestHomeHandler(t *testing.T) {
	a, _ := initTestAdapter(t)
	go a.Run()
	defer a.Finish()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	if assert.NoError(t, a.HomeHandler(c)) {
		assert.Equal(t, rec.Body.String(), "Welcome to the JEX.\n")
	}
}

func TestStopAdapter(t *testing.T) {
	a, _ := initTestAdapter(t)
	go a.Run()
	defer a.Finish()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/stop/:invocation_id")
	c.SetParamNames("invocation_id")
	c.SetParamValues("c654e8bb-d535-4f7a-bd0f-aff0f0c189b1")

	if assert.NoError(t, a.StopHandler(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
	}
}

func TestLaunchHandler(t *testing.T) {
	a, _ := initTestAdapter(t)
	go a.Run()
	defer a.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCondorLaunchJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	if assert.NoError(t, a.LaunchHandler(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
	}
}

func TestCondorLaunchDefaultMillicores(t *testing.T) {
	a, _ := initTestAdapter(t)
	go a.Run()
	defer a.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCondorCustomLaunchJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	if assert.NoError(t, a.LaunchHandler(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
	}
}
