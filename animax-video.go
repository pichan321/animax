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

type Args map[string][]string

type File interface {
	Trim()
}

type Video struct {
	FileName string
	Fps float64
	Width int64
	Height int64
	Duration int64
	Volume int32
	Format string
	args Args
	IsMuted bool
}

type Audio struct {
	FileName string
	Duration int64
	args Args
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


func PrintHello() {
	fmt.Println("Hello")
}

func (args Args) addArg(key string, value string) {
    args[key] = append(args[key], value)
}

func pullStats(videoPath string) {
	cmd := exec.Command("ffmpeg", "-i", videoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		
	}
	fmt.Println(string(output))
}


func Load(videoPath string) (video Video, err error) { 
	file, err := os.Stat(videoPath)
	if err != nil {
		logger.Error(fmt.Sprintf(`videoPath: %s does not exist`, videoPath))
		return Video{}, err
	}

	if file.IsDir() {
		logger.Error(fmt.Sprintf(`videoPath: %s is a directory`, videoPath))
		return Video{}, errors.New("videoPath: %s is a directory")
	}

	pullStats(videoPath)

	return Video{
		FileName: filepath.Base(videoPath),
		Format: filepath.Ext(videoPath),
		args: make(Args),
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
	return video
}

func (video *Video) CropOutTop(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w:in_h-%d:0:out_h:`, pixels))
	return video
}

func (video *Video) CropOutBottom(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w:in_h-%d:0:0`, pixels))
	return video
}

func (video *Video) CropOutLeft(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w-%d:in_h:%d:0:`, pixels, pixels))
	return video
}

func (video *Video) CropOutRight(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`crop=in_w-%d:in_h:0:0:`, pixels))
	return video
}

func (video *Video) Blur(intensity int16) (modifiedVideo *Video) {
	if (intensity < 0 || intensity > 50) {
		logger.Warn("Blur intensity should be between 0 and 50")
		return video
	}
 	video.args.addArg("-filter_complex", fmt.Sprintf(`boxblur=%d`, intensity))
	return video
}

func MultiFilter() {
	
}

func (video *Video) Skipper(skipDuration int64, skipInterval int64) {
	
}


func (video *Video) NewAspectRatio(aspectRatio float32) (modifiedVideo *Video) {
	
	return video
}

func (video *Video) ChangeVideoVolume() (modifiedVideo *Video) {
	// video.args.addArg("-filter:a", fmt.Sprintf(`volume=db`))
	return 
}

func (video *Video) MuteAudio() {
	
	return
}




func (video Video) queryBuilder(outputPath string) []string {
	query := []string{"ffmpeg", "-i", video.FileName, "-filter_complex"}

	// if reflect.TypeOf(v).Name() == "string" {
	// 	// query = append(query, v)
	// }
	filter := ""
	current := ""
	for index, val := range video.args["-filter_complex"] {
		if index == 0 {
			current = uuid.New().String()[0:4]
			filter += fmt.Sprintf(`[0]%s[%s];`, val, current)
			continue
		}
		filter += fmt.Sprintf(`[%s]`, current)
		current = uuid.New().String()[0:4]
		filter += fmt.Sprintf(`%s[%s];`, val, current)
	}
	query = append(query, filter[0:len(filter) - 1])
	if len(filter) == 0 {
		query = append(query, []string{"-map", "[" + current +"]", "-c", "copy", outputPath}...)
	}
	query = append(query, []string{"-map", "[" + current +"]", "-c:v", VIDEO_ENCODINGS.Best, outputPath}...)
	// query = append(query, []string{"-c:v", "copy", "-c:a", "copy", outputPath}...)
	// query = append(query, []string{outputPath}...)
	fmt.Println(query)
	return query
}

func (video Video) Render(outputPath string) {
	_, err := os.Stat(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	// outputQuery := video.queryBuilder()
	// cmd := exec.Command()
	query := video.queryBuilder(outputPath)
	cmd := exec.Command(query[0], query[1:]...)
	logger.Info("Command to be executed: " + cmd.String())
	ouput, err := cmd.Output()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Info(string(ouput))
}
