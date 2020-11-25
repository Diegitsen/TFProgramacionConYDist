package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//Person
type Person struct {
	ID                   int    `json:"ID"`
	Age                  string `json:"age"`
	Sex                  string `json:"sex"`
	RestingBloodPressure string `json:"blood_Pressure"`
	SerumCholestrol      string `json:"cholestrol"`
}

//Persons Array
type allPersons []Person

var persons = allPersons{
	{
		ID:                   1,
		Age:                  "50",
		Sex:                  "Female",
		RestingBloodPressure: "",
		SerumCholestrol:      "",
	},
}

func indexRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my REST API developed in GO!")
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	var newPerson Person
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Insert a Valid Person Data")
	}

	json.Unmarshal(reqBody, &newPerson)
	newPerson.ID = len(persons) + 1
	persons = append(persons, newPerson)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPerson)

}

func getAllPersons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}

func getPersonByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personID, err := strconv.Atoi(vars["id"])
	if err != nil {
		return
	}

	for _, personObject := range persons {
		if personObject.ID == personID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(personObject)
		}
	}
}

func updatePerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personID, err := strconv.Atoi(vars["id"])
	var updatedPerson Person

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please Enter Valid Data")
	}
	json.Unmarshal(reqBody, &updatedPerson)

	for i, p := range persons {
		if p.ID == personID {
			persons = append(persons[:i], persons[i+1:]...)

			updatedPerson.ID = p.ID
			persons = append(persons, updatedPerson)

			// w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(updatedPerson)
			fmt.Fprintf(w, "The task with ID %v has been updated successfully", personID)
		}
	}

}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, "Invalid User ID")
		return
	}

	for i, p := range persons {
		if p.ID == personID {
			persons = append(persons[:i], persons[i+1:]...)
			fmt.Fprintf(w, "The task with ID %v has been remove successfully", personID)
		}
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", indexRoute)
	router.HandleFunc("/persons", createPerson).Methods("POST")
	router.HandleFunc("/persons", getAllPersons).Methods("GET")
	router.HandleFunc("/persons/{id}", getPersonByID).Methods("GET")
	router.HandleFunc("/persons/{id}", deletePerson).Methods("DELETE")
	router.HandleFunc("/persons/{id}", updatePerson).Methods("PUT")

	log.Fatal(http.ListenAndServe(":3000", router))
}
