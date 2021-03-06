package emr

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// ListScalingActivity invokes the emr.ListScalingActivity API synchronously
// api document: https://help.aliyun.com/api/emr/listscalingactivity.html
func (client *Client) ListScalingActivity(request *ListScalingActivityRequest) (response *ListScalingActivityResponse, err error) {
	response = CreateListScalingActivityResponse()
	err = client.DoAction(request, response)
	return
}

// ListScalingActivityWithChan invokes the emr.ListScalingActivity API asynchronously
// api document: https://help.aliyun.com/api/emr/listscalingactivity.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListScalingActivityWithChan(request *ListScalingActivityRequest) (<-chan *ListScalingActivityResponse, <-chan error) {
	responseChan := make(chan *ListScalingActivityResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListScalingActivity(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// ListScalingActivityWithCallback invokes the emr.ListScalingActivity API asynchronously
// api document: https://help.aliyun.com/api/emr/listscalingactivity.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListScalingActivityWithCallback(request *ListScalingActivityRequest, callback func(response *ListScalingActivityResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListScalingActivityResponse
		var err error
		defer close(result)
		response, err = client.ListScalingActivity(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// ListScalingActivityRequest is the request struct for api ListScalingActivity
type ListScalingActivityRequest struct {
	*requests.RpcRequest
	ResourceOwnerId requests.Integer `position:"Query" name:"ResourceOwnerId"`
	HostGroupId     string           `position:"Query" name:"HostGroupId"`
	PageSize        requests.Integer `position:"Query" name:"PageSize"`
	ClusterId       string           `position:"Query" name:"ClusterId"`
	PageNumber      requests.Integer `position:"Query" name:"PageNumber"`
}

// ListScalingActivityResponse is the response struct for api ListScalingActivity
type ListScalingActivityResponse struct {
	*responses.BaseResponse
	RequestId           string              `json:"RequestId" xml:"RequestId"`
	PageNumber          int                 `json:"PageNumber" xml:"PageNumber"`
	PageSize            int                 `json:"PageSize" xml:"PageSize"`
	Total               int                 `json:"Total" xml:"Total"`
	ScalingActivityList ScalingActivityList `json:"ScalingActivityList" xml:"ScalingActivityList"`
}

// CreateListScalingActivityRequest creates a request to invoke ListScalingActivity API
func CreateListScalingActivityRequest() (request *ListScalingActivityRequest) {
	request = &ListScalingActivityRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Emr", "2016-04-08", "ListScalingActivity", "emr", "openAPI")
	return
}

// CreateListScalingActivityResponse creates a response to parse from ListScalingActivity response
func CreateListScalingActivityResponse() (response *ListScalingActivityResponse) {
	response = &ListScalingActivityResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
