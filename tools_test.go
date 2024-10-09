package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

// TestTools_RandomString is a unit test function that validates the RandomString method of the Tools type.
func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string return")
	}
}

var uploadTests = []struct {
	name         string
	allowedTypes []string
	renameFile   bool
	expectedErr  bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/png", "image/jpeg"}, renameFile: false, expectedErr: false},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// set up pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer func(writer *multipart.Writer) {
				err := writer.Close()
				if err != nil {

				}
			}(writer)
			defer wg.Done()

			// create the form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/test.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/test.png")
			if err != nil {
				t.Error(err)
			}
			defer func(f *os.File) {
				err := f.Close()
				if err != nil {

				}
			}(f)

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error("error encoding image", err)
			}
		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Set("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.expectedErr {
			t.Error(err)
		}

		if !e.expectedErr {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Error("file not uploaded")
			}

			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.expectedErr && err != nil {
			t.Errorf("%s: error expected to none received", e.name)
		}

		wg.Wait()
	}
}
