package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func (app *application) redirectUrlHandler(c echo.Context) error {
	wildcardValue := c.Param("*")
	shortUrl := strings.TrimSuffix(wildcardValue, "/")
	url, err := app.models.Urls.GetRedirect(shortUrl)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.Redirect(http.StatusPermanentRedirect, url.Original)
}

func (app *application) createUrlFormHandler(c echo.Context) error {
	data := app.newTemplateData(c)
	user, _ := app.userFromContext(c)
	data.User = user
	return c.Render(http.StatusOK, "create_url.tmpl.html", data)
}

func (app *application) createUrlHandlerPost(c echo.Context) error {
	original := c.FormValue("original")
	shortCode := c.FormValue("short_code")

	user, err := app.userFromContext(c)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Bad Request, are you logged in?")
		return c.Render(http.StatusBadRequest, "login.tmpl.html", app.newTemplateData(c))
	}

	urlReq := &model.UrlCreateRequest{
		Original:  original,
		ShortCode: shortCode,
		UserID:    user.ID,
	}
	url, err := app.models.Urls.Create(urlReq)
	if err != nil {
		switch err.Error() {
		case "url already exists":
			app.sessionManager.Put(c.Request().Context(), "flash_error", "URL already exists.")
		case "url cannot start with shrink.ch/s/":
			app.sessionManager.Put(c.Request().Context(), "flash_error", "URL cannot start with shrink.ch/s/")
		case "user id is required":
			app.sessionManager.Put(c.Request().Context(), "flash_error", "User ID is required.")
		case "short_url is too long":
			app.sessionManager.Put(c.Request().Context(), "flash_error", "Short URL is too long.")
		default:
			app.sessionManager.Put(c.Request().Context(), "flash_error", "Failed to create url.")
		}
		return c.Render(http.StatusBadRequest, "create_url.tmpl.html", app.newTemplateData(c))
	}

	qrCodeURL, err := app.createQRCode(genFullUrl(fmt.Sprintf(c.Scheme()+"://"+c.Request().Host), url.ShortUrl))
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Failed to create QR Code.")
	}

	err = app.models.Urls.SetQRCodeURL(&url, qrCodeURL)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Failed to set QR Code URL.")
	}

	app.sessionManager.Put(c.Request().Context(), "flash", "Url created successfully!")
	data := app.newTemplateData(c)
	data.User = user
	return app.dashboardHandler(c)
}

// createQRCode creates a QR Code for a given url. It returns the url to the QR Code.
func (app *application) createQRCode(original string) (string, error) {
	qrc, err := qrcode.New(original)
	if err != nil {
		return "", err
	}

	// generate a temporary file to store the image
	f, err := os.CreateTemp("", "qr-code-*.png")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())

	w, err := standard.New(f.Name())
	if err != nil {
		return "", err
	}

	// save file
	if err = qrc.Save(w); err != nil {
		fmt.Printf("could not save image: %v", err)
		return "", err
	}

	// upload file to bucket
	qrLocation, err := app.storeFileInBucket(f)
	if err != nil {
		fmt.Printf("could not upload file: %v", err)
		return "", err
	}

	return qrLocation, nil
}

// storeFileInBucket stores a file in a bucket.
func (app *application) storeFileInBucket(f *os.File) (string, error) {
	result, err := app.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(app.config.aws.bucket),
		Key:    aws.String(filepath.Base(f.Name())),
		Body:   f,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return aws.StringValue(&result.Location), nil
}

func (app *application) createUrlHandlerJsonPost(c echo.Context) error {
	urlReq := new(model.UrlCreateRequest)
	if err := c.Bind(urlReq); err != nil {
		return err
	}

	url, err := app.models.Urls.Create(urlReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, model.UrlResponse{
		ID:      url.ID,
		FullUrl: genFullUrl(fmt.Sprintf(c.Scheme()+"://"+c.Request().Host), url.ShortUrl),
	})
}

func (app *application) getUrlByUserHandlerJson(c echo.Context) error {
	userID := c.Param("user_id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	urls, err := app.models.Urls.GetUrlByUser(userUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, urls)
}

// deleteUrlHandlerPost handles the deletion of a url.
func (app *application) deleteUrlHandlerPost(c echo.Context) error {
	urlUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Bad Request?!")
		return app.dashboardHandler(c)
	}

	err = app.models.Urls.Delete(urlUUID)
	if err != nil {
		return err
	}

	app.sessionManager.Put(c.Request().Context(), "flash", "Url deleted successfully!")
	data := app.newTemplateData(c)
	user, _ := app.userFromContext(c)
	data.User = user
	return app.dashboardHandler(c)
}

// urlHandlerJsonDelete handles the deletion of a url via json.
func (app *application) urlHandlerJsonDelete(c echo.Context) error {
	urlReq := app.sessionManager.Get(c.Request().Context(), "urlReq").(*model.UrlDeleteRequest)

	err := app.models.Urls.Delete(urlReq.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, &model.UrlDeleteResponse{
		Message: "Url deleted successfully!",
	})
}

// genFullUrl generates the full url for a given short url
func genFullUrl(prefix, url string) string {
	return prefix + "/s/" + url
}
