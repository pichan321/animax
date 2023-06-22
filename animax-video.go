package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Video struct {
	FileName string
	FilePath string
	Fps float64
	Width int64
	Height int64
	Duration int64
	AspectRatio string
	Format string
	args Args
	IsMuted bool
}

type TrimSection struct {
	StartTime int64
	EndTime int64
	OutputName string
}

var VIDEO_ENCODINGS = struct {
	Best string
	Efficient string
	Compressed string
}{
	Best: "libx264",
	Efficient: "libvpx-vp9",
	Compressed: "libaom-av1",
}

var ASPECT_RATIOS = struct {
	Square float32
	Standard float32
	
	Shorts float32 //Youtube Shorts, Facebok Reels, Instagram Reels, TikTok Videos
	Videos float32 //Youtube Videos, Facebok Videos, General Videos
}{
	Square: 1.0,
	Standard: 16.0/9.0,

	Shorts: 9.0/16.0, //Youtube Shorts, Facebok Reels, Instagram Reels, TikTok Videos
	Videos: 16.0/9.0, //Youtube Videos, Facebok Videos, General Videos
}

func (args Args) addArg(key string, value string) {
    args[key] = append(args[key], value)
}

func pullVideoStats(videoPath string) (width int, height int, duration int, aspectRatio string) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height,duration,display_aspect_ratio", "-of", "default=noprint_wrappers=1", videoPath)
	output, _ := cmd.CombinedOutput()
	outputLines := strings.Split(string(output), "\n")

	if len(outputLines) < 5 {
		return -1, -1, -1, ""
	}

	width, _ = strconv.Atoi(strings.TrimSuffix(strings.Split(outputLines[0], "=")[1], "\r"))
	height, _ = strconv.Atoi(strings.TrimSuffix(strings.Split(outputLines[1], "=")[1], "\r"))
	aspectRatio = strings.TrimSuffix(strings.Split(outputLines[2], "=")[1], "\r")
	duration, err := strconv.Atoi(strings.Split(strings.TrimSuffix(strings.Split(outputLines[4], "=")[1], "\r"), ".")[0])
	if err != nil {
		duration, _ = strconv.Atoi(strings.Split(strings.TrimSuffix(strings.Split(outputLines[3], "=")[1], "\r"), ".")[0])
	}

	return width, height, duration, aspectRatio
}

func LoadVideo(videoPath string) (video Video, err error) {
	logger := GetLogger()
	file, err := os.Stat(videoPath)
	if err != nil {
		logger.Error(fmt.Sprintf(`videoPath: %s does not exist`, videoPath))
		return Video{}, err
	}

	if file.IsDir() {
		logger.Error(fmt.Sprintf(`videoPath: %s is a directory`, videoPath))
		return Video{}, errors.New("videoPath: %s is a directory")
	}

	width, height, duration, aspectRatio := pullVideoStats(videoPath)

	return Video{
		FileName:    filepath.Base(videoPath),
		FilePath:    videoPath,
		Format:      filepath.Ext(videoPath),
		Width:       int64(width),
		Height:      int64(height),
		Duration:    int64(duration),
		AspectRatio: aspectRatio,
		args:        make(Args),
	}, nil
}

func (video *Video) Resize(width int64, height int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, width, height))
	return video
}

func (video *Video) ResizeByWidth(width int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, width, -1))
	return video
}

func (video *Video) ResizeByHeight(height int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, -1 , height))
	return video
}

func (video *Video) Trim(startTime int64, endTime int64) (modifiedVideo *Video){
	logger := GetLogger()
	if startTime > endTime {
		logger.Error("start time cannot be bigger than end time")
		return &Video{}
	}
	
	video.args.addArg("-filter_complex", fmt.Sprintf(`trim=start=%d:end=%d`, startTime, endTime))
	return video
}

func (video *Video) MultipleTrim(concatenateAfter bool, trimSections []TrimSection) (modifiedVideo *Video) {
	for _, v := range trimSections {
		fmt.Print(v)
	}
	return video
}

func (video *Video) Crop(width int64, height int64, startingPositions ...int64) (modifiedVideo *Video) {
	var x, y int64 = 0, 0
	for index, value := range startingPositions {
		if index == 0 {x = value}
		if index == 1  {y = value}
	} 
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=%d:%d:%d:%d`, width, height, x, y))
	return modifiedVideo
}

func (video *Video) CropOutTop(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w:in_h-%d:0:out_h`, pixels))
	return video
}

func (video *Video) CropOutBottom(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w:in_h-%d:0:0`, pixels))
	return video
}

func (video *Video) CropOutLeft(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w-%d:in_h:%d:0`, pixels, pixels))
	return video
}

