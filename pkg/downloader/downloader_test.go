/*
Copyright 2016 The Kubernetes Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package downloader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/helm/pkg/getter"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	. "github.com/rodcloutier/draft-packs/pkg/getter"
	"github.com/rodcloutier/draft-packs/pkg/repo"
	"github.com/rodcloutier/draft-packs/pkg/repo/repotest"
)

func getterAll() (result getter.Providers) {
	result = getter.Providers{
		{
			Schemes: []string{"http", "https"},
			New:     NewHTTPGetter,
		},
	}
	return result
}

func TestResolveRef(t *testing.T) {
	tests := []struct {
		name, ref, expect, version string
		fail                       bool
	}{
		{name: "full URL", ref: "http://example.com/foo-1.2.3.tgz", expect: "http://example.com/foo-1.2.3.tgz"},
		{name: "full URL, HTTPS", ref: "https://example.com/foo-1.2.3.tgz", expect: "https://example.com/foo-1.2.3.tgz"},
		{name: "full URL, with authentication", ref: "http://username:password@example.com/foo-1.2.3.tgz", expect: "http://username:password@example.com/foo-1.2.3.tgz"},
		{name: "reference, testing repo", ref: "testing/alpine", expect: "http://example.com/alpine-1.2.3.tgz"},
		{name: "reference, version, testing repo", ref: "testing/alpine", version: "0.2.0", expect: "http://example.com/alpine-0.2.0.tgz"},
		{name: "reference, version, malformed repo", ref: "malformed/alpine", version: "1.2.3", expect: "http://dl.example.com/alpine-1.2.3.tgz"},
		{name: "full URL, HTTPS, irrelevant version", ref: "https://example.com/foo-1.2.3.tgz", version: "0.1.0", expect: "https://example.com/foo-1.2.3.tgz", fail: true},
		{name: "full URL, file", ref: "file:///foo-1.2.3.tgz", fail: true},
		{name: "invalid", ref: "invalid-1.2.3", fail: true},
		{name: "not found", ref: "nosuchthing/invalid-1.2.3", fail: true},
	}

	c := Downloader{
		Home:    draftpath.NewHome("testdata/helmhome"),
		Out:     os.Stderr,
		Getters: getterAll(),
	}

	for _, tt := range tests {
		u, _, err := c.ResolveVersion(tt.ref, tt.version)
		if err != nil {
			if tt.fail {
				continue
			}
			t.Errorf("%s: failed with error %s", tt.name, err)
			continue
		}
		if got := u.String(); got != tt.expect {
			t.Errorf("%s: expected %s, got %s", tt.name, tt.expect, got)
		}
	}
}

func TestVerifyFile(t *testing.T) {
	v, err := VerifyFile("testdata/signtest-0.1.0.tgz", "testdata/helm-test-key.pub")
	if err != nil {
		t.Fatal(err)
	}
	// The verification is tested at length in the provenance package. Here,
	// we just want a quick sanity check that the v is not empty.
	if len(v.FileHash) == 0 {
		t.Error("Digest missing")
	}
}

func TestIsTar(t *testing.T) {
	tests := map[string]bool{
		"foo.tgz":           true,
		"foo/bar/baz.tgz":   true,
		"foo-1.2.3.4.5.tgz": true,
		"foo.tar.gz":        false, // for our purposes
		"foo.tgz.1":         false,
		"footgz":            false,
	}

	for src, expect := range tests {
		if isTar(src) != expect {
			t.Errorf("%q should be %t", src, expect)
		}
	}
}

func TestDownloadTo(t *testing.T) {
	tmp, err := ioutil.TempDir("", "draft-downloadto-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	hh := draftpath.NewHome(tmp)
	dest := filepath.Join(hh.String(), "dest")
	configDirectories := []string{
		hh.String(),
		hh.Repository(),
		hh.Cache(),
		dest,
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			if err := os.MkdirAll(p, 0755); err != nil {
				t.Fatalf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			t.Fatalf("%s must be a directory", p)
		}
	}

	// Set up a fake repo
	srv := repotest.NewServer(tmp)
	defer srv.Stop()
	if _, err := srv.Copy("testdata/*.tgz*"); err != nil {
		t.Error(err)
		return
	}
	if err := srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	c := Downloader{
		Home:    hh,
		Out:     os.Stderr,
		Verify:  VerifyAlways,
		Keyring: "testdata/helm-test-key.pub",
		Getters: getterAll(),
	}
	cname := "/signtest-0.1.0.tgz"
	where, v, err := c.DownloadTo(srv.URL()+cname, "", dest)
	if err != nil {
		t.Error(err)
		return
	}

	if expect := filepath.Join(dest, cname); where != expect {
		t.Errorf("Expected download to %s, got %s", expect, where)
	}

	if v.FileHash == "" {
		t.Error("File hash was empty, but verification is required.")
	}

	if _, err := os.Stat(filepath.Join(dest, cname)); err != nil {
		t.Error(err)
		return
	}
}

// func TestDownloadTo_VerifyLater(t *testing.T) {
// 	tmp, err := ioutil.TempDir("", "helm-downloadto-")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(tmp)

// 	hh := helmpath.Home(tmp)
// 	dest := filepath.Join(hh.String(), "dest")
// 	configDirectories := []string{
// 		hh.String(),
// 		hh.Repository(),
// 		hh.Cache(),
// 		dest,
// 	}
// 	for _, p := range configDirectories {
// 		if fi, err := os.Stat(p); err != nil {
// 			if err := os.MkdirAll(p, 0755); err != nil {
// 				t.Fatalf("Could not create %s: %s", p, err)
// 			}
// 		} else if !fi.IsDir() {
// 			t.Fatalf("%s must be a directory", p)
// 		}
// 	}

// 	// Set up a fake repo
// 	srv := repotest.NewServer(tmp)
// 	defer srv.Stop()
// 	if _, err := srv.CopyCharts("testdata/*.tgz*"); err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if err := srv.LinkIndices(); err != nil {
// 		t.Fatal(err)
// 	}

// 	c := Downloader{
// 		Home:    hh,
// 		Out:     os.Stderr,
// 		Verify:  VerifyLater,
// 		Getters: getter.All(environment.EnvSettings{}),
// 	}
// 	cname := "/signtest-0.1.0.tgz"
// 	where, _, err := c.DownloadTo(srv.URL()+cname, "", dest)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if expect := filepath.Join(dest, cname); where != expect {
// 		t.Errorf("Expected download to %s, got %s", expect, where)
// 	}

// 	if _, err := os.Stat(filepath.Join(dest, cname)); err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if _, err := os.Stat(filepath.Join(dest, cname+".prov")); err != nil {
// 		t.Error(err)
// 		return
// 	}
// }

func TestScanReposForURL(t *testing.T) {
	hh := draftpath.NewHome("testdata/helmhome")
	c := Downloader{
		Home:    hh,
		Out:     os.Stderr,
		Verify:  VerifyLater,
		Getters: getterAll(),
	}

	u := "http://example.com/alpine-0.2.0.tgz"
	rf, err := repo.LoadRepositoriesFile(c.Home.RepositoryFile())
	if err != nil {
		t.Fatal(err)
	}

	entry, err := c.scanReposForURL(u, rf)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Name != "testing" {
		t.Errorf("Unexpected repo %q for URL %q", entry.Name, u)
	}

	// A lookup failure should produce an ErrNoOwnerRepo
	u = "https://no.such.repo/foo/bar-1.23.4.tgz"
	if _, err = c.scanReposForURL(u, rf); err != ErrNoOwnerRepo {
		t.Fatalf("expected ErrNoOwnerRepo, got %v", err)
	}
}
