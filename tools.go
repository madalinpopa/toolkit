package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// randomSourceString is a collection of characters used as the basis for generating random strings.
var randomSourceString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is a utility type that provides various helper methods.
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString generates a random string of the specified length using characters from randomSourceString.
func (t *Tools) RandomString(length int) string {

	s, r := make([]rune, length), []rune(randomSourceString)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// UploadedFile represents a file that has been uploaded, including its original and new names and its size.
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}
	return files[0], nil

}

// UploadFiles uploads files from a multipart form request to a specified directory.
// r: The HTTP request containing the multipart form data.
// uploadDir: The directory where the uploaded files will be saved.
// rename: Optional boolean indicating whether to rename the uploaded files. Defaults to true if not specified.
// Returns a slice of pointers to UploadedFile and an error if any occurs.
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}
	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file exceeds the maximum file size")
	}

	for _, files := range r.MultipartForm.File {
		for _, file := range files {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := file.Open()
				if err != nil {
					return nil, err
				}
				defer func(infile multipart.File) {
					err := infile.Close()
					if err != nil {
						fmt.Println("error closing file", err)
					}
				}(infile)

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, allowedFileType := range t.AllowedFileTypes {
						if strings.EqualFold(allowedFileType, fileType) {
							allowed = true
							break
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s.%s", t.RandomString(25), filepath.Ext(file.Filename))
					uploadedFile.OriginalFileName = file.Filename
				} else {
					uploadedFile.NewFileName = file.Filename
					uploadedFile.OriginalFileName = file.Filename
				}

				var outFile *os.File

				if outFile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outFile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}
				defer func(outFile *os.File) {
					err := outFile.Close()
					if err != nil {
						fmt.Println("error closing file", err)
					}
				}(outFile)

				uploadedFiles = append(uploadedFiles, &uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}
