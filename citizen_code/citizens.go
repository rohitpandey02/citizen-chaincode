package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("CDChaincode")

const AUTHORITY = "person"
const HEALTHCARE_USER = "healthcare_user"
const HEALTHCARE_ADMIN = "healthcare_admin"
const GOVT_ADMIN = "govt_admin"

type SimpleChaincode struct {
}

type Address struct {
	AddressLine1 string `json:"addressline1"`
	AddressLine2 string `json:"addressline2"`
	Locality     string `json:"locality"`
	City         string `json:"city"`
	State        string `json:"state"`
	AreaCode     string `json:"areacode"`
}

type Health struct {
	HealthRecordID     string  `json:"healthrecordid"`
	PhysicianName      string  `json:"physicianname"`
	FacilityName       string  `json:"facilityname"`
	FacilityAddress    Address `json:"facilityaddress"`
	TypeOfService      string  `json:"typeofservice"`
	ServiceDescription string  `json:"servicedescription"`
	DateOfService      string  `json:"dateofservice"`
	DateOfAdmission    string  `json:"dateofadmission"`
	DateOfDischarge    string  `json:"dateofdischarge"`
	DischargeSummary   string  `json:"dischargesummary"`
}

type Citizen struct {
	PersonID       string   `json:"personid"`
	GovtID         string   `json:"govtid"`
	Name           string   `json:"name"`
	Gender         string   `json:"gender"`
	DOB            string   `json:"dob"`
	BloodGroup     string   `json:"bloodgroup"`
	CurrentAddress Address  `json:"currentaddress"`
	PersonHealth   []Health `json:"personahealth"`
}

type ID_Holder struct {
	IDs []string `json:"ids"`
}

type User_and_eCert struct {
	User  string `json:"user"`
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

	var PersonIDs ID_Holder

	bytes, err := json.Marshal(PersonIDs)

	if err != nil {
		return nil, errors.New("Error creating ID_Holder record")
	}

	err = stub.PutState("PersonIDs", bytes)

	for i := 0; i < len(args); i = i + 2 {
		t.add_ecert(stub, args[i], args[i+1])
	}

	return nil, nil
}

func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil {
		return nil, errors.New("Couldn't retrieve ecert for user " + name)
	}

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

	username, err := stub.ReadCertAttribute("username")
	if err != nil {
		return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error())
	}
	return string(username), nil
}

func (t *SimpleChaincode) check_role(stub shim.ChaincodeStubInterface) (string, error) {
	role, err := stub.ReadCertAttribute("role")
	if err != nil {
		return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error())
	}
	return string(role), nil

}

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error) {

	user, err := t.get_username(stub)

	role, err := t.check_role(stub)

	if err != nil {
		return "", "", err
	}

	return user, role, nil
}

func (t *SimpleChaincode) retrieve_ID(stub shim.ChaincodeStubInterface, PersonID string) (Citizen, error) {

	var c Citizen

	bytes, err := stub.GetState(PersonID)

	if err != nil {
		fmt.Printf("RETRIEVE_ID: Failed to invoke citizen_code: %s", err)
		return c, errors.New("RETRIEVE_ID: Error retrieving person with ID = " + PersonID)
	}

	err = json.Unmarshal(bytes, &c)

	if err != nil {
		fmt.Printf("RETRIEVE_ID: Corrupt citizen record "+string(bytes)+": %s", err)
		return c, errors.New("RETRIEVE_ID: Corrupt Citizen record" + string(bytes))
	}

	return c, nil
}

