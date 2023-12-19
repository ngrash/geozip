package postcode_test

import (
	"github.com/ngrash/postcode"
	"net/http"
	"os"
	"testing"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestFetchCountry_NotModified(t *testing.T) {
	const requestEtag = "current_etag"
	postcode.HTTPClient.Transport = RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if got, want := r.URL.String(), "https://download.geonames.org/export/zip/DE.zip"; got != want {
			t.Errorf("client requested %q, want %q", got, want)
		}
		if got, want := r.Method, http.MethodGet; got != want {
			t.Errorf("client made %s request, want %s", got, want)
		}
		if got, want := r.Header.Get("If-None-Match"), requestEtag; got != want {
			t.Errorf("client sent If-None-Match = %s, want %s", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusNotModified,
		}, nil
	})
	entries, modified, newEtag, err := postcode.FetchCountry("de", requestEtag)
	if entries != nil {
		t.Errorf("entries = %v, want nil", entries)
	}
	if modified {
		t.Error("modified = true, want false")
	}
	if got, want := newEtag, requestEtag; got != want {
		t.Errorf("newEtag = %v, want %v", got, want)
	}
	if err != nil {
		t.Error("err != nil")
	}
}

func TestFetchCountry_Modified(t *testing.T) {
	const (
		requestEtag  = "old_etag"
		responseEtag = "new_etag"
	)
	postcode.HTTPClient.Transport = RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if got, want := r.URL.String(), "https://download.geonames.org/export/zip/DE.zip"; got != want {
			t.Errorf("client requested %q, want %q", got, want)
		}
		if got, want := r.Method, http.MethodGet; got != want {
			t.Errorf("client made %s request, want %s", got, want)
		}
		if got, want := r.Header.Get("If-None-Match"), requestEtag; got != want {
			t.Errorf("client sent If-None-Match = %s, want %s", got, want)
		}
		file, err := os.Open("test_data/DE.zip")
		if err != nil {
			t.Fatal("open test data", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       file,
			Header: http.Header{
				"Etag": []string{responseEtag},
			},
		}, nil
	})
	entries, modified, newEtag, err := postcode.FetchCountry("de", requestEtag)
	if got, want := len(entries), 16477; got != want {
		t.Errorf("len(entries) = %v, want %v", got, want)
	}
	if !modified {
		t.Error("modified = false, want true")
	}
	if got, want := newEtag, responseEtag; got != want {
		t.Errorf("newEtag = %v, want %v", got, want)
	}
	if err != nil {
		t.Error("err != nil")
	}

	// Samples derived from GeoNames postal database. Licensed under Creative Commons Attribution 4.0 License,
	// see https://creativecommons.org/licenses/by/4.0/.
	samples := map[int][]string{
		10351: {"DE", "54668", "Ferschweiler", "Rheinland-Pfalz", "RP", "", "00", "Eifelkreis Bitburg-Prüm", "07232", "49.8667", "6.4", "4"},
		11193: {"DE", "56479", "Neustadt (Westerwald)", "Rheinland-Pfalz", "RP", "", "00", "Westerwaldkreis", "07143", "50.6333", "8.0333", ""},
	}

	for line, fields := range samples {
		idx := line - 1
		entry := entries[idx]

		if got, want := entry[postcode.CountryCode], fields[0]; got != want {
			t.Errorf("entries[%d]: CountryCode = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.PostalCode], fields[1]; got != want {
			t.Errorf("entries[%d]: PostalCode = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.PlaceName], fields[2]; got != want {
			t.Errorf("entries[%d]: PlaceName = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminName1], fields[3]; got != want {
			t.Errorf("entries[%d]: AdminName1 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminCode1], fields[4]; got != want {
			t.Errorf("entries[%d]: AdminCode1 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminName2], fields[5]; got != want {
			t.Errorf("entries[%d]: AdminName2 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminCode2], fields[6]; got != want {
			t.Errorf("entries[%d]: AdminCode2 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminName3], fields[7]; got != want {
			t.Errorf("entries[%d]: AdminName3 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.AdminCode3], fields[8]; got != want {
			t.Errorf("entries[%d]: AdminCode3 = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.Latitude], fields[9]; got != want {
			t.Errorf("entries[%d]: Latitude = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.Longitude], fields[10]; got != want {
			t.Errorf("entries[%d]: Longitude = %v, want %v", idx, got, want)
		}
		if got, want := entry[postcode.Accuracy], fields[11]; got != want {
			t.Errorf("entries[%d]: Accuracy = %v, want %v", idx, got, want)
		}
	}
}
