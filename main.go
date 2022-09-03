package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func main() {
	fmt.Println("Reading Config")
	var data []byte
	var err error
	if _, fileError := os.Stat("/config/schema.json"); errors.Is(fileError, os.ErrNotExist) {
		data, err = os.ReadFile("./schema.json")
	} else {
		data, err = os.ReadFile("/config/schema.json")
	}

	if err != nil {
		fmt.Println("error reading /config/schema.json")
		return
	}

	schema := schema{}
	err = json.Unmarshal([]byte(data), &schema)
	if err != nil {
		fmt.Println("Error parsing config. ", err)
		return
	}

	//TODO: Stop hardcoding this
	var url = "http://admin:password@localhost:41000/"

	for _, database := range schema.Databases {
		resp, err := http.Get(fmt.Sprintf(url+"%s", database.Name))
		if err != nil {
			fmt.Println("Error getting database from server. ", err)
			continue
		}
		couchResponse := couchResponse{}
		respBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(respBytes, &couchResponse)
		if resp.StatusCode == 404 && couchResponse.Error == "not_found" {
			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(url+"%s", database.Name), nil)
			if err != nil {
				fmt.Println("Error creating database request. ", err)
				continue
			}
			_, err = client.Do(req)
			if err != nil {
				fmt.Println("Error sending request to create database. ", err)
				continue
			} else {
				fmt.Println("Successfully created: " + database.Name)
			}
		} else {
			fmt.Println(database.Name + " is already created")
		}
		for _, designdoc := range database.DesignDocs {
			resp, err = http.Get(fmt.Sprintf(url+"%s/_design/%s", database.Name, designdoc.Name))
			if err != nil {
				fmt.Println(fmt.Sprintf("Error getting view(%s) from server. ", designdoc.Name), err)
				continue
			}
			respBytes, _ := ioutil.ReadAll(resp.Body)
			json.Unmarshal(respBytes, &couchResponse)
			newDesignDoc := designDoc{}
			if resp.StatusCode == 404 && couchResponse.Error == "not_found" {
				newDesignDoc = designdoc
			} else {
				fmt.Println(designdoc.Name + " is already created for database " + database.Name)
				json.Unmarshal(respBytes, &newDesignDoc)
				for key, value := range designdoc.Views {
					newDesignDoc.Views[key] = value
				}
			}
			client := &http.Client{}
			designJson, _ := json.Marshal(newDesignDoc)
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(url+"%s/_design/%s", database.Name, designdoc.Name), bytes.NewBuffer(designJson))
			if err != nil {
				fmt.Println("Error creating designdoc request. ", err)
				continue
			}
			resp, err = client.Do(req)
			respBytes, _ = ioutil.ReadAll(resp.Body)
			json.Unmarshal(respBytes, &couchResponse)
			if err != nil || couchResponse.Error != "" {
				fmt.Println("Error sending request to create designdoc. ", err, couchResponse)
				continue
			} else {
				fmt.Println("Successfully created/updated: " + designdoc.Name + " status code: " + strconv.Itoa(resp.StatusCode))
			}
		}
	}
}

type schema struct {
	Databases []database `json:"databases"`
}

type database struct {
	Name       string      `json:"name"`
	DesignDocs []designDoc `json:"designdocs"`
}

type designDoc struct {
	Name     string                 `json:"name"`
	Language string                 `json:"language"`
	Views    map[string]interface{} `json:"views"`
	Id       string                 `json:"_id"`
	Rev      string                 `json:"_rev"`
}

type couchResponse struct {
	Error string `json:"error"`
}