func (video *Video) CropOutRight(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w-%d:in_h:0:0`, pixels))
	return video
}

func (video *Video) Blur(intensity int16) (modifiedVideo *Video) {
	if (intensity < 0 || intensity > 50) {
		logger := GetLogger()
		logger.Warn("Blur intensity should be between 0 and 50")
		return &Video{}
	}
 	video.args.addArg("-filter_complex", fmt.Sprintf(`boxblur=%d`, intensity))
	return video
}

func (video *Video) NewAspectRatio(aspectRatio float32) (modifiedVideo *Video) {
	video.args["-aspect"] = []string{fmt.Sprintf(`%f`,aspectRatio)}
	return video
}

func (video *Video) NewAspectRatioPadAuto(aspectRatio float32) (modifiedVideo *Video) {
	return video
}

func (video *Video) ChangeVolume(multiplier float64) (modifiedVideo *Video) {
	video.args["-filter:a"] = []string{fmt.Sprintf(`volume=%f`, multiplier)}
	return  video
}

func (video *Video) MuteAudio() (modifiedVideo *Video) {
	video.args["-filter:a"] = []string{`volume=0`}
	return video
}

func shouldProcessFilterComplex(filterComplex []string) ([]string, bool) {
	if len(filterComplex) == 1 && strings.HasPrefix(filterComplex[0], "trim=") {
		startTime := strings.Split(strings.Split(filterComplex[0], "start=")[1], ":end=")[0]
		endTime := strings.Split(filterComplex[0], ":end=")[1]
		return []string{"-ss", startTime, "-to", endTime, "-c", "copy"}, false
	}
	return []string{}, true
}

func (video Video) queryBuilder(outputPath string, videoEncoding string) []string {
	query := []string{"ffmpeg", "-i", video.FilePath, }

	// if reflect.TypeOf(v).Name() == "string" {
	// 	// query = append(query, v)
	// }

	if len(video.args["-aspect"]) > 0 {
		query = append(query, []string{"-aspect", video.args["-aspect"][0]}...)
	}

	if len(video.args["-filter:a"]) > 0 {
		query = append(query, []string{"-filter:a", video.args["-filter:a"][0]}...)
	}
	
	tag := ""
	output := []string{}
	newTrimCmd, shouldProcess := shouldProcessFilterComplex(video.args["-filter_complex"])
	if len(video.args["-filter_complex"]) > 0 && shouldProcess {
		query = append(query, "-filter_complex")
		filter := ""
		//create a helper function that returns output
		// current := ""
		for index, val := range video.args["-filter_complex"] {
			if index == 0 && strings.HasPrefix("trim=", val)  {
				tag = uuid.New().String()[0:4]
				filter += fmt.Sprintf(`[0]%s[%s];`, val, tag)
				continue
			}
			if index == 0  {
				tag = uuid.New().String()[0:4]
				filter += fmt.Sprintf(`[0]%s[%s];`, val, tag)
				continue
			}
			filter += fmt.Sprintf(`[%s]`, tag)
			tag = uuid.New().String()[0:4]
			filter += fmt.Sprintf(`%s[%s];`, val, tag)
		}
		query = append(query, filter[0:len(filter) - 1])

		if len(tag) == 0 {
			output = []string{"-map", "[" + tag +"]", "-c", "copy", outputPath}
		} else {
			output = []string{"-map", "[" + tag +"]", "-c:v", videoEncoding, outputPath}
		}
	}

	if !shouldProcess && len(video.args["-filter:a"]) == 0 {
		output = append(newTrimCmd, outputPath)
		query = append(query, output...)
		return query
	}

	if !shouldProcess && len(video.args["-filter:a"]) == 1 {
		newTrimCmd = append(newTrimCmd[0:len(newTrimCmd)-2], []string{"-c:v", VIDEO_ENCODINGS.Best}...)
		output = append(newTrimCmd, outputPath)
		query = append(query, output...)
		return query
	}

	if len(output) == 0 {
		output = []string{"-c:v", VIDEO_ENCODINGS.Best, outputPath}
	}
	logger := GetLogger()
	logger.Info("Final output " + strings.Join(output, " "))

	query = append(query, output...)
	// query = append(query, []string{"-c:v", "copy", "-c:a", "copy", outputPath}...)
	// query = append(query, []string{outputPath}...)
	return query
}

/*
	If there exists a file at the specified outputPath, the file will be overwritten.
*/
func (video Video) Render(outputPath string, videoEncoding string) (outputVideo Video){
	_, err := os.Stat(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	// outputQuery := video.queryBuilder()
	// cmd := exec.Command()
	if videoEncoding == "" {videoEncoding = VIDEO_ENCODINGS.Best}

	query := video.queryBuilder(outputPath, videoEncoding)
	cmd := exec.Command(query[0], query[1:]...)
	logger := GetLogger()
	logger.Info("Command to be executed: " + cmd.String())
	_, err = cmd.Output()
	if err != nil {
		logger.Error(err.Error())
		return Video{}
	}

	video.args = make(Args)
	var fileInterface interface{}
	fileInterface, err = LoadVideo(outputPath)
	outputVideo = fileInterface.(Video)
	if err != nil {
		logger.Error(fmt.Sprintf("outputVideo: %s cannot be loaded.", outputPath))
		return Video{}
	}
	return outputVideo
}
