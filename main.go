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
	"context"
	_ "expvar"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/cyverse-de/go-mod/otelutils"
	"github.com/cyverse-de/messaging/v9"
	"github.com/cyverse-de/version"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/cyverse-de/configurate"

	"github.com/cyverse-de/jex-adapter/adapter"
	"github.com/cyverse-de/jex-adapter/db"
	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/cyverse-de/jex-adapter/millicores"
	"github.com/cyverse-de/jex-adapter/previewer"

	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	_ "github.com/lib/pq"
)

const serviceName = "jex-adapter"

var log = logging.Log.WithFields(logrus.Fields{"package": "main"})

func dbConnection(cfg *viper.Viper) *sqlx.DB {
	log := log.WithFields(logrus.Fields{"context": "database configuration"})

	dbURI := cfg.GetString("db.uri")
	if dbURI == "" {
		log.Fatal("db.uri must be set in the configuration file")
	}

	dbURL, err := url.Parse(dbURI)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("db host is %s:%s%s?%s", dbURL.Hostname(), dbURL.Port(), dbURL.Path, dbURL.RawQuery)

	dbconn := otelsqlx.MustConnect("postgres", dbURI, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))

	log.Info("connected to the database")

	return dbconn
}

func amqpConnection(cfg *viper.Viper) (*messaging.Client, string) {
	log := log.WithFields(logrus.Fields{"context": "amqp configuration"})

	amqpURI := cfg.GetString("amqp.uri")
	if amqpURI == "" {
		log.Fatal("amqp.uri must be set in the configuration file")
	}

	aURL, err := url.Parse(amqpURI)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("amqp broker host is %s:%s%s", aURL.Hostname(), aURL.Port(), aURL.Path)

	exchangeName := cfg.GetString("amqp.exchange.name")
	if exchangeName == "" {
		log.Fatal("amqp.exchange.name must be set in the configuration file")
	}
	log.Infof("amqp exchange name is %s", exchangeName)

	amqpclient, err := messaging.NewClient(amqpURI, false)
	if err != nil {
		log.Fatal(err)
	}

	if err = amqpclient.SetupPublishing(exchangeName); err != nil {
		log.Fatal(err)
	}

	log.Info("set up AMQP connection")

	return amqpclient, exchangeName
}

func main() {
	log := log.WithFields(logrus.Fields{"context": "main function"})

	var (
		showVersion       = flag.Bool("version", false, "Print version information")
		cfgPath           = flag.String("config", "", "Path to the configuration file")
		addr              = flag.String("addr", ":60000", "The port to listen on for HTTP requests")
		defaultMillicores = flag.Float64("default-millicores", 4000.0, "The default number of millicores reserved for an analysis.")
		logLevel          = flag.String("log-level", "info", "One of trace, debug, info, warn, error, fatal, or panic.")
	)

	flag.Parse()
	logging.SetupLogging(*logLevel)

	var tracerCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	shutdown := otelutils.TracerProviderFromEnv(tracerCtx, serviceName, func(e error) { log.Fatal(e) })
	defer shutdown()
	log.Infof("log level is %s", *logLevel)
	log.Infof("default millicores is %f", *defaultMillicores)

	if *showVersion {
		version.AppVersion()
		os.Exit(0)
	}

	if *cfgPath == "" {
		fmt.Println("--config is required")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	log.Infof("config path is %s", *cfgPath)

	cfg, err := configurate.InitDefaults(*cfgPath, configurate.JobServicesDefaults)
	if err != nil {
		log.Fatal(err)
	}

	dbconn := dbConnection(cfg)
	amqpclient, exchangeName := amqpConnection(cfg)

	dbase := db.New(dbconn)
	detector, err := millicores.New(dbase, *defaultMillicores)
	if err != nil {
		log.Fatal(err)
	}
	messenger := adapter.NewAMQPMessenger(exchangeName, amqpclient)

	p := previewer.New()
	a := adapter.New(cfg, detector, messenger)

	go a.Run()
	defer a.Finish()

	router := echo.New()
	router.Use(otelecho.Middleware(serviceName))
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

	log.Infof("starting server on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}
