package rds

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

// ModifyDTCSecurityIpHostsForSQLServer invokes the rds.ModifyDTCSecurityIpHostsForSQLServer API synchronously
// api document: https://help.aliyun.com/api/rds/modifydtcsecurityiphostsforsqlserver.html
func (client *Client) ModifyDTCSecurityIpHostsForSQLServer(request *ModifyDTCSecurityIpHostsForSQLServerRequest) (response *ModifyDTCSecurityIpHostsForSQLServerResponse, err error) {
	response = CreateModifyDTCSecurityIpHostsForSQLServerResponse()
	err = client.DoAction(request, response)
	return
}

// ModifyDTCSecurityIpHostsForSQLServerWithChan invokes the rds.ModifyDTCSecurityIpHostsForSQLServer API asynchronously
// api document: https://help.aliyun.com/api/rds/modifydtcsecurityiphostsforsqlserver.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyDTCSecurityIpHostsForSQLServerWithChan(request *ModifyDTCSecurityIpHostsForSQLServerRequest) (<-chan *ModifyDTCSecurityIpHostsForSQLServerResponse, <-chan error) {
	responseChan := make(chan *ModifyDTCSecurityIpHostsForSQLServerResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ModifyDTCSecurityIpHostsForSQLServer(request)
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

// ModifyDTCSecurityIpHostsForSQLServerWithCallback invokes the rds.ModifyDTCSecurityIpHostsForSQLServer API asynchronously
// api document: https://help.aliyun.com/api/rds/modifydtcsecurityiphostsforsqlserver.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyDTCSecurityIpHostsForSQLServerWithCallback(request *ModifyDTCSecurityIpHostsForSQLServerRequest, callback func(response *ModifyDTCSecurityIpHostsForSQLServerResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ModifyDTCSecurityIpHostsForSQLServerResponse
		var err error
		defer close(result)
		response, err = client.ModifyDTCSecurityIpHostsForSQLServer(request)
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

// ModifyDTCSecurityIpHostsForSQLServerRequest is the request struct for api ModifyDTCSecurityIpHostsForSQLServer
type ModifyDTCSecurityIpHostsForSQLServerRequest struct {
	*requests.RpcRequest
	SecurityToken        string           `position:"Query" name:"SecurityToken"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	DBInstanceId         string           `position:"Query" name:"DBInstanceId"`
	SecurityIpHosts      string           `position:"Query" name:"SecurityIpHosts"`
	WhiteListGroupName   string           `position:"Query" name:"WhiteListGroupName"`
}

// ModifyDTCSecurityIpHostsForSQLServerResponse is the response struct for api ModifyDTCSecurityIpHostsForSQLServer
type ModifyDTCSecurityIpHostsForSQLServerResponse struct {
	*responses.BaseResponse
	RequestId    string `json:"RequestId" xml:"RequestId"`
	DBInstanceId string `json:"DBInstanceId" xml:"DBInstanceId"`
	DTCSetResult string `json:"DTCSetResult" xml:"DTCSetResult"`
	TaskId       string `json:"TaskId" xml:"TaskId"`
}

// CreateModifyDTCSecurityIpHostsForSQLServerRequest creates a request to invoke ModifyDTCSecurityIpHostsForSQLServer API
func CreateModifyDTCSecurityIpHostsForSQLServerRequest() (request *ModifyDTCSecurityIpHostsForSQLServerRequest) {
	request = &ModifyDTCSecurityIpHostsForSQLServerRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Rds", "2014-08-15", "ModifyDTCSecurityIpHostsForSQLServer", "rds", "openAPI")
	return
}

// CreateModifyDTCSecurityIpHostsForSQLServerResponse creates a response to parse from ModifyDTCSecurityIpHostsForSQLServer response
func CreateModifyDTCSecurityIpHostsForSQLServerResponse() (response *ModifyDTCSecurityIpHostsForSQLServerResponse) {
	response = &ModifyDTCSecurityIpHostsForSQLServerResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
