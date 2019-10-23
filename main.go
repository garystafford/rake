// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of the RAKE (Rapid Automatic Keyword Extraction) algorithm
//          by https://github.com/afjoseph/RAKE.Go

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	rake "github.com/garystafford/RAKE.Go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// A Keyword represents an individual Keyword Candidate and its Score.
type Keyword struct {
	Candidate string  `json:"candidate"` // The Keyword.
	Score     float64 `json:"score"`     //The Keyword's Score.
}

var (
	portClient = os.Getenv("RAKE_PORT")
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == os.Getenv("AUTH_KEY"), nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/keywords", getKeywords)

	// Start server
	e.Logger.Fatal(e.Start(portClient))
}

func getHealth(c echo.Context) error {
	var response interface{}
	err := json.Unmarshal([]byte(`{"status":"UP"}`), &response)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response)
}

func getKeywords(c echo.Context) error {
	var keywords []Keyword

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
	} else {
		text := jsonMap["text"]
		rakeCandidates := rake.RunRake(text.(string))
		for _, rakeCandidate := range rakeCandidates {
			keywords = append(keywords, Keyword{
				Candidate: rakeCandidate.Key,
				Score: rakeCandidate.Value,
			})
		}
	}

	return c.JSON(http.StatusOK, keywords)
}
