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

	"github.com/cyverse-de/configurate"

	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/gorilla/mux"
)

var log = logging.Log.WithFields(logrus.Fields{"package": "main"})

var exchangeName string

// NewRouter returns a newly configured *mux.Router.
func (j *JEXAdapter) NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", j.home).Methods("GET")
	router.HandleFunc("/", j.launch).Methods("POST")
	router.HandleFunc("/stop/{invocation_id}", j.stop).Methods("DELETE")
	router.HandleFunc("/arg-preview", j.preview).Methods("POST")
	router.Handle("/debug/vars", http.DefaultServeMux)
	return router
}

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

	app := New(cfg)

	router := mux.NewRouter()
	router.HandleFunc("/", j.home).Methods("GET")
	router.HandleFunc("/", j.launch).Methods("POST")
	router.HandleFunc("/stop/{invocation_id}", j.stop).Methods("DELETE")
	router.HandleFunc("/arg-preview", j.preview).Methods("POST")
	router.Handle("/debug/vars", http.DefaultServeMux)

	log.Fatal(http.ListenAndServe(*addr, router))
}
