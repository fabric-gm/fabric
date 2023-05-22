/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didcc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

var didLogger = flogging.MustGetLogger("did")

// Did The struct describes basic details of what makes up a DID
type Did struct {
	Did                  string                 `json:"did"`
	IdType               string                 `json:"idType"`
	IdName               string                 `json:"idName"`
	AdditionalAttributes map[string]interface{} `json:"additionalAttributes"`
	Role                 string                 `json:"role"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Paginator string `json:"paginator"`
	Records   []*Did `json:"records"`
}

type DidCC struct {
}

func New() *DidCC {
	return &DidCC{}
}

func (e *DidCC) Name() string              { return "didcc" }
func (e *DidCC) Chaincode() shim.Chaincode { return e }

func (e *DidCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	didLogger.Info("init Did contract")

	if len(args) != 0 {
		return shim.Error("Incorrect arguments. Not expecting any init arguments")
	}

	return shim.Success(nil)
}

func (e *DidCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	if fn == "CreateDid" {
		if len(args) != 5 {
			return shim.Error("Incorrect arguments number. Expecting 4 args passed")
		}
		uuid, idType, idName, additionalAttributesStr, role := args[0], args[1], args[2], args[3], args[4]
		did, err := CreateDid(stub, uuid, idType, idName, additionalAttributesStr, role)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(did))
	} else if fn == "QueryDid" { // assume 'get' even if fn is nil
		if len(args) != 1 {
			return shim.Error("Incorrect arguments number. Expecting 1 arg passed")
		}
		did := args[0]
		didAsBytes, err := QueryDid(stub, did)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(didAsBytes)
	} else if fn == "QueryDidWithPagination" {
		if len(args) != 2 {
			return shim.Error("Incorrect arguments number. Expecting 2 arg passed")
		}
		paginator, pageSizeStr := args[0], args[1]
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return shim.Error(err.Error())
		}
		didAsBytes, err := QueryDidWithPagination(stub, paginator, int32(pageSize))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(didAsBytes)
	} else {
		return shim.Error(fmt.Errorf("unsupported function %s", fn).Error())
	}
}

// CreateDid adds a new did to the world state with given details
func CreateDid(stub shim.ChaincodeStubInterface,
	uuid string, idType string, idName string, additionalAttributesStr string, role string) (string, error) {

	did := "did:dfi:wmdid:" + uuid
	persistedDid, err := stub.GetState(did)
	if err != nil {
		return "", err
	}

	if persistedDid != nil {
		return "", fmt.Errorf("did %s already exists", did)
	}

	additionalAttributes := make(map[string]interface{})
	err = json.Unmarshal([]byte(additionalAttributesStr), &additionalAttributes)
	if err != nil {
		return "", err
	}

	didStruct := Did{
		Did:                  did,
		IdType:               idType,
		IdName:               idName,
		AdditionalAttributes: additionalAttributes,
		Role:                 role,
	}

	// validate Did
	didAsBytes, _ := json.Marshal(didStruct)

	err = stub.PutState(didStruct.Did, didAsBytes)
	if err != nil {
		return "", err
	}
	return did, nil
}

// QueryDid returns the Did stored in the world state with given id
func QueryDid(stub shim.ChaincodeStubInterface, did string) ([]byte, error) {

	didAsBytes, err := stub.GetState(did)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if didAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", did)
	}

	return didAsBytes, nil
}

func QueryDidWithPagination(stub shim.ChaincodeStubInterface,
	paginator string, pageSize int32) ([]byte, error) {

	resultsIterator, queryResponseMetadata, err := stub.GetStateByRangeWithPagination(
		"", "", pageSize, paginator)

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resultsIterator.Close()
	}()

	var records []*Did

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}
		didStruct := new(Did)
		err = json.Unmarshal(queryResponse.Value, didStruct)

		if err != nil {
			return nil, err
		}

		records = append(records, didStruct)
	}

	queryResult := QueryResult{
		Paginator: queryResponseMetadata.Bookmark,
		Records:   records,
	}
	result, err := json.Marshal(queryResult)

	if err != nil {
		return nil, err
	}

	return result, nil
}
