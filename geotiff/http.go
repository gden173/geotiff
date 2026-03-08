package geotiff

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var errRangeNotSupported = errors.New("server does not support HTTP range requests")

// httpReadSeeker implements io.ReadSeeker for remote files using HTTP range
// requests, enabling cloud-optimized access to GeoTIFF files stored in remote
// object storage (e.g., S3, GCS, Azure Blob) without downloading the entire
// file.
type httpReadSeeker struct {
	url    string
	offset int64
	size   int64
	client *http.Client
}

// newHTTPReadSeeker creates a new httpReadSeeker for the given URL.
// It issues a HEAD request to determine the content length of the remote file.
func newHTTPReadSeeker(url string, client *http.Client) (*httpReadSeeker, error) {
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Head(url)
	if err != nil {
		return nil, fmt.Errorf("failed to HEAD URL %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP HEAD returned status %d for URL %q", resp.StatusCode, url)
	}

	size := resp.ContentLength
	if size < 0 {
		return nil, fmt.Errorf("could not determine content length for URL %q", url)
	}

	return &httpReadSeeker{
		url:    url,
		offset: 0,
		size:   size,
		client: client,
	}, nil
}

// Read implements io.Reader using HTTP range requests.
func (h *httpReadSeeker) Read(p []byte) (int, error) {
	if h.offset >= h.size {
		return 0, io.EOF
	}

	end := h.offset + int64(len(p)) - 1
	if end >= h.size {
		end = h.size - 1
	}

	req, err := http.NewRequest(http.MethodGet, h.url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", h.offset, end))

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed HTTP range request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("%w (status %d)", errRangeNotSupported, resp.StatusCode)
	}

	nBytes := end - h.offset + 1
	n, err := io.ReadFull(resp.Body, p[:nBytes])
	h.offset += int64(n)
	return n, err
}

// Seek implements io.Seeker.
func (h *httpReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = h.offset + offset
	case io.SeekEnd:
		newOffset = h.size + offset
	default:
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("seek to negative position %d", newOffset)
	}
	h.offset = newOffset
	return h.offset, nil
}

// ReadFromURL reads a GeoTIFF from a remote URL using HTTP range requests.
// This enables cloud-optimized access to GeoTIFF files stored in remote object
// storage without downloading the entire file upfront.
//
// If client is nil, http.DefaultClient is used.
func ReadFromURL(url string, client *http.Client) (*GeoTIFF, error) {
	rs, err := newHTTPReadSeeker(url, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP reader for %q: %w", url, err)
	}
	return Read(rs)
}