func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, c Citizen) (bool, error) {

	bytes, err := json.Marshal(c)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error converting citizen record: %s", err)
		return false, errors.New("Error converting citizen record")
	}

	err = stub.PutState(c.PersonID, bytes)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error storing citizen record: %s", err)
		return false, errors.New("Error storing citizen record")
	}

	return true, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	caller, caller_role, err := t.get_caller_data(stub)
	if err != nil {
		return nil, errors.New("Error retrieving caller information")
	}

	if function == "create_person" {
		return t.create_person(stub, caller, caller_role, args[0], args[1], args[2])
	} else if function == "ping" {
		return t.ping(stub)
	} else if function == "write" {
		return t.write(stub, args)
	} else {
		c, err := t.retrieve_ID(stub, args[0])
		if err != nil {
			fmt.Printf("INVOKE: Error retrieving ID: %s", err)
			return nil, errors.New("Error retrieving ID")
		}
		if function == "add_govtid" {
			return t.add_govtid(stub, c, caller, caller_role, args[0])
		} else if function == "add_name" {
			return t.add_name(stub, c, caller, caller_role, args[0])
		} else if function == "update_address" {
			return t.update_address(stub, c, caller, caller_role, args[0], args[1], args[2], args[3], args[4], args[5])
		} else if function == "add_healthrecord" {
			return t.add_healthrecord(stub, c, caller, caller_role, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], args[12])
		} else if function == "update_healthrecord" {
			return t.update_healthrecord(stub, c, caller, caller_role, args[0], args[1], args[2])
		}
	}

	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) create_person(stub shim.ChaincodeStubInterface, caller string, caller_role string, PersonID string, DOB string, Gender string) ([]byte, error) {
	var c Citizen

	personid := "\"PersonID\":\"" + PersonID + "\", "
	govtid := "\"GovtID\":\"UNDEFINED\", "
	name := "\"Name\":\"UNDEFINED\", "
	gender := "\"Gender\":\"" + DOB + "\", "
	dob := "\"DOB\":\"" + Gender + "\", "
	bloodgroup := "\"BloodGroup\":\"UNDEFINED\", "
	currentaddress := "\"CurrentAddress\":\"UNDEFINED\", "
	personhealth := "\"PersonHealth\":[]"

	citizen_json := "{" + personid + govtid + name + gender + dob + bloodgroup + currentaddress + personhealth + "}"

	if PersonID == "" {
		fmt.Printf("CREATE_PERSON: Invalid PersonID provided")
		return nil, errors.New("Invalid PersonID provided")
	}

	err := json.Unmarshal([]byte(citizen_json), &c)

	if err != nil {
		return nil, errors.New("Invalid JSON object")
	}

	record, err := stub.GetState(c.PersonID)

	if record != nil {
		return nil, errors.New("Citizen already exists")
	}

	if caller_role != GOVT_ADMIN {

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_person. %v === %v", caller_role, AUTHORITY))

	}

	_, err = t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("CREATE_PERSON: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	bytes, err := stub.GetState("PersonIDs")

	if err != nil {
		return nil, errors.New("Unable to get PersonIDs")
	}

	var PersonIDs ID_Holder

	err = json.Unmarshal(bytes, &PersonIDs)

	if err != nil {
		return nil, errors.New("Corrupt ID_Holder record")
	}

	PersonIDs.IDs = append(PersonIDs.IDs, PersonID)

	bytes, err = json.Marshal(PersonIDs)

	if err != nil {
		fmt.Print("Error creating ID_Holder record")
	}

	err = stub.PutState("PersonIDs", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

func (t *SimpleChaincode) add_govtid(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, GovtID string) ([]byte, error) {

	if caller_role == GOVT_ADMIN {

		c.GovtID = GovtID

	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. add_govtid"))
	}

	_, err := t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("ADD_GOVTID: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) add_name(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, Name string) ([]byte, error) {

	if caller_role == GOVT_ADMIN || caller_role == AUTHORITY {

		c.Name = Name

	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. add_name"))
	}

	_, err := t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("ADD_NAME: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) add_bloodgroup(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, BloodGroup string) ([]byte, error) {

	if caller_role == HEALTHCARE_ADMIN || caller_role == AUTHORITY {

		c.BloodGroup = BloodGroup

	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. add_bloodgroup"))
	}

	_, err := t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("ADD_BLOODGROUP: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) update_address(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, AddressLine1 string, AddressLine2 string, Locality string, City string, State string, AreaCode string) ([]byte, error) {

	var address Address

	addressline1 := "\"AddressLine1\":\"" + AddressLine1 + "\", "
	addressline2 := "\"AddressLine2\":\"" + AddressLine2 + "\", "
	locality := "\"Locality\":\"" + Locality + "\", "
	city := "\"City\":\"" + City + "\", "
	state := "\"State\":\"" + State + "\", "
	areacode := "\"AreaCode\":\"" + AreaCode + "\""

	address_json := "{" + addressline1 + addressline2 + locality + city + state + areacode + "}"

	err := json.Unmarshal([]byte(address_json), &address)

	if caller_role == GOVT_ADMIN || caller_role == AUTHORITY {

		c.CurrentAddress = address

	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. update_address"))
	}

	_, err = t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("UPDATE_ADDRESS: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) add_healthrecord(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, HealthRecordID string, PhysicianName string, FacilityName string, AddressLine1 string, AddressLine2 string, Locality string, City string, State string, AreaCode string, TypeOfService string, ServiceDescription string, DateOfService string, DateOfAdmission string) ([]byte, error) {

	var healthrecord Health

	addressline1 := "\"AddressLine1\":\"" + AddressLine1 + "\", "
	addressline2 := "\"AddressLine2\":\"" + AddressLine2 + "\", "
	locality := "\"Locality\":\"" + Locality + "\", "
	city := "\"City\":\"" + City + "\", "
	state := "\"State\":\"" + State + "\", "
	areacode := "\"AreaCode\":\"" + AreaCode + "\""

	address_json := "{" + addressline1 + addressline2 + locality + city + state + areacode + "}"

	healthrecordid := "\"HealthRecordID\":\"" + HealthRecordID + "\", "
	physicianname := "\"PhysicianName\":\"" + PhysicianName + "\", "
	facilityname := "\"FacilityName\":\"" + FacilityName + "\", "
	facilityaddress := "\"FacilityAddress\":\"" + address_json + "\", "
	typeofservice := "\"TypeOfService\":\"" + TypeOfService + "\", "
	servicedescription := "\"ServiceDescription\":\"" + ServiceDescription + "\", "
	dateofservice := "\"DateOfService\":\"" + DateOfService + "\", "
	dateofadmission := "\"DateOfAdmission\":\"" + DateOfAdmission + "\", "
	dateofdischarge := "\"DateOfDischarge\":\"UNDEFINED\", "
	dischargesummary := "\"DischargeSummary\":\"UNDEFINED\""

	healthrecord_json := "{" + healthrecordid + physicianname + facilityname + facilityaddress + typeofservice + servicedescription + dateofservice + dateofadmission + dateofdischarge + dischargesummary + "}"

	err := json.Unmarshal([]byte(healthrecord_json), &healthrecord)

	if caller_role == HEALTHCARE_ADMIN || caller_role == HEALTHCARE_USER {

		c.PersonHealth = append(c.PersonHealth, healthrecord)

	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. add_healthrecord"))
	}

	_, err = t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("ADD_HEALTHRECORD: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) update_healthrecord(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string, HealthRecordID string, DateOfDischarge string, DischargeSummary string) ([]byte, error) {

	if caller_role == HEALTHCARE_ADMIN || caller_role == HEALTHCARE_USER {
		for key, healthrecord := range c.PersonHealth {
			if healthrecord.HealthRecordID == HealthRecordID {
				fmt.Println("Updating health record")
				c.PersonHealth[key].DateOfDischarge = DateOfDischarge
				c.PersonHealth[key].DischargeSummary = DischargeSummary
			}
		}
	} else {
		return nil, errors.New(fmt.Sprint("Permission denied. update_healthrecord"))
	}

	_, err := t.save_changes(stub, c)

	if err != nil {
		fmt.Printf("UPDATE_HEALTHRECORD: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil

}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("query is running " + function)

	caller, caller_role, err := t.get_caller_data(stub)
	if err != nil {
		fmt.Printf("QUERY: Error retrieving caller details", err)
		return nil, errors.New("QUERY: Error retrieving caller details: " + err.Error())
	}

	logger.Debug("function: ", function)
	logger.Debug("caller: ", caller)
	logger.Debug("role: ", caller_role)

	if function == "get_person_details" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed")
			return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		c, err := t.retrieve_ID(stub, args[0])
		if err != nil {
			fmt.Printf("QUERY: Error retrieving ID: %s", err)
			return nil, errors.New("QUERY: Error retrieving ID " + err.Error())
		}
		return t.get_person_details(stub, c, caller, caller_role)
	} else if function == "get_health_details" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed")
			return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		c, err := t.retrieve_ID(stub, args[0])
		if err != nil {
			fmt.Printf("QUERY: Error retrieving ID: %s", err)
			return nil, errors.New("QUERY: Error retrieving ID " + err.Error())
		}
		return t.get_health_details(stub, c, caller, caller_role)
	} else if function == "check_unique_ID" {
		return t.check_unique_ID(stub, args[0], caller, caller_role)
	} else if function == "get_persons" {
		return t.get_persons(stub, caller, caller_role)
	} else if function == "get_ecert" {
		return t.get_ecert(stub, args[0])
	} else if function == "ping" {
		return t.ping(stub)
	} else if function == "read" {
		return t.read(stub, args)
	}

	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) get_person_details(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string) ([]byte, error) {

	var healthrecords []Health
	c.PersonHealth = healthrecords
	bytes, err := json.Marshal(c)

	if err != nil {
		return nil, errors.New("GET_PERSON_DETAILS: Invalid person object")
	}

	if caller_role == AUTHORITY {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied. get_person_details")
	}

}

func (t *SimpleChaincode) get_health_details(stub shim.ChaincodeStubInterface, c Citizen, caller string, caller_role string) ([]byte, error) {

	bytes, err := json.Marshal(c)

	if err != nil {
		return nil, errors.New("GET_HEALTH_DETAILS: Invalid person object")
	}

	if caller_role == AUTHORITY {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied. get_person_details")
	}

}

func (t *SimpleChaincode) get_persons(stub shim.ChaincodeStubInterface, caller string, caller_role string) ([]byte, error) {
	bytes, err := stub.GetState("PersonIDs")

	if err != nil {
		return nil, errors.New("Unable to get PersonIDs")
	}

	var PersonIDs ID_Holder

	err = json.Unmarshal(bytes, &PersonIDs)

	if err != nil {
		return nil, errors.New("Corrupt ID_Holder")
	}

	result := "["

	var temp []byte
	var c Citizen

	for _, ID := range PersonIDs.IDs {

		c, err = t.retrieve_ID(stub, ID)

		if err != nil {
			return nil, errors.New("Failed to retrieve ID")
		}

		temp, err = t.get_person_details(stub, c, caller, caller_role)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

func (t *SimpleChaincode) check_unique_ID(stub shim.ChaincodeStubInterface, ID string, caller string, caller_role string) ([]byte, error) {
	_, err := t.retrieve_ID(stub, ID)
	if err == nil {
		return []byte("false"), errors.New("ID is not unique")
	} else {
		return []byte("true"), nil
	}
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
