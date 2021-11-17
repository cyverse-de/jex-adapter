package previewer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cyverse-de/model"
	"github.com/labstack/echo/v4"
)

type Previewer struct {
	router *echo.Echo
}

func New(router *echo.Echo) *Previewer {
	return &Previewer{
		router: router,
	}
}

func (p *Previewer) Router() *echo.Echo {
	p.router.POST("", p.PreviewHandler)
	p.router.POST("/", p.PreviewHandler)

	return p.router
}

func (p *Previewer) PreviewHandler(c echo.Context) {
	body := c.Request().Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		log.Error(err)
		http.Error(writer, "Request had no body", http.StatusBadRequest)
		return
	}

	var params model.PreviewableStepParam
	err = json.Unmarshal(bodyBytes, &params)
	if err != nil {
		log.Error(err)
		http.Error(
			writer,
			fmt.Sprintf("Error parsing preview JSON: %s", err.Error()),
			http.StatusBadRequest,
		)
		return
	}

	p := map[string]string{
		"params": params.String(),
	}

	return c.JSON(http.StatusOK, p)
}
