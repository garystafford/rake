// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of the RAKE algorithm, by https://github.com/afjoseph/RAKE.Go

package main

import (
	"encoding/json"
	rake "github.com/garystafford/RAKE.Go"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	port = ":8080"
)

// Keyword represents a candidate and its score
type Keyword struct {
	Candidate string  `json:"candidate"`
	Score     float64 `json:"score"`
}

func handler(c echo.Context) error {
	var keywords []Keyword

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
	} else {
		text := jsonMap["text"]
		candidates := rake.RunRake(text.(string))
		for _, candidate := range candidates {
			keywords = append(keywords, Keyword{Candidate: candidate.Key, Score: candidate.Value})
		}
	}

	return c.JSON(http.StatusOK, keywords)
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/keywords", handler)

	// Start server
	e.Logger.Fatal(e.Start(port))
}
