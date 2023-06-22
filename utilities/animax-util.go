package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pichan321/animax"
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
func AddOverlayBackground(video animax.Video, outputPath string) (err error) {
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
	logger := animax.GetLogger()
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
func AddOverlayBackgroundAndLogo(video animax.Video, logoPath string, outputPath string) (err error) {
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
	logger := animax.GetLogger()
	if err != nil {
		logger.Error("failed to add background")
		return err
	}
	logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

func ConcatenateVideos(videos []animax.Video, outputPath string) (err error) {
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
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func ConcatenateVideosFromDir(directoryPath string, outputPath string) error {
	logger := animax.GetLogger()

	dir, err := os.Stat(directoryPath); 
	if os.IsNotExist(err) {
		logger.Errorf("%s does not exist", directoryPath)
		return errors.New("path does not exist")
	}

	if !dir.IsDir() {
		logger.Error("Invalid directoryPath specified")
		return errors.New("invalid directory path")
	}

	filesInDir, err :=os.ReadDir(directoryPath)
	if err != nil {
		logger.Errorf("Could not read files from directory %s", directoryPath)
		return err
	}
	
	videosInDir := []animax.Video{}
	for _, file := range filesInDir {
		videoPath := fmt.Sprintf(`%s/%s`, directoryPath, file.Name())
		extension := filepath.Ext(videoPath)
		if extension == ".mp4" {
			video, err := animax.LoadVideo(videoPath)
			if err != nil {
				logger.Info(fmt.Sprintf(`file: %s is used in concatenation since it is not a video`, filepath.Base(videoPath)))
				continue
			}
			logger.Infof("Appending %s for video concatenation", file.Name())
			videosInDir = append(videosInDir, video)
		}
	}

	err = ConcatenateVideos(videosInDir, outputPath)
	if err != nil {
		logger.Error("Error during concatenation phase")
		return err
	}

	return nil
}

func Skipper(video animax.Video, skipDuration float64, skipInterval float64, outputPath string) error {
	logger := animax.GetLogger()

	nextFrameToSkip := 0.0
	workingDir := uuid.New().String()

	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		os.Mkdir(workingDir, os.ModePerm)
	}
	videoDuration := video.Duration

	logger.Infof("Video: %s | Path: %s | Initiating skipper", video.FileName, video.FilePath)
	for i := 0.0; i < float64(videoDuration); i++ {
		currentFrame := i

		if (currentFrame >= float64(video.Duration)) || (nextFrameToSkip + skipDuration > float64(videoDuration)) || (nextFrameToSkip >= float64(videoDuration)) {break}
		if currentFrame < nextFrameToSkip {continue}

		start := nextFrameToSkip
		end := nextFrameToSkip + skipInterval

		if end >= float64(videoDuration) {end = float64(videoDuration)}

		clipUuid, _ := uuid.NewUUID()
		clipName := fmt.Sprintf(`%s/%s.mp4`, workingDir, clipUuid)
		originalVideo, err := animax.LoadVideo(video.FilePath)
		if err != nil {
			logger.Errorf("Video: %s | Path: %s | Error reading file", originalVideo.FileName, originalVideo.FilePath)
			logger.Infof("Video: %s | Path: %s | Exiting skipper loop", originalVideo.FileName, originalVideo.FilePath)
			break
		}
		originalVideo.Trim(int64(start), int64(end)).Render(clipName, animax.VIDEO_ENCODINGS.Best)
		logger.Infof("Video: %s | Path: %s | Subclip %s generated", video.FileName, video.FilePath, clipName)
		video.Duration -= (int64(end) - int64(start))
		nextFrameToSkip = end + skipDuration

		if nextFrameToSkip >= float64(video.Duration) {
			logger.Infof("Video: %s | Path: %s | Exiting skipper loop", originalVideo.FileName, originalVideo.FilePath)
			break
		}
	}

	logger.Infof("Video: %s | Path: %s | Concatenating all clips in working directory %s", video.FileName, video.FilePath, workingDir)
	err := ConcatenateVideosFromDir(workingDir, outputPath)
	if err != nil {
		logger.Error("Error during concatenation")
		return err
	}
	logger.Infof("Video: %s | Path: %s | Completed concatenation in working directory %s", video.FileName, video.FilePath, workingDir)
	logger.Infof("Video: %s | Path: %s | Cleaning up working directory %s", video.FileName, video.FilePath, workingDir)
	defer os.RemoveAll(workingDir)
	return nil
}

func ExtractAudio(videoPath string, outputPath string) error {
	return nil
}