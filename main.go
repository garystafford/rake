// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of the RAKE (Rapid Automatic Keyword Extraction) algorithm
//          by https://github.com/afjoseph/RAKE.Go

package main

import (
	"encoding/json"
	rake "github.com/afjoseph/RAKE.Go"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// A Keyword represents an individual Keyword Candidate and its Score.
type Keyword struct {
	Candidate string  `json:"candidate"` // The Keyword.
	Score     float64 `json:"score"`     //The Keyword's Score.
}

var (
	serverPort = ":" + getEnv("RAKE_PORT", "8080")
	apiKey     = getEnv("API_KEY", "")
	log        = logrus.New()

	// Echo instance
	e = echo.New()
)

func init() {
	log.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
}

func main() {
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
			log.Debugf("API_KEY: %v", apiKey)
			return key == apiKey, nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/keywords", getKeywords)

	// Start server
	e.Logger.Fatal(e.Start(serverPort))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getHealth(c echo.Context) error {
	var response interface{}
	err := json.Unmarshal([]byte(`{"status":"UP"}`), &response)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, response)
}

func getKeywords(c echo.Context) error {
	var keywords []Keyword

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
