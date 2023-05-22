/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didcc

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
)

//func checkInit(t *testing.T, stub *shimtest.MockStub, args [][]byte) {
//	res := stub.MockInit("1", args)
//	if res.Status != shim.OK {
//		fmt.Println("Init failed", string(res.Message))
//		t.FailNow()
//	}
//}
//
//func checkState(t *testing.T, stub *shimtest.MockStub, name string, value string) {
//	bytes := stub.State[name]
//	if bytes == nil {
//		fmt.Println("State", name, "failed to get value")
//		t.FailNow()
//	}
//	if string(bytes) != value {
//		fmt.Println("State value", name, "was not", value, "as expected")
//		t.FailNow()
//	}
//}

func checkQuery(t *testing.T, stub *shimtest.MockStub, queryFunction string, name string, value string) {
	res := stub.MockInvoke("1", [][]byte{[]byte(queryFunction), []byte(name)})

	if res.Status != shim.OK {
		fmt.Println("Query", name, "failed", res.Message)
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("Query", name, "failed to get value")
		t.FailNow()
	}
	if string(res.Payload) != value {
		fmt.Println("Query value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shimtest.MockStub, args [][]byte) string {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", res.Message)
		t.FailNow()
	}
	return string(res.Payload)
}

func TestDidcc_Query(t *testing.T) {
	cc := new(DidCC)
	stub := shimtest.NewMockStub("did", cc)

	// Invoke  set
	did := checkInvoke(t, stub,
		[][]byte{[]byte("CreateDid"),
			[]byte("did1"),
			[]byte("Company"),
			[]byte("King"),
			[]byte("{\"a\":1,\"b\":\"b\"}"),
			[]byte("admin")})

	// Query by hash
	checkQuery(t, stub,
		"QueryDid",
		did,
		"{\"did\":\""+
			did+
			"\",\"idType\":\"Company\",\"idName\":\"King\","+
			"\"additionalAttributes\":"+
			"{\"a\":1,\"b\":\"b\"},\"Role\":\"admin\"}")
}
