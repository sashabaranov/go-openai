package openai

import (
	"io"
	"mime/multipart"
	"os"
)

type FormBuilder interface {
	CreateFormFile(fieldname string, file *os.File) error
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
	fieldWriter, err := fb.writer.CreateFormFile(fieldname, file.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(fieldWriter, file)
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
