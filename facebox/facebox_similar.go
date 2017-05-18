package facebox

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Similar checks the image in the io.Reader for similar faces.
func (c *Client) Similar(image io.Reader) ([]Similar, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "image.dat")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, image)
	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	req, err := http.NewRequest("POST", u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseSimilarResponse(resp.Body)
}

// SimilarURL checks the image at the specified URL for similar faces.
func (c *Client) SimilarURL(imageURL *url.URL) ([]Similar, error) {
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	if !imageURL.IsAbs() {
		return nil, errors.New("url must be absolute")
	}
	form := url.Values{}
	form.Set("url", imageURL.String())
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseSimilarResponse(resp.Body)
}

func (c *Client) parseSimilarResponse(r io.Reader) ([]Similar, error) {
	var similarResponse struct {
		Success bool
		Error   string
		Similar []Similar
	}
	if err := json.NewDecoder(r).Decode(&similarResponse); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	if !similarResponse.Success {
		return nil, ErrFacebox(similarResponse.Error)
	}
	return similarResponse.Similar, nil
}
