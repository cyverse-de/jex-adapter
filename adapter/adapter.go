package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cockroachdb/apd"
	"github.com/cyverse-de/go-mod/gotelnats"
	"github.com/cyverse-de/go-mod/pbinit"
	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/cyverse-de/jex-adapter/millicores"
	"github.com/cyverse-de/jex-adapter/types"
	"github.com/cyverse-de/messaging/v9"
	"github.com/cyverse-de/model/v6"
	"github.com/cyverse-de/p/go/qms"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
)

var log = logging.Log.WithFields(logrus.Fields{"package": "adapter"})

const otelName = "github.com/cyverse-de/jex-adapter/adapter"

type Messenger interface {
	Launch(context context.Context, job *model.Job) error
	Stop(context context.Context, id string) error
}

func amqpError(err error) {
	amqpErr, ok := err.(*amqp.Error)
	if !ok || amqpErr.Code != 0 {
		log.Fatal(err)
	}
}

type AMQPMessenger struct {
	exchange string
	client   *messaging.Client
	natsConn *nats.EncodedConn
}

func NewAMQPMessenger(exchange string, client *messaging.Client, natsConn *nats.EncodedConn) *AMQPMessenger {
	return &AMQPMessenger{
		exchange: exchange,
		client:   client,
		natsConn: natsConn,
	}
}

func (a *AMQPMessenger) Stop(context context.Context, id string) error {
	err := a.client.SendStopRequestContext(context, id, "root", "because I said to")
	if err != nil {
		amqpError(err)
	}
	return err
}

func (a *AMQPMessenger) getResourceOveragesForUser(ctx context.Context, username string) (*qms.OverageList, error) {
	var err error

	subject := "cyverse.qms.user.overages.get"

	req := &qms.AllUserOveragesRequest{
		Username: username,
	}

	_, span := pbinit.InitAllUserOveragesRequest(req, subject)
	defer span.End()

	resp := pbinit.NewOverageList()

	if err = gotelnats.Request(ctx, a.natsConn, subject, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *AMQPMessenger) validateLaunch(ctx context.Context, job *model.Job) error {
	overages, err := a.getResourceOveragesForUser(ctx, job.Submitter)
	if err != nil {
		return err
	}

	if overages != nil && len(overages.Overages) != 0 {
		var inOverage bool

		for _, ov := range overages.Overages {
			if ov.Usage >= ov.Quota && ov.ResourceName == "cpu.hours" {
				inOverage = true
			}
		}

		if inOverage {
			return fmt.Errorf("%s has resource overages", job.Submitter)
		}
	}

	return nil
}

func (a *AMQPMessenger) Launch(context context.Context, job *model.Job) error {
	var (
		err        error
		launchJSON []byte
	)

	ctx, span := otel.Tracer(otelName).Start(context, "Launch")
	defer span.End()

	if err = a.validateLaunch(ctx, job); err != nil {
		amqpError(err)
		return err
	}

	// This is included just for the side effect of creating the stop
	// queue before the job launches, otherwise some stop requests can be missed.
	s, err := a.client.CreateQueue(
		messaging.StopQueueName(job.InvocationID),
		a.exchange,
		messaging.StopRequestKey(job.InvocationID),
		false,
		true,
	)
	if err != nil {
		amqpError(err)
		s.Close()
		return err
	}
	defer s.Close()

	if launchJSON, err = json.Marshal(messaging.NewLaunchRequest(job)); err != nil {
		return err
	}

	if err = a.client.PublishContext(ctx, messaging.LaunchesKey, launchJSON); err != nil {
		amqpError(err)
		return err
	}

	return nil
}

type millicoresJob struct {
	ID                 uuid.UUID
	Job                model.Job
	MillicoresReserved *apd.Decimal
}

// JEXAdapter contains the application state for jex-adapter.
type JEXAdapter struct {
	cfg       *viper.Viper
	detector  *millicores.Detector
	messenger Messenger
	jobs      map[string]bool
	addJob    chan millicoresJob
	jobDone   chan uuid.UUID
	exit      chan bool
}

// New returns a *JEXAdapter
func New(cfg *viper.Viper, detector *millicores.Detector, messenger Messenger) *JEXAdapter {
	return &JEXAdapter{
		cfg:       cfg,
		messenger: messenger,
		detector:  detector,
		addJob:    make(chan millicoresJob),
		jobDone:   make(chan uuid.UUID),
		exit:      make(chan bool),
		jobs:      map[string]bool{},
	}
}

func (j *JEXAdapter) Run() {
	for {
		select {
		case mj := <-j.addJob:
			j.jobs[mj.ID.String()] = true
			go func(mj millicoresJob) {
				ctx, span := otel.Tracer(otelName).Start(context.Background(), "millicores iteration")
				defer span.End()

				var err error

				log.Infof("storing %s millicores reserved for %s", mj.MillicoresReserved.String(), mj.Job.InvocationID)
				if err = j.detector.StoreMillicoresReserved(ctx, &mj.Job, mj.MillicoresReserved); err != nil {
					log.Error(err)
				}
				log.Infof("done storing %s millicores reserved for %s", mj.MillicoresReserved.String(), mj.Job.InvocationID)

				j.jobDone <- mj.ID
			}(mj)

		case doneJobID := <-j.jobDone:
			delete(j.jobs, doneJobID.String())

		case <-j.exit:
			break
		}
	}
}

func (j *JEXAdapter) StoreMillicoresReserved(job model.Job, millicoresReserved *apd.Decimal) error {
	newjob := millicoresJob{
		ID:                 uuid.New(),
		Job:                job,
		MillicoresReserved: millicoresReserved,
	}

	j.addJob <- newjob

	return nil
}

func (j *JEXAdapter) Finish() {
	j.exit <- true
}

func (j *JEXAdapter) Routes(router types.Router) types.Router {
	log := log.WithFields(logrus.Fields{"context": "adding routes"})

	router.GET("", j.HomeHandler)
	router.GET("/", j.HomeHandler)
	log.Info("added handler for GET /")

	router.POST("", j.LaunchHandler)
	router.POST("/", j.LaunchHandler)
	log.Info("added handler for POST /")

	router.DELETE("/stop/:invocation_id", j.StopHandler)
	log.Info("added handler for DELETE /stop/:invocation_id")

	return router
}

func (j *JEXAdapter) HomeHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to the JEX.\n")
}

