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

	rake "github.com/garystafford/RAKE.Go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// A Keyword represents an individual keyword candidate and its score.
type Keyword struct {
	Candidate string  `json:"candidate"` // The keyword.
	Score     float64 `json:"score"`     //The keyword's score.
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
	port := os.Getenv("RAKE_PORT")

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == os.Getenv("AUTH_KEY"), nil
	}))

	// Routes
	e.POST("/keywords", handler)

	// Start server
	e.Logger.Fatal(e.Start(port))
}
