package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

func VerifyFilePath(filePath string) (err error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return errors.New("file does not exist")
	}	

	if file.IsDir() {
		return errors.New("filePath specified is a directory")
	}
	return nil
}

/***
	Automatically recale video to 9:16 for Shorts Video. This is intentional.
***/
func AddOverlayBackground(video Video, outputPath string) (err error) {
	err = VerifyFilePath(video.FilePath)
	if err != nil {
		return err
	}

	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}

	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-filter_complex", "[1]scale=1080:600[vid]; [0]scale=1080:1920[img]; [img][vid] overlay=(W-w)/2:(H-h)/2", "-acodec", "copy", outputPath)
	err = cmd.Run()
	if err != nil {
		logger.Error("failed to add background")
		return err
	}
	logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

/***
	Automatically recale video to 9:16 for Shorts Video. This is intentional.
***/
func AddOverlayBackgroundAndLogo(video Video, logoPath string, outputPath string) (err error) {
	err = VerifyFilePath(logoPath)
	if err != nil {
		return err
	}
	err = VerifyFilePath(video.FilePath)
	if err != nil {
		return err
	}

	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	
	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-i", logoPath, "-filter_complex", "[1]scale=1080:600[vid]; [0]crop=540:960:(in_w-540)/2:(in_h-960)/2,scale=1080:1920[img]; [img]boxblur=15[blurred]; [blurred][vid]overlay=(W-w)/2:(H-h)/2[with_logo]; [2:v]scale=320:220[logo_resized]; [with_logo][logo_resized]overlay=20:H-h-300", "-acodec", "copy", outputPath)

	_, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("failed to add background")
		return err
	}
	logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

func ConcatenateVideos(videos []Video, outputPath string) (err error) {
	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}

	inputTextFileName := fmt.Sprintf(`%s-temp-input.txt`, uuid.New().String()[0:8])
	inputTextFile, err := os.Create(inputTextFileName)
	if err != nil {
		return errors.New("unable to create a temp ")
	}

	defer os.Remove(inputTextFileName)

	for _, video := range videos {
		inputTextFile.WriteString(fmt.Sprintf(`file '%s'`, video.FilePath) + "\n")
	}
	inputTextFile.Close()

	cmd := exec.Command("ffmpeg", "-f", "concat", "-i", inputTextFileName, "-c", "copy", outputPath)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func ConcatenateVideosFromDir(directoryPath string, outputPath string) error {
	dir, err := os.Stat(directoryPath)
	if err != nil {
		logger.Error("unable to open directory")
		return err
	}

	if !dir.IsDir() {
		logger.Error("invalid directoryPath specified")
		return errors.New("invalid directory path")
	}

	filesInDir, err :=os.ReadDir(directoryPath)
	if err != nil {
		return err
	}
	
	videosInDir := []Video{}
	for _, file := range filesInDir {
		videoPath := fmt.Sprintf(`%s/%s`, directoryPath, file.Name())
		extension := filepath.Ext(videoPath)
		if extension == ".mp4" {
			video, err := Load(videoPath)
			if err != nil {
				logger.Info(fmt.Sprintf(`file: %s is used in concatenation since it is not a video`, filepath.Base(videoPath)))
				continue
			}
			videosInDir = append(videosInDir, video)
		}
	}

	err = ConcatenateVideos(videosInDir, outputPath)
	if err != nil {
		return err
	}
	return nil
}	