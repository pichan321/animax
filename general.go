package animax

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Args map[string][]subArg

func secondsToHMS(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func fixSpace(slice *[]string) {
	for i := 0; i < len(*slice); i++ {
		splits := strings.Fields((*slice)[i])
		if len(splits) > 1 {
			*slice = append((*slice)[:i], append(splits[:], (*slice)[i+1:]...)...)
		}
	}
	for i := 0; i < len(*slice); i++ {
		if len(strings.Fields((*slice)[i])) > 1 {
			fixSpace(slice)
			return
		}
	}
}

func isTrim(cmd *[]string) bool {
	for _, v := range *cmd {
		if strings.Contains(v, "-ss") {
			return true
		}
	}
	return false
}

func shouldEncode(cmd *[]string, currentIndex int, renderStages *[][]string) {
	if currentIndex == len(*renderStages)-1 {
		*cmd = append(*cmd, []string{"-c:v", "libx264", "-y"}...)
		return
	}

	next := (*renderStages)[currentIndex+1]

	if isTrim(&next) {
		// fmt.Println("Next is trim")
		fixTrim(cmd)
		*cmd = append(*cmd, []string{"-c", "copy", "-y"}...)
		fmt.Printf("After %+v\n", cmd)
		return
	}

	*cmd = append(*cmd, []string{"-c", "copy", "-y"}...)
	// *cmd = append(*cmd, []string{"-c:v", "libx264"}...)
}

func fixTrim(cmd *[]string) {
	input := (*cmd)[1:3]
	temp := make([]string, len(input))
	copy(temp, input)
	start := (*cmd)[3:5]
	copy((*cmd)[1:3], start)
	copy((*cmd)[3:5], temp)
	*cmd = (*cmd)[:7]

	startTime, err := strconv.ParseFloat((*cmd)[2], 64)
	if err != nil {
		return
	}
	endTime, err := strconv.ParseFloat((*cmd)[6], 64)
	if err != nil {
		return
	}

	(*cmd)[6] = fmt.Sprintf("%.5f", endTime-startTime)
}

func removeFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

func removeIfExists(path string) {
	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}
}

func startRender(renderStages *[][]string, file File, finalOutputPath string) {
	base := []string{"ffmpeg", "-i"}

	workingDir := uuid.New().String()
	os.Mkdir(workingDir, os.ModePerm)
	defer os.RemoveAll(workingDir)
	temp := uuid.New().String()

	inputPath := file.GetFilePath()
	nextPath := fmt.Sprintf("%s/%s%s", workingDir, temp, file.GetExtension())

	for i := 0; i < len(*renderStages); i++ {
		if len((*renderStages)[i]) == 0 {
			continue
		}

		cmd := base
		cmd = append(cmd, inputPath)

		fixSpace(&(*renderStages)[i])
		cmd = append(cmd, (*renderStages)[i]...)

		if isTrim(&cmd) {
			// fixTrim(&cmd)
			shouldEncode(&cmd, i, renderStages)
		} else {
			switch file.GetType() {
			case video:
				cmd = append(cmd, []string{"-c:v", "libx264", "-y"}...)
			case audio:
				cmd = append(cmd, []string{"-y"}...)
			}
		}
		cmd = append(cmd, nextPath)

		execute := exec.Command(cmd[0], cmd[1:]...)

		Logger.Infoln("Command to be executed: " + execute.String())
		output, err := execute.CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
		}
		inputPath = nextPath
		nextPath = fmt.Sprintf("%s/%s.mp4", workingDir, uuid.New().String())
	}

	os.Rename(inputPath, finalOutputPath)

}