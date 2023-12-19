// Package postcode provides functionality specifically for downloading and parsing
// postal code data from the GeoNames geographical database (https://www.geonames.org/).
package postcode

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is a global http.Client instance used for making HTTP requests.
// This can be replaced or configured as needed to change the default HTTP behavior.
var HTTPClient http.Client

// Entry represents a single postal code entry. It is an array of 12 strings, each representing a specific field of data.
type Entry [12]string

// Field represents a specific field in a postal code entry.
type Field int

const (
	// CountryCode is the index for the country code in a postal code entry.
	CountryCode Field = iota
	// PostalCode is the index for the postal code in a postal code entry.
	PostalCode
	// PlaceName is the index for the place name in a postal code entry.
	PlaceName
	// AdminName1 is the index for the first level of administrative division name in a postal code entry.
	AdminName1
	// AdminCode1 is the index for the first level of administrative division code in a postal code entry.
	AdminCode1
	// AdminName2 is the index for the second level of administrative division name in a postal code entry.
	AdminName2
	// AdminCode2 is the index for the second level of administrative division code in a postal code entry.
	AdminCode2
	// AdminName3 is the index for the third level of administrative division name in a postal code entry.
	AdminName3
	// AdminCode3 is the index for the third level of administrative division code in a postal code entry.
	AdminCode3
	// Latitude is the index for the latitude in a postal code entry.
	Latitude
	// Longitude is the index for the longitude in a postal code entry.
	Longitude
	// Accuracy is the index for the accuracy in a postal code entry.
	Accuracy
)

// FetchCountry fetches postal code entries for a specific country code from the GeoNames database.
// It leverages the HTTP ETag mechanism to minimize data transfer for unchanged postal code data.
//
// The function takes two arguments:
//
//	cc: The country code for which postal code data is to be fetched.
//	etag: An ETag value from a previous request to this function.
//
// If the data for the given country code has not changed since the last request with the provided ETag,
// the function returns with 'modified' set to false, and no new data is fetched.
//
// If the data has changed, or if this is the first request (indicated by an empty etag),
// the function fetches the updated data, sets 'modified' to true, and returns the new data along with the new ETag.
//
// Example usage:
//
//	entries, modified, newEtag, err := FetchCountry("US", previousEtag)
//	if err != nil {
//	    // Handle error
//	}
//	if modified {
//	    // Process new entries
//	    // Save newEtag for future requests
//	}
//
// See https://download.geonames.org/export/zip/ for a list of available countries.
func FetchCountry(cc, etag string) (entries []Entry, modified bool, newEtag string, err error) {
	cc, err = normalizeCountryCode(cc)
	if err != nil {
		return
	}

	url := downloadURL(cc)
	zipData, modified, newEtag, err := download(url, etag)
	if !modified || err != nil {
		return
	}

	filename := zippedFile(cc)
	csvData, err := unzipFile(zipData, filename)
	if err != nil {
		return
	}

	entries, err = parseCSV(csvData)

	return
}

func normalizeCountryCode(cc string) (string, error) {
	r := strings.ToUpper(cc)
	if got, want := len(cc), 2; got != want {
		return r, fmt.Errorf("country code %q has %d bytes, want %d", cc, got, want)
	}
	return r, nil
}

func downloadURL(cc string) string {
	return fmt.Sprintf("https://download.geonames.org/export/zip/%s.zip", cc)
}

func download(url, etag string) ([]byte, bool, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, "", err
	}
	req.Header.Add("If-None-Match", etag)
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, false, "", err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)

	if resp.StatusCode == http.StatusNotModified {
		// No new codes and no error.
		return nil, false, etag, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, "", fmt.Errorf("status = %s, want 200", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, "", fmt.Errorf("read response body: %w", err)
	}

	return body, true, resp.Header.Get("Etag"), nil
}

func zippedFile(cc string) string {
	return fmt.Sprintf("%s.txt", cc)
}

func unzipFile(data []byte, filename string) (_ []byte, err error) {
	unzip, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("create unzipping reader: %w", err)
	}
	var file *zip.File
	for _, f := range unzip.File {
		if f.Name == filename {
			file = f
			break
		}
	}
	if file == nil {
		return nil, fmt.Errorf("zipfile missing %s", filename)
	}

	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open zipped %s: %w", filename, err)
	}
	defer func(rc io.ReadCloser) {
		err = errors.Join(err, rc.Close())
	}(rc)

	return io.ReadAll(rc)
}

func parseCSV(data []byte) ([]Entry, error) {
	r := bytes.NewReader(data)
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	table, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	es := make([]Entry, len(table))
	for i, columns := range table {
		var e Entry
		for ii, col := range columns {
			e[ii] = col
		}
		es[i] = e
	}
	return es, nil
}
