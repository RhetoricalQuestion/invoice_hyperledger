/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	//"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type Invoice struct {
	InvoiceNumber   string  `json:"invoiceNumber"`
	BilledTo        string  `json:"billedTo"`
	InvoiceDate     string  `json:"invoiceDate"`
	InvoiceAmount   float64 `json:"invoiceAmount"`
	ItemDescription string  `json:"itemDescription"`
	Gr              bool    `json:"gr"`
	IsPaid          bool    `json:"isPaid"`
	PaidAmount      float64 `json:"paidAmount"`
	IsRepaid        bool    `json:"isRepaid"`
	RepaymentAmount float64 `json:"repaymentAmount"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createInvoice" {
		return s.createInvoice(APIstub, args)
	} else if function == "displayAllInvoices" {
		return s.displayAllInvoices(APIstub)
	} else if function == "isGoodsReceived" {
		return s.isGoodsReceived(APIstub, args)
	} else if function == "isPaidToSupplier" {
		return s.isPaidToSupplier(APIstub, args)
	} else if function == "isRepaidToBank" {
		return s.isRepaidToBank(APIstub, args)
	} else if function == "getAuditHistoryForInvoice" {
		return s.getAuditHistoryForInvoice(APIstub, args)
	} /*else if function == "getUser" {
		return s.getUser(APIstub, args)
	} else if function == "raiseInvoiceWithJsonInput" {
		return s.raiseInvoiceWithJsonInput(APIstub, args)
	} */

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	invoices := []Invoice{
		Invoice{
			InvoiceNumber:   "1000",
			BilledTo:        "Mom",
			InvoiceDate:     "13JULY1997",
			InvoiceAmount:   1337.00,
			ItemDescription: "Baby",
			Gr:              false,
			IsPaid:          false,
			PaidAmount:      0.00,
			IsRepaid:        false,
			RepaymentAmount: 0.00},
	}

	i := 0
	for i < len(invoices) {
		fmt.Println("i is ", i)
		invoicesAsBytes, _ := json.Marshal(invoices[i])
		APIstub.PutState("INVOICE"+strconv.Itoa(i), invoicesAsBytes)
		fmt.Println("Added", invoices[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createInvoice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Requires 6.")
	}

	invoAmount, _ := strconv.ParseFloat(args[4], 64)

	var invoice = Invoice{
		InvoiceNumber:   args[1],
		BilledTo:        args[2],
		InvoiceDate:     args[3],
		InvoiceAmount:   invoAmount,
		ItemDescription: args[5],
		Gr:              false,
		IsPaid:          false,
		PaidAmount:      0.00,
		IsRepaid:        false,
		RepaymentAmount: 0.00}

	invoiceAsBytes, _ := json.Marshal(invoice)
	APIstub.PutState(args[0], invoiceAsBytes)
	return shim.Success(nil)
}

func (s *SmartContract) displayAllInvoices(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "INVOICE0"
	endKey := "INVOICE999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- displayAllInvoices:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) isGoodsReceived(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Requires 1.")
	}

	invoice := Invoice{}

	invoiceAsBytes, _ := APIstub.GetState(args[0])

	json.Unmarshal(invoiceAsBytes, &invoice)
	invoice.Gr = true

	invoiceAsBytes, _ = json.Marshal(invoice)
	APIstub.PutState(args[0], invoiceAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) isPaidToSupplier(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Requires 2.")
	}

	invoice := Invoice{}

	invoiceAsBytes, _ := APIstub.GetState(args[0])
	amountPaid, _ := strconv.ParseFloat(args[1], 64)

	json.Unmarshal(invoiceAsBytes, &invoice)

	if amountPaid < invoice.InvoiceAmount {
		invoice.PaidAmount = amountPaid
		invoice.IsPaid = true
	} else {
		return shim.Error("MALI KA, ULIT")
	}

	invoiceAsBytes, _ = json.Marshal(invoice)
	APIstub.PutState(args[0], invoiceAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) isRepaidToBank(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Requires 2.")
	}

	invoice := Invoice{}

	invoiceAsBytes, _ := APIstub.GetState(args[0])
	amountRepaid, _ := strconv.ParseFloat(args[1], 64)

	json.Unmarshal(invoiceAsBytes, &invoice)

	if invoice.InvoiceAmount > amountRepaid {
		invoice.RepaymentAmount = amountRepaid
		invoice.IsRepaid = true
	} else {
		return shim.Error("MALI KA, ULIT")
	}

	invoiceAsBytes, _ = json.Marshal(invoice)
	APIstub.PutState(args[0], invoiceAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

func (s *SmartContract) getAuditHistoryForInvoice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	invoiceKey := args[0]

	resultsIterator, err := APIstub.GetHistoryForKey(invoiceKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		buffer.WriteString(string(response.Value))

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}
