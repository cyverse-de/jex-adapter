package previewer

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cyverse-de/jex-adapter/logging"
	"github.com/cyverse-de/jex-adapter/types"
	"github.com/cyverse-de/model/v6"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var log = logging.Log.WithFields(logrus.Fields{"package": "previewer"})

type Previewer struct{}

func New() *Previewer {
	return &Previewer{}
}

func (p *Previewer) Routes(router types.Router) {
	router.POST("", p.PreviewHandler)
	router.POST("/", p.PreviewHandler)
}

func (p *Previewer) PreviewHandler(c echo.Context) error {
	body := c.Request().Body

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Error(err)
		return err
	}

	var params model.PreviewableStepParam
	err = json.Unmarshal(bodyBytes, &params)
	if err != nil {
		log.Error(err)
		return err
	}

	output := map[string]string{
		"params": params.String(),
	}

	return c.JSON(http.StatusOK, output)
}
