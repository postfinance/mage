package git

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const testData = "./testdata"

func TestNotag(t *testing.T) {
	f, err := os.Open(path.Join(testData, "notag.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	err = untar(testData, f)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path.Join(testData, "notag"))
	expected := &Info{
		Commit: Commit{
			Hash:      "6ea1790235210db239ddb8b6d0191db571f5bd64",
			ShortHash: "6ea17902",
			Timestamp: time.Time{},
			Author: author{
				Name:  "Rene Zbinden",
				Email: "rene.zbinden@postfinance.ch",
			},
		},
		Dirty: false,
		Tag: Tag{
			Name:  "",
			Count: 2,
		},
		s:    "6ea17902",
		tmpl: dfltTemplate,
	}

	t.Run("clean", func(t *testing.T) {
		i, err := New(path.Join(testData, "notag"))
		if err != nil {
			t.Fatal(err)
		}
		expectedTimestamp, _ := time.Parse(time.RFC1123Z, "Wed, 12 Sep 2018 08:00:11 +0200")
		if !expectedTimestamp.Equal(i.Commit.Timestamp) {
			t.Errorf("expected commit date is not %v but %v", expectedTimestamp, i.Commit.Timestamp)
		}
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})

	t.Run("dirty without dirty check", func(t *testing.T) {
		err := ioutil.WriteFile(path.Join(testData, "notag", "file2"), []byte("data"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		i, err := New(path.Join(testData, "notag"))
		if err != nil {
			t.Fatal(err)
		}
		expected.s = "6ea17902"
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})

	t.Run("dirty with dirty check", func(t *testing.T) {
		err := ioutil.WriteFile(path.Join(testData, "notag", "file2"), []byte("data"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		i, err := New(path.Join(testData, "notag"), WithDirtyCheck())
		if err != nil {
			t.Fatal(err)
		}
		expected.Dirty = true
		expected.dirty = true
		expected.s = "6ea17902-dirty"
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})
}

func TestTag(t *testing.T) {
	f, err := os.Open(path.Join(testData, "tag.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	err = untar(testData, f)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path.Join(testData, "tag"))
	expected := &Info{
		Commit: Commit{
			Hash:      "8183ffffe4851ee839927180a54eba94f5bc8396",
			ShortHash: "8183ffff",
			Timestamp: time.Time{},
			Author: author{
				Name:  "Rene Zbinden",
				Email: "rene.zbinden@postfinance.ch",
			},
		},
		Dirty: false,
		Tag: Tag{
			Name:  "v1.0.0",
			Count: 1,
		},
		s:    "v1.0.0-1-8183ffff",
		tmpl: dfltTemplate,
	}

	t.Run("clean", func(t *testing.T) {
		i, err := New(path.Join(testData, "tag"))
		if err != nil {
			t.Fatal(err)
		}
		expectedTimestamp, _ := time.Parse(time.RFC1123Z, "Wed, 12 Sep 2018 08:03:49 +0200")
		if !expectedTimestamp.Equal(i.Commit.Timestamp) {
			t.Errorf("expected commit date is not %v but %v", expectedTimestamp, i.Commit.Timestamp)
		}
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})

	t.Run("clean with pkgTemplate", func(t *testing.T) {
		i, err := New(path.Join(testData, "tag"), WithPackageTemplate())
		if err != nil {
			t.Fatal(err)
		}
		expected.s = "v1.0.0-1"
		expected.tmpl = pkgTemplate
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})

	t.Run("clean with customTemplate", func(t *testing.T) {
		const tmpl = "template"
		i, err := New(path.Join(testData, "tag"), WithTemplate(tmpl))
		if err != nil {
			t.Fatal(err)
		}
		expected.s = tmpl
		expected.tmpl = tmpl
		i.Commit.Timestamp = time.Time{} // time is never equal
		if !reflect.DeepEqual(i, expected) {
			t.Errorf("expected %#v is not equal %#v", expected, i)
		}
	})
}

func untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

func TestTemplates(t *testing.T) {
	withTag := Info{
		Commit: Commit{
			Hash:      "ffe4abc82c404c3726863cbb1a1459f67c8b5ed3",
			ShortHash: "ffe4abc8",
		},
		Tag: Tag{
			Name:  "0.1",
			Count: 10,
		},
	}
	noTag := Info{
		Commit: Commit{
			Hash:      "ffe4abc82c404c3726863cbb1a1459f67c8b5ed3",
			ShortHash: "ffe4abc8",
		},
		Tag: Tag{
			Count: 10,
		},
	}
	var tt = []struct {
		name     string
		info     Info
		tmpl     string
		expected string
	}{
		{"dfltTemplate - with tag",
			withTag,
			dfltTemplate,
			"0.1-10-ffe4abc8",
		},
		{"dfltTemplate - no tag",
			noTag,
			dfltTemplate,
			"ffe4abc8",
		},
		{"pkgTemplate - with tag",
			withTag,
			pkgTemplate,
			"0.1-10",
		},
		{"pkgTemplate - no tag",
			noTag,
			pkgTemplate,
			"10",
		},
		{"semVerTemplate - with tag",
			withTag,
			semVerTemplate,
			"0.1.10",
		},
		{"semVerTemplate - no tag",
			noTag,
			semVerTemplate,
			"0.0.10",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New("string").Parse(tc.tmpl)
			if err != nil {
				t.Fatal(err)
			}
			var val bytes.Buffer
			if err := tmpl.Execute(&val, tc.info); err != nil {
				if err != nil {
					t.Fatal(err)
				}
			}
			if string(val.Bytes()) != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, string(val.Bytes()))
			}
		})
	}

}