func (j *JEXAdapter) StopHandler(c echo.Context) error {
	var err error

	context := c.Request().Context()

	log := log.WithFields(logrus.Fields{"context": "stop app"})

	invID := c.Param("invocation_id")
	if invID == "" {
		err = errors.New("missing job id in URL")
		log.Error(err)
		return err
	}

	log = log.WithFields(logrus.Fields{"external_id": invID})

	log.Debug("starting sending stop message")
	err = j.messenger.Stop(context, invID)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("Done sending stop message")

	log.Info("sent stop message")

	return c.NoContent(http.StatusOK)
}

func (j *JEXAdapter) LaunchHandler(c echo.Context) error {
	request := c.Request()
	context := request.Context()

	log := log.WithFields(logrus.Fields{"context": "app launch"})

	log.Debug("reading request body")
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done reading request body")

	log.Debug("parsing request body JSON")
	job, err := model.NewFromData(j.cfg, bodyBytes)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done parsing request body JSON")

	log = log.WithFields(logrus.Fields{"external_id": job.InvocationID})

	log.Debug("sending launch message")
	if err = j.messenger.Launch(context, job); err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done sending launch message")

	log.Debug("finding number of millicores reserved")
	millicoresReserved, err := j.detector.NumberReserved(job)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done finding number of millicores reserved")

	log.Debug("before asynchronous StoreMillicoresReserved call")
	if err = j.StoreMillicoresReserved(*job, millicoresReserved); err != nil {
		log.Error(err)
		return err
	}
	log.Debug("after asynchronous StoreMillicoresReserved call")

	log.Infof("launched with %f millicores reserved", millicoresReserved)

	return c.NoContent(http.StatusOK)
}
