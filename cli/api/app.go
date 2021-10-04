package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (c *Client) UploadApp(filepath string) (*models.App, error) {
	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"test_target":  mustOpen(filepath),
	}

	return c.uploadApp(values)
}

func (c *Client) uploadApp(values map[string]io.Reader) (*models.App, error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		var err error
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/app/upload", c.BaseURL), &b)
	if err != nil {
		return nil, err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check the response
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("bad status: %s", res.Status)
		return nil, err
	}

	var app models.App
	if err = json.NewDecoder(res.Body).Decode(&app); err != nil {
		return nil, err
	}


	return &app, nil
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}