// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of github.com/afjoseph/RAKE.Go package
//          implements the RAKE (Rapid Automatic Keyword Extraction) algorithm
//          by https://github.com/afjoseph/RAKE.Go
// modified: 2021-06-13

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	rake "github.com/afjoseph/RAKE.Go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// A Keyword represents an individual Keyword Candidate and its Score.
type Keyword struct {
	Candidate string  `json:"candidate"` // The Keyword.
	Score     float64 `json:"score"`     //The Keyword's Score.
}

var (
	logLevel   = getEnv("LOG_LEVEL", "1") // DEBUG
	serverPort = getEnv("RAKE_PORT", ":8080")
	apiKey     = getEnv("API_KEY", "ChangeMe")
	e          = echo.New()
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getHealth(c echo.Context) error {
	healthStatus := struct {
		Status string `json:"status"`
	}{"Up"}
	return c.JSON(http.StatusOK, healthStatus)
}

func getKeywords(c echo.Context) error {
	var keywords []Keyword

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		log.Errorf("json.NewDecoder Error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		text := jsonMap["text"]
		rakeCandidates := rake.RunRake(text.(string))
		for _, rakeCandidate := range rakeCandidates {
			keywords = append(keywords, Keyword{
				Candidate: rakeCandidate.Key,
				Score:     rakeCandidate.Value,
			})
		}
	}

	return c.JSON(http.StatusOK, keywords)
}

func run() error {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:X-API-Key",
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			e.Logger.Debugf("API_KEY: %v", apiKey)
			return key == apiKey, nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/keywords", getKeywords)

	// Start server
	return e.Start(serverPort)
}

func init() {
	level, _ := strconv.Atoi(logLevel)
	e.Logger.SetLevel(log.Lvl(level))
}

func main() {
	if err := run(); err != nil {
		e.Logger.Fatal(err)
		os.Exit(1)
	}
}
