// jex-adapter
//
// jex-adapter allows the apps service to submit job requests through an AMQP
// broker by implementing the portion of the old JEX API that apps interacted
// with. Instead of writing the files out to disk and calling condor_submit like
// the JEX service did, it serializes the request as JSON and pushes it out
// as a message on the "jobs" exchange with a routing key of "jobs.launches".
//
package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/cyverse-de/version"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/cyverse-de/messaging.v6"

	"github.com/cyverse-de/configurate"

	"github.com/cyverse-de/jex-adapter/adapter"
	"github.com/cyverse-de/jex-adapter/db"
	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/cyverse-de/jex-adapter/millicores"
	"github.com/cyverse-de/jex-adapter/previewer"
)

var log = logging.Log.WithFields(logrus.Fields{"package": "main"})

func main() {
	var (
		showVersion       = flag.Bool("version", false, "Print version information")
		cfgPath           = flag.String("config", "", "Path to the configuration file")
		addr              = flag.String("addr", ":60000", "The port to listen on for HTTP requests")
		defaultMillicores = flag.Float64("default-millicores", 4000.0, "The default number of millicores reserved for an analysis.")
		logLevel          = flag.String("log-level", "info", "One of trace, debug, info, warn, error, fatal, or panic.")
	)

	flag.Parse()
	logging.SetupLogging(*logLevel)

	if *showVersion {
		version.AppVersion()
		os.Exit(0)
	}

	if *cfgPath == "" {
		fmt.Println("--config is required")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	cfg, err := configurate.InitDefaults(*cfgPath, configurate.JobServicesDefaults)
	if err != nil {
		log.Fatal(err)
	}

	amqpURI := cfg.GetString("amqp.uri")
	if amqpURI == "" {
		log.Fatal("amqp.uri must be set in the configuration file")
	}

	exchangeName := cfg.GetString("amqp.exchange.name")
	if exchangeName == "" {
		log.Fatal("amqp.exchange.name must be set in the configuration file")
	}

	dbURI := cfg.GetString("db.uri")
	if dbURI == "" {
		log.Fatal("db.uri must be set in the configuration file")
	}

	dbconn := sqlx.MustConnect("postgres", dbURI)
	dbase := db.New(dbconn)
	detector := millicores.New(dbase, *defaultMillicores)

	amqpclient, err := messaging.NewClient(amqpURI, false)
	if err != nil {
		log.Fatal(err)
	}

	messenger := adapter.NewAMQPMessenger(exchangeName, amqpclient)

	p := previewer.New()
	a := adapter.New(cfg, detector, messenger)

	router := echo.New()
	router.HTTPErrorHandler = logging.HTTPErrorHandler

	routerLogger := log.Writer()
	defer routerLogger.Close()

	router.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: "${method} ${uri} ${status}",
			Output: routerLogger,
		},
	))

	a.Routes(router)

	previewrouter := router.Group("/arg-preview")
	p.Routes(previewrouter)

	log.Fatal(http.ListenAndServe(*addr, router))
}
