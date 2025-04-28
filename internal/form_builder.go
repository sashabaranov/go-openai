package openai

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path"
	"strings"
)

type FormBuilder interface {
	CreateFormFile(fieldname string, file *os.File) error
	CreateFormFileReader(fieldname string, r io.Reader, filename string) error
	WriteField(fieldname, value string) error
	Close() error
	FormDataContentType() string
}

type DefaultFormBuilder struct {
	writer *multipart.Writer
}

func NewFormBuilder(body io.Writer) *DefaultFormBuilder {
	return &DefaultFormBuilder{
		writer: multipart.NewWriter(body),
	}
}

func (fb *DefaultFormBuilder) CreateFormFile(fieldname string, file *os.File) error {
	return fb.createFormFile(fieldname, file, file.Name())
}

func (fb *DefaultFormBuilder) CreateFormFileReader(fieldname string, r io.Reader, filename string) error {
	return fb.createFormFile(fieldname, r, path.Base(filename))
}


var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (fb *DefaultFormBuilder) createFormFile(fieldname string, r io.Reader, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", "image/png")
	fieldWriter, err := fb.writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(fieldWriter, r)
	if err != nil {
		return err
	}

	return nil
}

func (fb *DefaultFormBuilder) WriteField(fieldname, value string) error {
	return fb.writer.WriteField(fieldname, value)
}

func (fb *DefaultFormBuilder) Close() error {
	return fb.writer.Close()
}

func (fb *DefaultFormBuilder) FormDataContentType() string {
	return fb.writer.FormDataContentType()
}
