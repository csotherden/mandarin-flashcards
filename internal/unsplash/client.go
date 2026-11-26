package unsplash

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/csotherden/mandarin-flashcards/internal/models"
)

// ErrNotConfigured means UNSPLASH_ACCESS_KEY is unset.
var ErrNotConfigured = errors.New("unsplash is not configured")

type Client struct {
	accessKey string
	http      *http.Client
}

func New(accessKey string) *Client {
	return &Client{
		accessKey: strings.TrimSpace(accessKey),
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.accessKey != ""
}

type searchResponse struct {
	Results []struct {
		ID   string `json:"id"`
		URLs struct {
			Small   string `json:"small"`
			Regular string `json:"regular"`
		} `json:"urls"`
		Links struct {
			HTML             string `json:"html"`
			DownloadLocation string `json:"download_location"`
		} `json:"links"`
		User struct {
			Name  string `json:"name"`
			Links struct {
				HTML string `json:"html"`
			} `json:"links"`
		} `json:"user"`
	} `json:"results"`
}

// Search returns up to six photo candidates for a query.
func (c *Client) Search(ctx context.Context, query string) ([]models.ImageCandidate, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty search")
	}
	endpoint := "https://api.unsplash.com/search/photos?per_page=6&query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Client-ID "+c.accessKey)
	req.Header.Set("Accept-Version", "v1")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsplash: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unsplash status %d", res.StatusCode)
	}
	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make([]models.ImageCandidate, 0, len(parsed.Results))
	for _, item := range parsed.Results {
		out = append(out, models.ImageCandidate{
			ID:               item.ID,
			ThumbURL:         item.URLs.Small,
			FullURL:          item.URLs.Regular,
			PhotographerName: item.User.Name,
			PhotographerURL:  item.User.Links.HTML,
			PhotoURL:         item.Links.HTML,
			DownloadLocation: item.Links.DownloadLocation,
		})
	}
	return out, nil
}

// TrackDownload notifies Unsplash that a photo was chosen. Failure is non-fatal.
func (c *Client) TrackDownload(ctx context.Context, downloadLocation string) error {
	if !c.Enabled() || strings.TrimSpace(downloadLocation) == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadLocation, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Client-ID "+c.accessKey)
	req.Header.Set("Accept-Version", "v1")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<16))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("unsplash download status %d", res.StatusCode)
	}
	return nil
}
