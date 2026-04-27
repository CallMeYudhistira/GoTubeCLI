package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckFFmpeg checks if ffmpeg is installed and available in PATH
func CheckFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg is not installed or not in PATH. Please install ffmpeg to use this feature")
	}
	return nil
}

// MergeVideoAudio merges a video file and an audio file into a single output file
func MergeVideoAudio(videoPath, audioPath, outputPath string) error {
	if err := CheckFFmpeg(); err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg",
		"-y", // Overwrite output file without asking
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac", // Convert audio to aac for better compatibility in mp4
		outputPath,
	)

	// We don't want to show ffmpeg output unless there's an error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to merge video and audio: %s\n%s", err, string(output))
	}

	// Clean up temporary files
	_ = os.Remove(videoPath)
	_ = os.Remove(audioPath)

	return nil
}

// ConvertToMP3 converts an input audio/video file to MP3
func ConvertToMP3(inputPath, outputPath string) error {
	if err := CheckFFmpeg(); err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inputPath,
		"-q:a", "0", // Best quality for variable bitrate
		"-map", "a", // Map only audio
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert to MP3: %s\n%s", err, string(output))
	}

	// Clean up the original downloaded file
	_ = os.Remove(inputPath)

	return nil
}
