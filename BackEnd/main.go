package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {

	endpoints := []string{"localhost:2379"}

	client, err := getClient(endpoints)

	if err != nil {
		fmt.Println("Failed to create etcd client: ", err)
		return
	}

	defer client.Close()

	//Set a key-value pair
	key := "example_key"
	value := "example_value"

	if err := setValue(client, key, value); err != nil {
		fmt.Println("Failed to put key-value pairs: ", err)
		return
	}

	//Retrieve a value by key
	if val, err := getValue(client, key); err != nil {
		fmt.Println("Failed to get value for key: ", err)
	} else {
		fmt.Println("Value for key:", val)
	}

	//Watch for changes
	watchChanges(client, key)
}

func getClient(endpoints []string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
}

func setValue(client *clientv3.Client, key, value string) error {

	ctx := context.TODO()
	_, err := client.Put(ctx, key, value)
	return err
}

func getValue(client *clientv3.Client, key string) (string, error) {

	ctx := context.TODO()
	resp, err := client.Get(ctx, key)

	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("Key not found")
	}

	return string(resp.Kvs[0].Value), nil
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
