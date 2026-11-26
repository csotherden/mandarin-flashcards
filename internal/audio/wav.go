package audio

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToMonoWAV re-encodes a browser recording to 16 kHz mono WAV.
// iOS Safari records audio/mp4 and Chrome records audio/webm; TonePerfect accepts
// both, and we still normalize so the grader always sees the same format.
func ToMonoWAV(ctx context.Context, input []byte, contentType string) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty audio")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg is required to normalize pronunciation audio")
	}

	dir, err := os.MkdirTemp("", "assess-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	inPath := filepath.Join(dir, "in"+extension(contentType))
	outPath := filepath.Join(dir, "out.wav")
	if err := os.WriteFile(inPath, input, 0o600); err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, "ffmpeg", "-y", "-i", inPath, "-ac", "1", "-ar", "16000", "-f", "wav", outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 400 {
			msg = msg[len(msg)-400:]
		}
		return nil, fmt.Errorf("ffmpeg: %w: %s", err, msg)
	}
	return os.ReadFile(outPath)
}

func extension(contentType string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "webm"):
		return ".webm"
	case strings.Contains(ct, "mp4"), strings.Contains(ct, "m4a"), strings.Contains(ct, "aac"):
		return ".m4a"
	case strings.Contains(ct, "ogg"):
		return ".ogg"
	case strings.Contains(ct, "mpeg"), strings.Contains(ct, "mp3"):
		return ".mp3"
	case strings.Contains(ct, "wav"):
		return ".wav"
	default:
		return ".audio"
	}
}
