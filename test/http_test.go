// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var jsonEndpoints = []string{
	"/rest/db/completion?device=I6KAH76-66SLLLB-5PFXSOA-UFJCDZC-YAOMLEK-CP2GB32-BV5RQST-3PSROAU&folder=default",
	"/rest/db/ignores?folder=default",
	"/rest/db/need?folder=default",
	"/rest/db/status?folder=default",
	"/rest/db/browse?folder=default",
	"/rest/events?since=-1&limit=5",
	"/rest/stats/device",
	"/rest/stats/folder",
	"/rest/svc/deviceid?id=I6KAH76-66SLLLB-5PFXSOA-UFJCDZC-YAOMLEK-CP2GB32-BV5RQST-3PSROAU",
	"/rest/svc/lang",
	"/rest/svc/report",
	"/rest/system/browse?current=.",
	"/rest/system/config",
	"/rest/system/config/insync",
	"/rest/system/connections",
	"/rest/system/discovery",
	"/rest/system/error",
	"/rest/system/ping",
	"/rest/system/status",
	"/rest/system/upgrade",
	"/rest/system/version",
}

func TestGetIndex(t *testing.T) {
	st := syncthingProcess{
		argv:     []string{"-home", "h2"},
		port:     8082,
		instance: "2",
	}
	err := st.start()
	if err != nil {
		t.Fatal(err)
	}
	defer st.stop()

	res, err := st.get("/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Status %d != 200", res.StatusCode)
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) < 1024 {
		t.Errorf("Length %d < 1024", len(bs))
	}
	if !bytes.Contains(bs, []byte("</html>")) {
		t.Error("Incorrect response")
	}
	if res.Header.Get("Set-Cookie") == "" {
		t.Error("No set-cookie header")
	}
	res.Body.Close()

	res, err = st.get("/")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Status %d != 200", res.StatusCode)
	}
	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) < 1024 {
		t.Errorf("Length %d < 1024", len(bs))
	}
	if !bytes.Contains(bs, []byte("</html>")) {
		t.Error("Incorrect response")
	}
	if res.Header.Get("Set-Cookie") == "" {
		t.Error("No set-cookie header")
	}
	res.Body.Close()
}

func TestGetIndexAuth(t *testing.T) {
	st := syncthingProcess{
		argv:     []string{"-home", "h1"},
		port:     8081,
		instance: "1",
		apiKey:   "abc123",
	}
	err := st.start()
	if err != nil {
		t.Fatal(err)
	}
	defer st.stop()

	// Without auth should give 401

	res, err := http.Get("http://127.0.0.1:8081/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 401 {
		t.Errorf("Status %d != 401", res.StatusCode)
	}

	// With wrong username/password should give 401

	req, err := http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("testuser", "wrongpass")

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatalf("Status %d != 401", res.StatusCode)
	}

	// With correct username/password should succeed

	req, err = http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("testuser", "testpass")

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("Status %d != 200", res.StatusCode)
	}
}

func TestGetJSON(t *testing.T) {
	st := syncthingProcess{
		argv:     []string{"-home", "h2"},
		port:     8082,
		instance: "2",
	}
	err := st.start()
	if err != nil {
		t.Fatal(err)
	}
	defer st.stop()

	for _, path := range jsonEndpoints {
		res, err := st.get(path)
		if err != nil {
			t.Error(err)
		}

		if ct := res.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("Incorrect Content-Type %q for %q", ct, path)
		}

		var intf interface{}
		err = json.NewDecoder(res.Body).Decode(&intf)
		res.Body.Close()

		if err != nil {
			t.Error(err)
		}
	}
}

func TestPOSTWithoutCSRF(t *testing.T) {
	st := syncthingProcess{
		argv:     []string{"-home", "h2"},
		port:     8082,
		instance: "2",
	}
	err := st.start()
	if err != nil {
		t.Fatal(err)
	}
	defer st.stop()

	// Should fail without CSRF

	req, err := http.NewRequest("POST", "http://127.0.0.1:8082/rest/system/error/clear", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 403 {
		t.Fatalf("Status %d != 403 for POST", res.StatusCode)
	}

	// Get CSRF

	req, err = http.NewRequest("GET", "http://127.0.0.1:8082/", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	hdr := res.Header.Get("Set-Cookie")
	if !strings.Contains(hdr, "CSRF-Token") {
		t.Error("Missing CSRF-Token in", hdr)
	}

	// Should succeed with CSRF

	req, err = http.NewRequest("POST", "http://127.0.0.1:8082/rest/system/error/clear", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-CSRF-Token", hdr[len("CSRF-Token="):])
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("Status %d != 200 for POST", res.StatusCode)
	}

	// Should fail with incorrect CSRF

	req, err = http.NewRequest("POST", "http://127.0.0.1:8082/rest/system/error/clear", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-CSRF-Token", hdr[len("CSRF-Token="):]+"X")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 403 {
		t.Fatalf("Status %d != 403 for POST", res.StatusCode)
	}
}
