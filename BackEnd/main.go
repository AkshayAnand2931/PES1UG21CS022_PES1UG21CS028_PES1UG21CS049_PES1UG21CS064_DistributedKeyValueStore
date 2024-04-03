package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	client *clientv3.Client
)

func main() {

	endpoints := []string{"localhost:2379"}
	duration := 5 * time.Second

	//Create client for etcd
	var err error
	client, err = getClient(endpoints, duration)

	if err != nil {
		fmt.Println("Failed to create etcd client: ", err)
		return
	}

	defer client.Close()

	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/getAll", getAllHandler)

	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func getClient(endpoints []string, duration time.Duration) (*clientv3.Client, error) {

	//Get client for etcd
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: duration,
	})
}

func setHandler(w http.ResponseWriter, r *http.Request) {

	var keyvalue KeyValue

	if err := json.NewDecoder(r.Body).Decode(&keyvalue); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//Set key-value pair
	ctx := context.TODO()
	_, err := client.Put(ctx, keyvalue.Key, keyvalue.Value)

	if err != nil {
		http.Error(w, "Failed to set key-value pair in etcd", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query().Get("key")

	ctx := context.TODO()
	resp, err := client.Get(ctx, key)

	if err != nil {
		http.Error(w, "Failed to get value for key from etcd", http.StatusInternalServerError)
		return
	}

	if len(resp.Kvs) == 0 {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	value := string(resp.Kvs[0].Value)
	jsonResponse(w, KeyValue{Key: key, Value: value})
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()
	resp, err := client.Get(ctx, "", clientv3.WithPrefix())

	if err != nil {
		http.Error(w, "Failed to get all key-value pairs from etcd", http.StatusInternalServerError)
		return
	}

	var keyvalues []KeyValue

	for _, kv := range resp.Kvs {
		keyvalues = append(keyvalues, KeyValue{Key: string(kv.Key), Value: string(kv.Value)})
	}

	jsonResponse(w, keyvalues)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func watchChanges(client *clientv3.Client, key string) {

	ctx := context.TODO()
	watcher := client.Watch(ctx, key, clientv3.WithPrefix())

	for wresp := range watcher {

		for _, event := range wresp.Events {
			fmt.Printf("Event recieved! Type: %s, Key: %s, Value: %s \n", event.Type, event.Kv.Key, event.Kv.Value)
		}
	}
}
