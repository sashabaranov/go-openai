package openai

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestAudioWithFailingFormBuilder(t *testing.T) {
	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()
	path := filepath.Join(dir, "fake.mp3")
	test.CreateTestFile(t, path)

	req := AudioRequest{
		FilePath:    path,
		Prompt:      "test",
		Temperature: 0.5,
		Language:    "en",
		Format:      AudioResponseFormatSRT,
	}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder := &mockFormBuilder{}

	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return mockFailedErr
	}
	err := audioMultipartForm(req, mockBuilder)
	checks.ErrorIs(t, err, mockFailedErr, "audioMultipartForm should return error if form builder fails")

	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return nil
	}

	var failForField string
	mockBuilder.mockWriteField = func(fieldname, _ string) error {
		if fieldname == failForField {
			return mockFailedErr
		}
		return nil
	}

	failOn := []string{"model", "prompt", "temperature", "language", "response_format"}
	for _, failingField := range failOn {
		failForField = failingField
		mockFailedErr = fmt.Errorf("mock form builder fail on field %s", failingField)

		err = audioMultipartForm(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "audioMultipartForm should return error if form builder fails")
	}
}

func TestCreateFileField(t *testing.T) {
	t.Run("createFileField failing file", func(t *testing.T) {
		dir, cleanup := test.CreateTestDirectory(t)
		defer cleanup()
		path := filepath.Join(dir, "fake.mp3")
		test.CreateTestFile(t, path)

		req := AudioRequest{
			FilePath: path,
		}

		mockFailedErr := fmt.Errorf("mock form builder fail")
		mockBuilder := &mockFormBuilder{
			mockCreateFormFile: func(string, *os.File) error {
				return mockFailedErr
			},
		}

		err := createFileField(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "createFileField using a file should return error if form builder fails")
	})

	t.Run("createFileField failing reader", func(t *testing.T) {
		req := AudioRequest{
			FilePath: "test.wav",
			Reader:   bytes.NewBuffer([]byte(`wav test contents`)),
		}

		mockFailedErr := fmt.Errorf("mock form builder fail")
		mockBuilder := &mockFormBuilder{
			mockCreateFormFileReader: func(string, io.Reader, string) error {
				return mockFailedErr
			},
		}

		err := createFileField(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "createFileField using a reader should return error if form builder fails")
	})

	t.Run("createFileField failing open", func(t *testing.T) {
		req := AudioRequest{
			FilePath: "non_existing_file.wav",
		}

		mockBuilder := &mockFormBuilder{}

		err := createFileField(req, mockBuilder)
		checks.HasError(t, err, "createFileField using file should return error when open file fails")
	})
}

func Test_validateSpeed(t *testing.T) {
	type args struct {
		speed float32
	}
	tests := []struct {
		name    string
		args    args
		wantOk  bool
		wantErr bool
	}{
		{
			name: "validateSpeed should fail when speed is less than 0.25",
			args: args{
				speed: 0.24,
			},
			wantOk:  false,
			wantErr: true,
		},
		{
			name: "validateSpeed should fail when speed is more that 4.0",
			args: args{
				speed: 4.00001,
			},
			wantOk:  false,
			wantErr: true,
		},
		{
			name: "validateSpeed should pass when speed 0.25",
			args: args{
				speed: 0.25,
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name: "validateSpeed should pass when speed 4.0",
			args: args{
				speed: 4.0,
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name: "validateSpeed should pass when speed is inbetween 0.25 and 4.0",
			args: args{
				speed: 1.2,
			},
			wantOk:  true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, err := validateSpeed(tt.args.speed)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("validateSpeed() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestWithSpeed(t *testing.T) {
	type args struct {
		speed float32
	}
	tests := []struct {
		name string
		args args
		want SpeechRequest
	}{
		{
			name: "WithSpeed should set speed on request",
			args: args{
				speed: 100.1,
			},
			want: SpeechRequest{
				Speed: 100.1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SpeechRequest{
				Speed: -1.0,
			}
			if WithSpeed(tt.args.speed)(&request); !reflect.DeepEqual(request, tt.want) {
				t.Errorf("WithSpeed() = %v, want %v", request, tt.want)
			}
		})
	}
}

func TestWithResponseFormat(t *testing.T) {
	type args struct {
		format AudioSpeechResponseFormat
	}
	tests := []struct {
		name string
		args args
		want SpeechRequest
	}{
		{
			name: "WithResponseFormat should set Request ResponseFormat",
			args: args{
				AudioSpeachResponseFlac,
			},
			want: SpeechRequest{
				ResponseFormat: AudioSpeachResponseFlac,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SpeechRequest{
				ResponseFormat: AudioSpeachResponseFlac,
			}
			if WithResponseFormat(tt.args.format)(&request); !reflect.DeepEqual(request, tt.want) {
				t.Errorf("WithResponseFormat() = %v, want %v", request, tt.want)
			}
		})
	}
}

func TestNewSpeechRequest(t *testing.T) {
	type args struct {
		text  string
		model AudioSpeechModel
		voice AudioSpeechVoice
		opts  []speechRequestOption
	}
	tests := []struct {
		name string
		args args
		want SpeechRequest
	}{
		{
			name: "NewSpeechRequest without options",
			args: args{
				text: "test",
				model: AudioSpeachModelTTS1,
				voice: AudioVoiceFable,
			},
			want: SpeechRequest{
				Prompt: "test",
				Model: AudioSpeachModelTTS1,
				Voice: AudioVoiceFable,
				Speed: 1.0,
				ResponseFormat: AudioSpeachResponseMp3,
			},
		},
		{
			name: "NewSpeechRequest with speed 2.0",
			args: args{
				text: "test",
				model: AudioSpeachModelTTS1,
				voice: AudioVoiceFable,
				opts: []speechRequestOption{
					WithSpeed(2.0),
				},
			},
			want: SpeechRequest{
				Prompt: "test",
				Model: AudioSpeachModelTTS1,
				Voice: AudioVoiceFable,
				Speed: 2.0,
				ResponseFormat: AudioSpeachResponseMp3,
			},
		},
		{
			name: "NewSpeechRequest with output Flac",
			args: args{
				text: "test",
				model: AudioSpeachModelTTS1,
				voice: AudioVoiceFable,
				opts: []speechRequestOption{
					WithResponseFormat(AudioSpeachResponseFlac),
				},
			},
			want: SpeechRequest{
				Prompt: "test",
				Model: AudioSpeachModelTTS1,
				Voice: AudioVoiceFable,
				Speed: 1.0,
				ResponseFormat: AudioSpeachResponseFlac,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSpeechRequest(tt.args.text, tt.args.model, tt.args.voice, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpeechRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
