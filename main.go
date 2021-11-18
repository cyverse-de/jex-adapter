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
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/cyverse-de/configurate"

	"github.com/cyverse-de/jex-adapter/adapter"
	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/cyverse-de/jex-adapter/previewer"
)

var log = logging.Log.WithFields(logrus.Fields{"package": "main"})

func main() {
	var (
		showVersion = flag.Bool("version", false, "Print version information")
		cfgPath     = flag.String("config", "", "Path to the configuration file")
		addr        = flag.String("addr", ":60000", "The port to listen on for HTTP requests")
	)

	flag.Parse()

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

	//dbconn := sqlx.MustConnect("postgres", dbURI)

	p := previewer.New()
	a := adapter.New(cfg, amqpURI, exchangeName)

	router := echo.New()
	router.HTTPErrorHandler = logging.HTTPErrorHandler

	a.Routes(router)

	previewrouter := router.Group("/arg-preview")
	p.Routes(previewrouter)

	log.Fatal(http.ListenAndServe(*addr, router))
}
