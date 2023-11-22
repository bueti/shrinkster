package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bueti/shrinkster/internal/config"
	"github.com/bueti/shrinkster/internal/model"
	"github.com/bueti/shrinkster/internal/shrink"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"
)

type application struct {
	client *shrink.Client
	cli    *cli.App
	cfg    config.Config
	logger log.Logger
}

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	c := shrink.NewClient("")
	if os.Getenv("DEBUG") == "true" {
		logger.SetLevel(log.DebugLevel)
		c.Host = "https://localhost:8080"
		c.HttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.HttpClient.Timeout = 0
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Debug("no configuration found", err)
	}

	app := &application{
		client: c,
		cfg:    cfg,
		logger: *logger,
	}

	app.cli = &cli.App{
		Name:        config.AppName,
		Description: "Shrinkster (shrink.ch) is a URL shortener written in Go.",
		Commands: []*cli.Command{
			{
				Name:   "login",
				Usage:  "Login to Shrinkster",
				Action: app.login,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "username",
						Value: "",
						Usage: "Your shrink.ch username",
					},
					&cli.StringFlag{
						Name:  "password",
						Value: "",
						Usage: "Your shrink.ch password",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List current URLs",
				Action:  app.list,
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new URL",
				Action:  app.create,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "original",
						Value:    "",
						Usage:    "The original URL",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "short_code",
						Value: "",
						Usage: "The short code for the URL",
					},
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete an URL",
				Action:  app.delete,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Value: "",
						Usage: "The ID of the URL to delete",
					},
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print the version",
				Action:  app.version,
			},
		},
	}

	if err := app.cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// list all urls from the logged in user
func (app *application) list(context *cli.Context) error {
	fmt.Println("List URLs")

	token, err := app.getToken(app.cfg.Email)
	if err != nil {
		app.logger.Error("failed to get token", err)
		return err
	}

	app.client.Token = token

	var urlReq model.UrlByUserRequest
	urlReq.ID = app.cfg.ID

	marshalled, err := json.Marshal(urlReq)
	if err != nil {
		app.logger.Error("failed to marshall", err)
		return err
	}

	res, err := app.client.DoRequest("GET", fmt.Sprintf("/api/urls/%s", app.cfg.ID), bytes.NewReader(marshalled))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// check the response
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", res.Status)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		app.logger.Error("failed to read all body of response: %s", err)
		return err
	}

	var urlsResp []model.UrlByUserResponse
	err = json.Unmarshal(resBody, &urlsResp)
	if err != nil {
		app.logger.Error("failed to unmarshall response: %s", err)
		return err
	}

	fmt.Println("ID\t\t\t\t\tShort\t\tVisits\tOriginal")
	for _, url := range urlsResp {
		fmt.Println(fmt.Sprintf("- %s\t%s\t%d\t%s", url.ID, url.ShortUrl, url.Visits, url.Original))
	}
	return nil

}

// login to shrinkster
func (app *application) login(context *cli.Context) error {
	fmt.Println("Logging in to Shrinkster...")

	var userReq model.UserLoginRequest
	userReq.Email = context.String("username")
	userReq.Password = context.String("password")

	marshalled, err := json.Marshal(userReq)
	if err != nil {
		app.logger.Error("failed to marshall", err)
		return err
	}

	// create a new POST request with username and password
	res, err := app.client.DoRequest("POST", "/api/login", bytes.NewReader(marshalled))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// check the response
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", res.Status)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		app.logger.Error("failed to read body of response: %s", err)
		return err
	}

	var userResp model.UserLoginResponse
	err = json.Unmarshal(resBody, &userResp)
	if err != nil {
		app.logger.Error("failed to unmarshall response: %s", err)
		return err
	}

	// store the username
	err = config.Save(userResp)
	if err != nil {
		app.logger.Error("failed to set username in config: %s", err)
		return err
	}

	err = app.setToken(context, userResp)
	if err != nil {
		app.logger.Error("failed to set token in keyring: %s", err)
		return err
	}

	return nil
}

// create creates a new url and returns the short url
func (app *application) create(context *cli.Context) error {
	var urlReq model.UrlCreateRequest
	urlReq.Original = context.String("original")
	urlReq.ShortCode = context.String("short_code")
	urlReq.UserID = app.cfg.ID

	marshalled, err := json.Marshal(urlReq)
	if err != nil {
		app.logger.Error("failed to marshall", err)
		return err
	}

	token, err := app.getToken(app.cfg.Email)
	app.client.Token = token

	res, err := app.client.DoRequest("POST", "/api/urls", bytes.NewReader(marshalled))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// check the response
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("login failed: %s", res.Status)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var urlResp model.UrlResponse
	err = json.Unmarshal(resBody, &urlResp)
	if err != nil {
		return err
	}

	fmt.Println(urlResp.FullUrl)
	return nil
}

// delete an existing url
func (app *application) delete(context *cli.Context) error {
	var (
		err    error
		urlReq model.UrlDeleteRequest
	)

	urlReq.ID, err = uuid.Parse(context.String("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id: %s", err)
	}

	marshalled, err := json.Marshal(urlReq)
	if err != nil {
		app.logger.Error("failed to marshall", err)
		return err
	}

	token, err := app.getToken(app.cfg.Email)
	app.client.Token = token

	res, err := app.client.DoRequest("DELETE", "/api/urls", bytes.NewReader(marshalled))
	if err != nil {
		app.logger.Error("failed to delete url", err.Error())
		return err
	}
	defer res.Body.Close()

	// check the response
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("deletion failed: %s", res.Status)
	}

	return nil
}

func (app *application) setToken(context *cli.Context, userResp model.UserLoginResponse) error {
	if err := keyring.Set("shrinkster", context.String("username"), userResp.Token); err != nil {
		return err
	}
	return nil
}

func (app *application) getToken(username string) (string, error) {
	token, err := keyring.Get(config.AppName, username)
	if err != nil {
		fmt.Println("Can't find token. Please login first")
		return "", err
	}
	return token, nil
}
