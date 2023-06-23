package animax

import (
	"bytes"
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
)

type PageUpload struct {
	FilePath string
	Title string
	Description string
	Token string
	PageId string
}

const MB = 1024 * 1024

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

	}
	params := requestUrl.Query()
	params.Add("access_token", upload.Token)
	params.Add("upload_phase", "start")
	requestUrl.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", requestUrl.String(), nil)
	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", bodyMap)
	uploadUrl := bodyMap["upload_url"].(string)
	videoId := bodyMap["video_id"].(string)
	file, err :=os.Open(upload.FilePath)
	if err != nil {
		return errors.New("unable to open file")
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
	fmt.Println(string(body))
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
	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
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

	// previewBuffer := make([]byte, 5 * MB)
	// n, err := file.Read(previewBuffer)
	// if err != nil || n <= 0 {
	// 	return errors.New("could not preview video file")
	// }
	// previewBuffer = nil
	// _, err = file.Seek(0, 0)
	// if err != nil {
	// 	return err
	// }

	
	
	// startOffset, _ = strconv.ParseInt(bodyMap["start_offset"].(string), 10, 64)
	// endOffset, _ = strconv.ParseInt(bodyMap["end_offset"].(string), 10, 64)

	for {

		buffer := make([]byte, endOffset - int64(startOffset))
		n, err := file.Read(buffer)
		fmt.Printf(`Start: %d, End: %d, Buffer size: %d`, startOffset, endOffset, n)
		fmt.Println()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		bodyMulti := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyMulti)

		// Add the video chunk as a form field
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

		}
		// params = uploadUrl.Query()
		// params.Add("upload_phase", "transfer")
		// params.Add("access_token", upload.Token)
		// params.Add("upload_session_id", sessionId)
		// params.Add("start_offset", "0")
		// // params.Add("video_file_chunk", string(buffer[:n]))
		// uploadUrl.RawQuery = params.Encode()
		fmt.Println("Number of btytes sending")
		fmt.Println(n)
		fmt.Println(uploadUrl.String())
		req, err = http.NewRequest("POST", uploadUrl.String(), bodyMulti)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
		// req.Header.Set("offset", "0")
		// req.Header.Set("file_size", fmt.Sprintf(`%d`, fileInfo.Size()))
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
	// requestUrl, err = url.Parse(fmt.Sprintf(`https://graph.facebook.com/v13.0/%s/video_reels`, upload.PageId))
	// if err != nil {
	// 	return err
	// }
	// params = requestUrl.Query()
	// params.Add("access_token", upload.Token)
	// params.Add("video_id", videoId)
	// params.Add("upload_phase", "finish")
	// params.Add("video_state", "PUBLISHED")
	// params.Add("description", upload.Description)
	// requestUrl.RawQuery = params.Encode()
	// req, err = http.NewRequest("POST", requestUrl.String(), nil)
	// req.Header.Set("Authorization", fmt.Sprintf(`OAuth %s`, upload.Token))
	// if err != nil {
	// 	return err
	// }
	// response, err = client.Do(req)
	// if err != nil {
	// 	return err
	// }
	// body, err = ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	return err
	// }
	// bodyMap = make(map[string]interface{})
	// err = json.Unmarshal(body, &bodyMap)
	// if err != nil {
	// 	return err
	// }
	
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
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println("Publish status")
	fmt.Println(string(body))

	defer file.Close()
	defer response.Body.Close()
	return nil
}