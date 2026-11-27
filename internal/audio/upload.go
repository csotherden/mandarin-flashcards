package audio

import (
	"errors"
	"io"
	"net/http"
)

// MaxUpload is the largest recording accepted for grading.
const MaxUpload = 8_000_000

// UploadError carries a message fit to show the user.
type UploadError struct {
	Message string
	Err     error
}

func (e *UploadError) Error() string { return e.Message }
func (e *UploadError) Unwrap() error { return e.Err }

// FromForm reads the "audio" file of an already-parsed multipart form and converts it to the
// mono WAV the grader takes. Errors are *UploadError.
func FromForm(r *http.Request) ([]byte, error) {
	file, header, err := r.FormFile("audio")
	if err != nil {
		return nil, &UploadError{"Record audio before grading", err}
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, MaxUpload+1))
	if err != nil {
		return nil, &UploadError{"Could not read the recording", err}
	}
	if len(raw) == 0 || len(raw) > MaxUpload {
		return nil, &UploadError{"Recording must be under 8 MB", errors.New("bad recording size")}
	}
	wav, err := ToMonoWAV(r.Context(), raw, header.Header.Get("Content-Type"))
	if err != nil {
		return nil, &UploadError{"Could not read that recording. Try again", err}
	}
	return wav, nil
}
