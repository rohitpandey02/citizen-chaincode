package main

import (
	"errors"
	"fmt"
	//"strconv"
	//"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	//"regexp"
)

var logger = shim.NewLogger("CDChaincode")

const   AUTHORITY      =  "person"
const   INSTITUTE   =  "institute_user"
const   INSTITUTE_ADMIN   =  "institute_admin"
const   RECORD_ADMIN =  "record_admin"

type SimpleChaincode struct {
}

type Address struct {
	AddressLine1  string    `json:"addressline1"`
	AddressLine2  string    `json:"addressline2"`
	Locality  string    `json:"locality"`
	City  string    `json:"city"`
	State  string    `json:"state"`
	AreaCode  string    `json:"areacode"`
}

type Academics struct {
	InstituteName     string  `json:"institutename"`
	InstituteAddress    Address  `json:"instituteaddress"`
	Qualification     string  `json:"qualification"`
	Description    string  `json:"description"`
	Percentage float64 `json:"percentage"`
}

type Citizen struct {
	PersonID     string  `json:"personid"`
	GovtID    string  `json:"govtid"`
	Name     string  `json:"name"`
	DOB    string  `json:"dob"`
	CurrentAddress Address `json:"currentaddress"`
	PersonAcademics    []Academics `json:"personacademics"`
}

type ID_Holder struct {
	IDs 	[]string `json:"ids"`
}

type User_and_eCert struct {
	User string `json:"user"`
	eCert string `json:"ecert"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting at least two.")
	}

	err := stub.PutState("test", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	var PersonIDs ID_Holder

	bytes, err := json.Marshal(PersonIDs)

    if err != nil { return nil, errors.New("Error creating ID_Holder record") }

	err = stub.PutState("PersonIDs", bytes)

	/*for i:=0; i < len(args); i=i+2 {
		t.add_ecert(stub, args[i], args[i+1])
	}*/

	return nil, nil
}

func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }

	return ecert, nil
}

func (t *SimpleChaincode) add_ecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {


	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}

func (t *SimpleChaincode) check_role(stub shim.ChaincodeStubInterface) (string, error) {
    role, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	return string(role), nil

}

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

  role, err := t.check_role(stub);

    if err != nil { return "", "", err }

	return user, role, nil
}

func (t *SimpleChaincode) retrieve_ID(stub shim.ChaincodeStubInterface, PersonID string) (Citizen, error) {

	var c Citizen

	bytes, err := stub.GetState(PersonID);

	if err != nil {	fmt.Printf("RETRIEVE_ID: Failed to invoke citizen_code: %s", err); return c, errors.New("RETRIEVE_ID: Error retrieving person with ID = " + PersonID) }

	err = json.Unmarshal(bytes, &c);

    if err != nil {	fmt.Printf("RETRIEVE_ID: Corrupt citizen record "+string(bytes)+": %s", err); return c, errors.New("RETRIEVE_ID: Corrupt Citizen record"+string(bytes))	}

	return c, nil
}

func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, c Citizen) (bool, error) {

	bytes, err := json.Marshal(c)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting citizen record: %s", err); return false, errors.New("Error converting citizen record") }

	err = stub.PutState(c.PersonID, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing citizen record: %s", err); return false, errors.New("Error storing citizen record") }

	return true, nil
}



func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "ping" {
    return t.ping(stub)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	if function == "read" {
		return t.read(stub, args)
	}else if function == "ping" {
		return t.ping(stub)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Alive!!!"), nil
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]
	value = args[1]
	err = stub.PutState(key, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

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
