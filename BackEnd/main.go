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
	endpoints := []string{"127.0.0.1:4243", "127.0.0.1:4244", "127.0.0.1:4245"} // ETCD server endpoints
	duration := 5 * time.Second

	// Create client for etcd
	var err error
	client, err = getClient(endpoints, duration)

	if err != nil {
		fmt.Println("Failed to create etcd client: ", err)
		return
	}

	defer client.Close()

	// Set up HTTP handlers
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/getAll", getAllHandler)
	http.HandleFunc("/delete", deleteHandler)

	// Enable CORS middleware
	handler := enableCORS(http.DefaultServeMux)
	fmt.Println("Server listening on http://127.0.0.1:4242")
	fmt.Println(http.ListenAndServe("127.0.0.1:4242", handler))
}

// Middleware function to enable CORS
func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func getClient(endpoints []string, duration time.Duration) (*clientv3.Client, error) {
	// Get client for etcd
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: duration,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
        http.Error(w, "Only DELETE is allowed", http.StatusMethodNotAllowed)
        return
    }
	// Parse the key from the request URL
	key := r.URL.Query().Get("key")

	// Check if the key is empty
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Delete the key from the KV store
	ctx := context.TODO()
	_, err := client.Delete(ctx, key)

	if err != nil {
		http.Error(w, "Failed to delete key from etcd", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Key '%s' deleted", key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
        return
    }

	var keyvalue KeyValue
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&keyvalue); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set key-value pair
	ctx := context.TODO()
	_, err := client.Put(ctx, keyvalue.Key, keyvalue.Value)

	if err != nil {
		http.Error(w, "Failed to set key-value pair in etcd", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
        http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
        return
    }

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
	if r.Method != http.MethodGet {
        http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
        return
    }
	
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
