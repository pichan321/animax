package animax

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pichan321/animax"
)

type PageUpload struct {
	FilePath string
	Title string
	Description string
	Token string
	PageId string
}

const mb = 1024 * 1024 // 1 MB
const max_reel_size = 250 * 1024 * 1024 //max reel size allowed is 250 MB
const max_video_size = 10000 * 1024 * 1024 //max video size allowed is 10 GB

func previewBytes(file *os.File) bool {
	buffer := make([]byte, 5 * mb)
	n, err := file.Read(buffer)
	if n == 0 {
		return false
	}
	if err != nil {
		return false
	}
	file.Seek(0, 0)
	return true
}

func UploadToFacebookReelPage(upload PageUpload) error {
	if !(len(upload.Token) > 0) {
		return errors.New("token length must be at least 1 character")
	}

	fileInfo, err := os.Stat(upload.FilePath)
	if err != nil {
		return errors.New("invalid video path")
	}

	if fileInfo.IsDir() {
		return errors.New("path must be a file, not a directory")
	}

	client := &http.Client{}
	requestUrl, err := url.Parse(fmt.Sprintf(`https://graph.facebook.com/v13.0/%s/video_reels`, upload.PageId))
	if err != nil {
		return err
	}

	params := requestUrl.Query()
	params.Add("access_token", upload.Token)
	params.Add("upload_phase", "start")
	requestUrl.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", requestUrl.String(), nil)
	if err != nil {return err}
	response, err := client.Do(req)
	if err != nil {return err}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {return err}

	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {return err}

	uploadUrl := bodyMap["upload_url"].(string)
	videoId := bodyMap["video_id"].(string)
	file, err :=os.Open(upload.FilePath)
	if err != nil {
		return errors.New("unable to open file")
	}
	if fileInfo.Size() > max_reel_size {
		return errors.New(fmt.Sprintf(`max allowed file size is %d`, max_reel_size))
	}
	ok := previewBytes(file)
	if !ok {
		return errors.New("unable to read file content")
	}

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.New("unable to read file content")
	} 
	req, err = http.NewRequest("POST", uploadUrl, bytes.NewBuffer(fileContent))
	req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
	req.Header.Set("offset", "0")
	req.Header.Set("file_size", fmt.Sprintf(`%d`, fileInfo.Size()))
	if err != nil {
		return err
	}
	response, err = client.Do(req)
	if err != nil {
		return err
	}
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	bodyMap = make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	requestUrl, err = url.Parse(fmt.Sprintf(`https://graph.facebook.com/v13.0/%s/video_reels`, upload.PageId))
	if err != nil {
		return err
	}
	params = requestUrl.Query()
	params.Add("access_token", upload.Token)
	params.Add("video_id", videoId)
	params.Add("upload_phase", "finish")
	params.Add("video_state", "PUBLISHED")
	params.Add("description", upload.Description)
	requestUrl.RawQuery = params.Encode()
	req, err = http.NewRequest("POST", requestUrl.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
	if err != nil {
		return err
	}
	response, err = client.Do(req)
	if err != nil {
		return err
	}
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	bodyMap = make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	defer file.Close()
	defer response.Body.Close()
	return nil
}

func UploadToFacebookVideoPage(upload PageUpload) error {
	if !(len(upload.Token) > 0) {
		return errors.New("token length must be at least 1 character")
	}

	fileInfo, err := os.Stat(upload.FilePath)
	if err != nil {
		return errors.New("invalid video path")
	}

	if fileInfo.IsDir() {
		return errors.New("path must be a file, not a directory")
	}

	client := &http.Client{}
	requestUrl, err := url.Parse(fmt.Sprintf(`https://graph-video.facebook.com/v17.0/%s/videos`, upload.PageId))
	if err != nil {
		return err
	}
	params := requestUrl.Query()
	params.Add("access_token", upload.Token)
	params.Add("upload_phase", "start")
	params.Add("file_size", fmt.Sprintf(`%d`, fileInfo.Size()))
	requestUrl.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", requestUrl.String(), nil)
	if err != nil {return err}
	response, err := client.Do(req)
	if err != nil {return err}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	sessionId := bodyMap["upload_session_id"].(string)
	var startOffset int64 = 0
	endOffsetStr := bodyMap["end_offset"].(string)
	endOffset, err := strconv.ParseInt(endOffsetStr, 10, 64)
	if err != nil {
		return err
	}

	file, err :=os.Open(upload.FilePath)
	if err != nil {
		return errors.New("unable to open file")
	}
	ok := previewBytes(file)
	if !ok {
		return errors.New("unable to read file content")
	}
	background := context.Background()
	deadline, cancel := context.WithDeadline(background, time.Now().Add(time.Hour * 1))
	for {
		select {
			case <-deadline.Done():
				cancel()
				return errors.New("upload expired")
			default: 
			animax.Logger.Infof("Uploading in progress: %f%", float64(startOffset)*100/float64(fileInfo.Size()))
		}

		buffer := make([]byte, endOffset - startOffset)
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			cancel()
			return err
		}

		bodyMulti := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyMulti)

		videoPart, err := writer.CreateFormFile("video_file_chunk", "video_chunk.mp4")
		if err != nil {
			fmt.Printf("Failed to create form file: %v\n", err)

		}
		n, err = videoPart.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error writing to multipart")
		}


		err = writer.WriteField("upload_phase", "transfer")
		if err != nil {
			fmt.Println(err.Error())
		}
		err = writer.WriteField("access_token", upload.Token)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = writer.WriteField("upload_session_id", sessionId)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = writer.WriteField("start_offset", fmt.Sprintf("%d", startOffset))
		if err != nil {
			fmt.Println(err.Error())
		}

		// Close the multipart writer
		err = writer.Close()
		if err != nil {
			fmt.Printf("Failed to close multipart writer: %v\n", err)

		}



		uploadUrl, err := url.Parse(fmt.Sprintf(`https://graph-video.facebook.com/v17.0/%s/videos`, upload.PageId))
		if err != nil {
			return err
		}

		req, err = http.NewRequest("POST", uploadUrl.String(), bodyMulti)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))

		if err != nil {
			return err
		}
		response, err = client.Do(req)
		if err != nil {
			return err
		}
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		bodyMap = make(map[string]interface{})
		err = json.Unmarshal(body, &bodyMap)
		if err != nil {
			return err
		}
		startOffset, err = strconv.ParseInt(bodyMap["start_offset"].(string), 10, 64) 
		if err != nil {
			return err
		}
		endOffset, err = strconv.ParseInt(bodyMap["end_offset"].(string), 10, 64) 
		if err != nil {
			return err
		}

		if (startOffset >= endOffset) {
			break
		}
	}
	
	
	publishUrl, err := url.Parse(fmt.Sprintf(`https://graph-video.facebook.com/v17.0/%s/videos`, upload.PageId))
	if err != nil {
		return err
	}
	params = publishUrl.Query()
	params.Add("upload_phase", "finish")
	params.Add("access_token", upload.Token)
	params.Add("upload_session_id", sessionId)
	params.Add("title", upload.Title)
	params.Add("description", upload.Description)
	publishUrl.RawQuery = params.Encode()

	req, err = http.NewRequest("POST", publishUrl.String(), nil)
	if err != nil {
		return err
	}
	response, err = client.Do(req)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	defer file.Close()
	defer cancel()
	defer response.Body.Close()
	return nil
}