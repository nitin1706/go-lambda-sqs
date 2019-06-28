package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"io/ioutil"
	"net/http"
)


func handle(ctx context.Context, sqsEvent events.SQSEvent) (string, error) {
	msgData := ReadEventMessage(sqsEvent)
	urlStr := fmt.Sprintf("%v", msgData["targetUrl"])
	restCallStatus, _, _ := GetData(urlStr)

	targetTopicArn := fmt.Sprintf("%v", msgData["topicArn"])

	fmt.Println("targetTopicArn", targetTopicArn)
	snsPublishResp, _ := SnsPublish(targetTopicArn, "Status calling : " + urlStr, "Call status of `" +urlStr + "` : "+ restCallStatus)

	return snsPublishResp, nil
}

func main() {
	lambda.Start(handle)
}

func ReadEventMessage(sqsEvent events.SQSEvent) map[string]interface{} {
	msgBodyMap := make(map[string]interface{})
	var msgAttributes = make(map[string]events.SQSMessageAttribute)

	for _, message := range sqsEvent.Records {
		msgAttributes = message.MessageAttributes

		msgBodyBytes := []byte(message.Body)
		err := json.Unmarshal(msgBodyBytes, &msgBodyMap)

		if err != nil {
			fmt.Println("Error in unmarshalling SQS message body.", err)
		}

		msgBodyMap["msgId"] = message.MessageId
		msgBodyMap["eventSource"] = message.EventSource
	}
	fmt.Println("msgAttributes:", msgAttributes)
	return msgBodyMap
}

func GetData(qUrl string) (respStatus string, respBody string, errorDetails error) {
	fmt.Println("GetData", qUrl)
	return GetDataWithHeaders(qUrl, nil)
}

func GetDataWithHeaders(qUrl string, headersMap map[string]string) (respStatus string, respBody string, errorDetails error) {
	fmt.Println("GetDataWithHeaders", qUrl, headersMap)

	client := &http.Client{}
	req, reqErr := http.NewRequest("GET", qUrl, nil)

	if reqErr != nil {
		return "501", "Error in creating GET request for url " + qUrl, reqErr
	}

	for headerKey, headerVal := range headersMap {
		req.Header.Add(headerKey, headerVal)
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		return "501", "Error in GET call for url " + qUrl, respErr
	}

	respStatus = resp.Status

	responseBody, respReadErr := ioutil.ReadAll(resp.Body)
	if respReadErr != nil {
		return respStatus, "Error in reading GET call response from url " + qUrl, respReadErr
	}

	fmt.Println("Get call response : " + string(responseBody))
	return respStatus, string(responseBody), nil
}

func SnsPublish(topicArn string, subjectStr string, msgStr string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sns.New(sess)

	publishResp, publishErr := svc.Publish(&sns.PublishInput{
		TopicArn: &topicArn,
		Subject: aws.String(subjectStr),
		Message: aws.String(msgStr),
	})

	if publishErr != nil {
		fmt.Println("Error in publishing to SNS", publishErr.Error())
		return "Publish Error", publishErr
	}

	return publishResp.String(), nil
}
