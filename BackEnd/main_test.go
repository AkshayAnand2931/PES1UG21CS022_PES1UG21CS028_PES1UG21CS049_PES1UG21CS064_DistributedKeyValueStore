package main_test

import(
	"testing"
	"net/http"
	"bytes"
	"encoding/json"
)

const REQUEST_URL string = "http://localhost:4242/"

func TestGetNotSet(t *testing.T){
	res, _ := http.Get(REQUEST_URL+"get?key=doesnotexist")
	if res.StatusCode != 404{
		t.Fatalf("Did not get a not found error on doesnotexist")
	}
}

func TestGetAlreadySet(t *testing.T){
	res, _ := http.Get(REQUEST_URL+"get?key=exists")
	if res.StatusCode != 200{
		t.Fatalf("Not OK error on exists")
	}
}

func TestSet(t *testing.T){
	jsonBod, _ := json.Marshal(map[string]string{
		"key": "onlySet",
		"value": "TheValueSet",
	})
	reqBody := bytes.NewBuffer(jsonBod)
	res, err := http.Post(REQUEST_URL+"set", "application/json", reqBody)
	if res.StatusCode != 200 || err != nil{
		t.Fatalf("Value not set")
	}
}

func TestSetAndGet(t *testing.T){
	jsonBod, _ := json.Marshal(map[string]string{
		"key": "setAndGet",
		"value": "TheValueSet",
	})
	reqBody := bytes.NewBuffer(jsonBod)
	res, err := http.Post(REQUEST_URL+"set", "application/json", reqBody)
	if res.StatusCode != 200 || err != nil{
		t.Fatalf("Value not set")
	}

	res, err = http.Get(REQUEST_URL + "get?key=setAndGet")
	if res.StatusCode != 200{
		t.Fatalf("Not OK error on setAndGet")
	}
}

func TestDelete(t *testing.T){
	jsonBod, _ := json.Marshal(map[string]string{
		"key": "onlyDel",
		"value": "TheValueSet",
	})
	reqBody := bytes.NewBuffer(jsonBod)
	res, err := http.Post(REQUEST_URL+"set", "application/json", reqBody)
	if res.StatusCode != 200 || err != nil{
		t.Fatalf("Value not set")
	}

	client := &http.Client{}
	req, err1 := http.NewRequest("DELETE", REQUEST_URL+"delete?key=onlyDel", nil)
	if err1 != nil{
		t.Fatalf("Could not send request")
	}
	res, err = client.Do(req)
	if res.StatusCode != 200 || err != nil{
		t.Fatalf("Delete gave error")
	}
	defer res.Body.Close()
}

func TestGetAll(t *testing.T){
	res, _ := http.Get(REQUEST_URL+"getAll")
	if res.StatusCode != 200{
		t.Fatalf("Error on getAll")
	}
}
