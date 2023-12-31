package animax

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"

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
	body, err := io.ReadAll(response.Body)
	if err != nil {return err}

	bodyMap := make(map[string]interface{})
	
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {return err}
	if _, ok := bodyMap["upload_url"]; !ok {
		return errors.New("invalid body response from Facebook")
	}
	if _, ok := bodyMap["video_id"]; !ok {
		return errors.New("invalid body response from Facebook")
	}

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

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return errors.New("unable to read file content")
	} 
	req, err = http.NewRequest("POST", uploadUrl, bytes.NewBuffer(fileContent))
	req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
	req.Header.Set("offset", "0")
	req.Header.Set("file_size", fmt.Sprintf(`%d`, fileInfo.Size()))
	if err != nil {return err}

	response, err = client.Do(req)
	if err != nil {return err}

	body, err = io.ReadAll(response.Body)
	fmt.Println("Body" + string(body))
	fmt.Printf("Body length %d\n",  len(body))
	if err != nil {return err}

	bodyMap = make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {return err}
	if bodyMap == nil {return errors.New("invalid body response from Facebook")}

	requestUrl, err = url.Parse(fmt.Sprintf(`https://graph.facebook.com/v13.0/%s/video_reels`, upload.PageId))
	if err != nil {return err}

	params = requestUrl.Query()
	params.Add("access_token", upload.Token)
	params.Add("video_id", videoId)
	params.Add("upload_phase", "finish")
	params.Add("video_state", "PUBLISHED")
	params.Add("description", upload.Description)
	requestUrl.RawQuery = params.Encode()
	req, err = http.NewRequest("POST", requestUrl.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
	if err != nil {return err}

	response, err = client.Do(req)
	if err != nil {return err}

	body, err = io.ReadAll(response.Body)
	fmt.Println("Body" + string(body))
	fmt.Printf("Body length %d\n",  len(body))
	if err != nil {return err}

	bodyMap = make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {return err}
	if bodyMap == nil {return errors.New("invalid body response from Facebook")}

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

	if fileInfo.Size() > max_video_size {
		return errors.New(fmt.Sprintf(`max allowed file size is %d`, max_video_size))
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}
	if bodyMap == nil {return errors.New("invalid body response from Facebook")}
	if _, ok := bodyMap["upload_session_id"]; !ok {
		return errors.New("invalid body response from Facebook")
	}
	if _, ok := bodyMap["end_offset"]; !ok {
		return errors.New("invalid body response from Facebook")
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
	// background := context.Background()
	// deadline, cancel := context.WithDeadline(background, time.Now().Add(time.Hour * 1))
	for {
		buffer := make([]byte, endOffset - startOffset)
		n, err := file.Read(buffer)
		animax.Logger.Errorf("Reading %d bytes", n)
		if n == 0 {break}
		if err != nil {
			if err == io.EOF {
				break
			}
			// cancel()
			return err
		}
	
		bodyMulti := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyMulti)

		videoPart, err := writer.CreateFormFile("video_file_chunk", "video_chunk.mp4")
		if err != nil {
			fmt.Printf("Failed to create form file: %v\n", err)

		}
		nCopy, err := io.Copy(videoPart, bytes.NewBuffer(buffer[:n]))
		animax.Logger.Errorf("Writing %d bytes", n)
		if err != nil || nCopy == 0 {
			animax.Logger.Error("Error writing to multipart")
		}


		err = writer.WriteField("upload_phase", "transfer")
		if err != nil {
			animax.Logger.Error(err.Error())
		}
		err = writer.WriteField("access_token", upload.Token)
		if err != nil {
			animax.Logger.Error(err.Error())
		}
		err = writer.WriteField("upload_session_id", sessionId)
		if err != nil {
			animax.Logger.Error(err.Error())
		}
		err = writer.WriteField("start_offset", fmt.Sprintf("%d", startOffset))
		if err != nil {
			animax.Logger.Error(err.Error())
		}

		// Close the multipart writer
		err = writer.Close()
		if err != nil {
			animax.Logger.Errorf("Failed to close multipart writer: %v\n", err)
			return err
		}

		uploadUrl, err := url.Parse(fmt.Sprintf(`https://graph-video.facebook.com/v17.0/%s/videos`, upload.PageId))
		if err != nil {return err}

		req, err = http.NewRequest("POST", uploadUrl.String(), bodyMulti)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))

		if err != nil {return err}

		response, err = client.Do(req)
		if err != nil {return err}

		body, err = io.ReadAll(response.Body)
		if err != nil {return err}

		bodyMap = make(map[string]interface{})
		err = json.Unmarshal(body, &bodyMap)
		if err != nil {return err}
		if bodyMap == nil {return errors.New("invalid body response from Facebook")}
		if _, ok := bodyMap["start_offset"]; !ok {
			return errors.New("invalid body response from Facebook")
		}
		if _, ok := bodyMap["end_offset"]; !ok {
			return errors.New("invalid body response from Facebook")
		}
	
		startOffset, err = strconv.ParseInt(bodyMap["start_offset"].(string), 10, 64) 
		if err != nil {return err}

		endOffset, err = strconv.ParseInt(bodyMap["end_offset"].(string), 10, 64) 
		if err != nil {return err}

		animax.Logger.Warnf("Start: %d, End: %d", startOffset, endOffset)
		// select {
		// case <-deadline.Done():
		// 	cancel()
		// 	return errors.New("upload expired")
		// default: 
		// 	animax.Logger.Infof("Uploading in progress: %f %%", float64(startOffset)*100/float64(fileInfo.Size()))
		// }

		animax.Logger.Warn("Loop running")
		if startOffset >= endOffset || startOffset >= fileInfo.Size() {
			break
		}
	}
	
	animax.Logger.Warn("Loop exited")
	publishUrl, err := url.Parse(fmt.Sprintf(`https://graph-video.facebook.com/v17.0/%s/videos`, upload.PageId))
	if err != nil {
		// cancel()
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
		// cancel()
		return err
	}

	response, err = client.Do(req)
	if err != nil {
		// cancel()
		return err
	}

	body, err = io.ReadAll(response.Body)
	if err != nil {
		// cancel()
		return err
	}

	animax.Logger.Infof("Upload published: %s", body)

	defer file.Close()
	// defer cancel()
	defer response.Body.Close()
	return nil
}