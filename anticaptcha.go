package anticaptcha

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
//	baseURL      = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
	sendInterval = 10 * time.Second
)

stype Client struct {
	APIKeya string
	baseURL string
}

// Method to create the task to process the recaptcha, returns the task_id
func (c *Client) createTaskRecaptcha(websiteURL string, recaptchaKey string) float64 {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":       "NoCaptchaTaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	return responseBody["taskId"].(float64)
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Client) getTaskResult(taskID float64) map[string]interface{} {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody
}

// SendRecaptcha Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendRecaptcha(websiteURL string, recaptchaKey string) string {
	// Create the task on anti-captcha api and get the task_id
	taskID := c.createTaskRecaptcha(websiteURL, recaptchaKey)

	// Check if the result is ready, if not loop until it is
	response := c.getTaskResult(taskID)
	for {
		if response["status"] == "processing" {
			log.Println("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(sendInterval)
			response = c.getTaskResult(taskID)
		} else {
			log.Println("Result is ready.")
			break
		}
	}
	return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string)
}

// Method to create the task to process the image captcha, returns the task_id
func (c *Client) createTaskImage(imgString string) float64 {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	return responseBody["taskId"].(float64)
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendImage(imgString string) string {
	// Create the task on anti-captcha api and get the task_id
	taskID := c.createTaskImage(imgString)

	// Check if the result is ready, if not loop until it is
	response := c.getTaskResult(taskID)
	for {
		if response["status"] == "processing" {
			log.Println("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(sendInterval)
			response = c.getTaskResult(taskID)
		} else {
			log.Println("Result is ready.")
			break
		}
	}
	return response["solution"].(map[string]interface{})["text"].(string)
}
