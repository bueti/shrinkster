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
		Name: config.AppName,
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
		},
	}

	if err := app.cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (app *application) list(context *cli.Context) error {
	fmt.Println("List URLs")

	token, err := app.getToken(app.cfg.Email)
	if err != nil {
		app.logger.Error("impossible to get token", err)
		return err
	}

	app.client.Token = token

	var urlReq model.UrlByUserRequest
	urlReq.ID = app.cfg.ID

	marshalled, err := json.Marshal(urlReq)
	if err != nil {
		app.logger.Error("impossible to marshall", err)
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
		app.logger.Error("impossible to read all body of response: %s", err)
		return err
	}

	var urlsResp []model.UrlByUserResponse
	err = json.Unmarshal(resBody, &urlsResp)
	if err != nil {
		app.logger.Error("impossible to unmarshall response: %s", err)
		return err
	}

	fmt.Println("Short\t\t-> Original\tVisits")
	for _, url := range urlsResp {
		fmt.Println(fmt.Sprintf("- %s\t-> %s\t%d", url.ShortUrl, url.Original, url.Visits))
	}
	return nil

}

func (app *application) login(context *cli.Context) error {
	fmt.Println("Logging in to Shrinkster...")

	var userReq model.UserLoginRequest
	userReq.Email = context.String("username")
	userReq.Password = context.String("password")

	marshalled, err := json.Marshal(userReq)
	if err != nil {
		app.logger.Error("impossible to marshall", err)
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
		app.logger.Error("impossible to read all body of response: %s", err)
		return err
	}

	var userResp model.UserLoginResponse
	err = json.Unmarshal(resBody, &userResp)
	if err != nil {
		app.logger.Error("impossible to unmarshall response: %s", err)
		return err
	}

	// store the username
	err = config.Save(userResp)
	if err != nil {
		app.logger.Error("impossible to set username in config: %s", err)
		return err
	}

	err = app.setToken(context, userResp)
	if err != nil {
		app.logger.Error("impossible to set token in keyring: %s", err)
		return err
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
