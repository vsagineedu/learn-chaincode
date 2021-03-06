/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

////==============================================================================================================================
//	SupplyItem - Defines the structure for a SupplyItem object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//=============================================================================================================================
type SupplyItem struct {
	SupplierID			string `json:"supplierID"`
	OperatorID			string `json:"operatorID"`
  Longitude       string `json:"longitude"`
	Latitude        string `json:"latitude"`
	Description     string `json:"description"`
	MaterialType    string `json:"materialType"`
	MaterialQty     string `json:"materialQuantity"`
	UnitOfMeasure   string `json:"unitOfMeasure"`
	Photo						string `json:"photo"`
	SupplyItemID    string `json:"supplyItemID"`
}

//==============================================================================================================================
//	SupplyItems Holder - Defines the structure that holds all the SupplyItemIDs for SupplyItems that have been created.
//				Used as an index when querying all SupplyItems.
//==============================================================================================================================

type SupplyItemIDs_Holder struct {
	SupplyItemIDs 	[]string `json:"supplyitemids"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	//Args
	//				0
	//			peer_address

  fmt.Println("invoke is running " + function)
	var supplyItemIDs SupplyItemIDs_Holder

	bytes, err := json.Marshal(supplyItemIDs)

  if err != nil { return nil, errors.New("Error creating SupplyItemIDs_Holder record") }

	err = stub.PutState("supplyItemIDs", bytes)

	return nil, nil
}


// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

////=================================================================================================================================
//	 check_unique_supplyItem
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_supplyItem(stub shim.ChaincodeStubInterface, supplyItemID string) ([]byte, error) {
	_, err := t.retrieve_SupplyItem(stub, supplyItemID)
	if err == nil {
		return []byte("false"), errors.New("SupplItem is not unique")
	} else {
		return []byte("true"), nil
	}
}

//==============================================================================================================================
//	 retrieve_supplyItemID - Gets the state of the data at supplyItemID in the ledger then converts it from the stored
//					JSON into the SupplyItem struct for use in the contract. Returns the SupplYItem struct.
//					Returns empty SupplyItem if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_SupplyItem(stub shim.ChaincodeStubInterface, supplyItemID string) (SupplyItem, error) {

	var sItem SupplyItem

	bytes, err := stub.GetState(supplyItemID);

	if err != nil {	fmt.Printf("RETRIEVE_SupplyItem: Failed to invoke supplyitem_id: %s", err); return sItem, errors.New("RETRIEVE_SupplyItem: Error retrieving supplyitem with supplyItemID = " + supplyItemID) }

	err = json.Unmarshal(bytes, &sItem);

    if err != nil {	fmt.Printf("RETRIEVE_SupplyItem: Corrupt supplyItem record "+string(bytes)+": %s", err); return sItem, errors.New("RETRIEVE_SupplyItem: Corrupt supplyItem record"+string(bytes))	}

	return sItem, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the SupplyItem struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, sItem SupplyItem) (bool, error) {

	bytes, err := json.Marshal(sItem)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting supplyitem record: %s", err); return false, errors.New("Error converting supply item record") }

	err = stub.PutState(sItem.SupplyItemID, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing supplyitem record: %s", err); return false, errors.New("Error storing supplyitem record") }

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "create_supplyItem" {
        return t.create_supplyItem(stub, args)
	} else if function == "update_supplyItem" {
		  sItem, err := t.retrieve_SupplyItem(stub, args[0])
 		  if err != nil { fmt.Printf("INVOKE: Error retrieving supplyItemID: %s", err); return nil, errors.New("Error retrieving supplyItem") }
      return t.update_supplyItem(stub, sItem, args[1])
    }
		return nil, errors.New("Function of the name "+ function +" doesn't exist.")

	}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create SupplyItem - Creates the initial JSON for the SupplyItem and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_supplyItem(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var sItem SupplyItem

	supplyItemID   := "\"SupplyItemID\":\""+args[0]+"\", "   // Variables to define the JSON
	supplierID		 := "\"SupplierID\":\""+args[1]+"\", "
	operatorID		 := "\"OperatorID\":\""+args[2]+"\", "
	longitude      := "\"Longitude\":\""+args[3]+"\", "
	latitude       := "\"Latitude\":\""+args[4]+"\", "
	description    := "\"Description\":\""+args[5]+"\", "
	materialType   := "\"MaterialType\":\""+args[6]+"\", "
	materialQty    := "\"MaterialQty\":\""+args[7]+"\", "
	unitOfMeasure  := "\"UnitOfMeasure\":\""+args[8]+"\", "
	photo					 := "\"Photo\":\""+args[9]+"\""

	supplyitem_json := "{"+supplyItemID+supplierID+operatorID+longitude+latitude+description+materialType+materialQty+unitOfMeasure+photo+"}" 	// Concatenates the variables to create the total JSON object


	if 	supplyItemID  == "" {
							fmt.Printf("CREATE_SUPPLYITEM: Invalid supplyItemID provided");
							return nil, errors.New("Invalid supplyItemID provided")
	}

	json.Unmarshal([]byte(supplyitem_json), &sItem)							// Convert the JSON defined above into a SupplyItem object for go

	//if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(sItem.SupplyItemID) 								// If not an error then a record exists so cant create a new supplyitem with this SupplyItemID as it must be unique

																		if record != nil { return nil, errors.New("SupplyItem already exists") }


	_, err  = t.save_changes(stub, sItem)

																		if err != nil { fmt.Printf("CREATE_SUPPLYITEM: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("supplyItemIDs")

																		if err != nil { return nil, errors.New("Unable to get supplyItemIDs") }

	var supplyItemIDsHolder SupplyItemIDs_Holder

	err = json.Unmarshal(bytes, &supplyItemIDsHolder)

																		if err != nil {	return nil, errors.New("Corrupt SupplyItemIDs_Holder record") }

	supplyItemIDsHolder.SupplyItemIDs = append(supplyItemIDsHolder.SupplyItemIDs, args[0])


	bytes, err = json.Marshal(supplyItemIDsHolder)

															if err != nil { fmt.Print("Error creating supplyItemIDsHolder record") }

	err = stub.PutState("supplyItemIDs", bytes)

															if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

//=================================================================================================================================
//	 update_supplyItem
//=================================================================================================================================
func (t *SimpleChaincode) update_supplyItem(stub shim.ChaincodeStubInterface, sItem SupplyItem, new_value string) ([]byte, error) {
	sItem.OperatorID = new_value
	_, err := t.save_changes(stub, sItem)
		if err != nil { fmt.Printf("UPDATE_MAKE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}
