package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Audio struct {
	FileName string
	FilePath string
	// renders [ ][ ]string
	Duration int64
	Format   string
	args Args
}

const VOLUME_MULTIPLIER_CAP = 100.0

func pullAudioStats(audioPath string) (duration int64) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=duration", audioPath)
	cmd.CombinedOutput()
	// outputLines := strings.Split(string(output), "\n")
	// for _, line := range outputLines {
	// 	fmt.Println(line)
	// }
	return -1
}

func (a Audio) GetType() string {
	return audio
}

func (a Audio) GetFilename() string {
	return a.FileName
}

func (a Audio) GetFilePath() string {
	return a.FilePath
}

func (a Audio) GetExtension() string {
	return filepath.Ext(a.FilePath)
}

func LoadAudio(audioPath string) (audio Audio, err error) {
	file, err := os.Stat(audioPath)
	if err != nil {
		Logger.Error(fmt.Sprintf(`audioPath: %s does not exist`, audioPath))
		return Audio{}, err
	}

	if file.IsDir() {
		Logger.Error(fmt.Sprintf(`videoPath: %s is a directory`, audioPath))
		return Audio{}, errors.New("videoPath: %s is a directory")
	}

	duration := pullAudioStats(audioPath)

	return Audio{
		FileName:    filepath.Base(audioPath),
		FilePath:    audioPath,
		Format:      filepath.Ext(audioPath),
		Duration:    int64(duration),
		// renders:        [ ][ ]string{},
		args: make(Args),
	}, nil
}

func (audio *Audio) Trim(startTime int64, endTime int64) (modifiedAudio *Audio) {
	if startTime > endTime {
		Logger.Error("start time cannot be bigger than end time")
		return &Audio{}
	}
	// audio.renders = append(audio.renders, []string{"-filter_complex", fmt.Sprintf(`[0]trim=start=%d:end=%d[aout]`, startTime, endTime), "-map", "[aout]"})
	//audio.renders = append(audio.renders, []string{"-ss", fmt.Sprintf(`%d`, startTime), "-to", fmt.Sprintf("%d", endTime), "-c:a", "copy"})
	audio.args.addArg("-ss", 
		subArg{
			Key: "ss",
			Value: fmt.Sprintf(`%d -to %d`, startTime, endTime),
		},
	)

	return audio
}

func (audio *Audio) ChangeVolume(multiplier float64) (modifiedAudio *Audio) {
	// audio.renders = append(audio.renders, []string{"-filter:a", fmt.Sprintf(`volume=%f`, multiplier)})
	audio.args.addArg("-filter:a",
		subArg {
			Key: "volume",
			Value: fmt.Sprintf(`volume=%f`, multiplier),
		},
	)
	return audio
}

func (audio *Audio) Nightcore() (modifiedAudio *Audio) {
	// audio.renders = append(audio.renders, []string{"-filter_complex", "asetrate=44100*1.25,atempo=1.25"})
	audio.args.addArg("-filter_complex", 
		subArg{
			Key: "nightcore",
			Value: "asetrate=44100*1.25,atempo=1.25",
		},
	)
	return audio
}

func (audio *Audio) BassBoost() (modifiedAudio *Audio) {
	// audio.renders = append(audio.renders, []string{"-af", "equalizer=f=80:width_type=h:width=50:g=12"})
	audio.args.addArg("-af", 
		subArg{
			Key: "equalizer",
			Value: "equalizer=f=80:width_type=h:width=50:g=12",
		},
	)
	return audio
}

func (audio *Audio) SpeedUp(multiplier float64) (modifiedAudio *Audio) {
	// audio.renders = append(audio.renders, []string{"-filter:a", fmt.Sprintf(`atemp=%f`, multiplier)})
	audio.args.addArg("-filter:a", 
		subArg{
			Key: "speedup",
			Value: fmt.Sprintf(`atemp=%f`, multiplier),
		},
	)
	return audio
}


// func (audio Audio) AddBackgroundImage(imagePath string, outputPath string) {
// 	cmd := exec.Command("ffmpeg", "-loop", "1", "-i", imagePath, "-i", audio.FilePath, "-vf", "crop=in_w*9/16,scale=1920x1080", "-c:v", "libx264", "-c:a", "copy", "-shortest", outputPath)
// 	Logger.Infof(`Adding background`)
// 	combinedOutput, err := cmd.CombinedOutput()
// 	if err != nil {
// 		Logger.Error(string(combinedOutput))
// 	}
// 	Logger.Infof(`Background has been added`)
// }

// func (audio Audio) AddBackgroundVideo(videoPath string, outputPath string) {
// 	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audio.FilePath, "-map", "0:v", "-map", "1:a", "-c:v", "copy", "-y", outputPath)
// 	Logger.Infof(`Adding background`)
// 	combinedOutput, err := cmd.CombinedOutput()
// 	if err != nil {
// 		Logger.Error(string(combinedOutput))
// 	}
// 	Logger.Infof(`Background has been added`)
// }


 
func (audio Audio) Render(outputPath string) (outputAudio Audio) {
	removeIfExists(outputPath)
	// if videoEncoding == "" {videoEncoding = VIDEO_ENCODINGS.Best}

	g := GetRenderGraph(AudioGraph)
	renderStages := g.ProduceOrdering(audio.args, &audio)
	
	if len(renderStages) == 0 {
		Logger.Errorf("No effects applied. Aborting render.\n")
		return audio
	}

	// fmt.Printf("\nALL STAGES %+v\n", renderStages)
	startRender(&renderStages, audio, outputPath)
	outputAudio , err := LoadAudio(outputPath)

	if err != nil {
		return audio
	}
	return outputAudio
}
