package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
)

func create(item interface{}, dbPath string) (err error) {
	// Use service account
	ctx := context.Background()
	sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
	app, appErr := firebase.NewApp(ctx, nil, sa)
	if appErr != nil {
		fmt.Println("Failed creating db app.")
		return appErr
	}

	client, clientErr := app.Firestore(ctx)
	if clientErr != nil {
		fmt.Println("Failed creating db client.")
		return clientErr
	}
	defer client.Close()

	// save data
	_, _, dbErr := client.Collection(dbPath).Add(ctx, item)
	if dbErr != nil {
		fmt.Println("Failed saving data to db.")
		return dbErr
	}
	return nil
}

// if err == nil && succeed = false, then the id is already used
func createWithId(id string, item interface{}, dbPath string) (succeed bool, err error) {
	// Use service account
	ctx := context.Background()
	sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
	app, appErr := firebase.NewApp(ctx, nil, sa)
	if appErr != nil {
		fmt.Println("Failed creating db app.")
		return false, appErr
	}

	client, clientErr := app.Firestore(ctx)
	if clientErr != nil {
		fmt.Println("Failed creating db client.")
		return false, clientErr
	}
	defer client.Close()

	// check if the id is used
	if queryResult, _ := readById(id, dbPath); queryResult != nil {
		return false, nil
	}

	// save data
	_, dbErr := client.Collection(dbPath).Doc(id).Set(ctx, item)
	if dbErr != nil {
		fmt.Println("Failed saving data to db.")
		return false, dbErr
	}
	return true, nil
}

// if err == nil && succeed = false, then the id doesn't exist
func updateById(id string, item interface{}, dbPath string) (succeed bool, err error) {
	// Use service account
	ctx := context.Background()
	sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
	app, appErr := firebase.NewApp(ctx, nil, sa)
	if appErr != nil {
		fmt.Println("Failed creating db app.")
		return false, appErr
	}

	client, clientErr := app.Firestore(ctx)
	if clientErr != nil {
		fmt.Println("Failed creating db client.")
		return false, clientErr
	}
	defer client.Close()

	// check if the id is used
	if queryResult, _ := readById(id, dbPath); queryResult == nil {
		return false, nil
	}

	// save data
	// firestore.MergeAll will only override provided properties rather than entire document
	_, dbErr := client.Collection(dbPath).Doc(id).Set(ctx, item, firestore.MergeAll)
	if dbErr != nil {
		fmt.Printf("Failed saving data to db. The error is %v\n", dbErr)
		return false, dbErr
	}
	return true,nil
}

//func readOrderFromDBByUser(username string, dbPath string) (queryResult []Order, err error) {
//	// Use a service account
//	ctx := context.Background()
//	sa := option.WithCredentialsFile(FireBasePrivateKeyPath)
//	app, appErr := firebase.NewApp(ctx, nil, sa)
//	if appErr != nil {
//		fmt.Println("Failed creating db app.")
//		return nil, appErr
//	}
//
//	client, clientErr := app.Firestore(ctx)
//	if clientErr != nil {
//		fmt.Println("Failed creating db client.")
//		return nil, clientErr
//	}
//	defer client.Close()
//
//	// read data
//	iter := client.Collection(dbPath).Where("username", "==", username).Documents(ctx)
//	for {
//		doc, docErr := iter.Next()
//		if docErr == iterator.Done {
//			break
//		}
//		if docErr != nil {
//			return nil, docErr
//		}
//		var order Order
//		dbDecodeErr := mapstructure.Decode(doc.Data(), &order)
//		if dbDecodeErr != nil {
//			return nil, dbDecodeErr
//		}
//		queryResult = append(queryResult, order)
//	}
//
//	return queryResult, nil
//}

func readById(id string, dbPath string) (map[string]interface{}, error) {
		ctx := context.Background()
		sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
		app, appErr := firebase.NewApp(ctx, nil, sa)
		if appErr != nil {
			fmt.Println("Failed creating db app.")
			return nil, appErr
		}

		client, clientErr := app.Firestore(ctx)
		if clientErr != nil {
			fmt.Println("Failed creating db client.")
			return nil, clientErr
		}
		defer client.Close()

		// read data
		document, dbErr := client.Collection(dbPath).Doc(id).Get(ctx)
		if dbErr != nil {
			return nil, dbErr
		}

		if !document.Exists() {
			return nil, dbErr
		}

		return document.Data(), nil
}
//
//func readAll(username string, dbPath string) (queryResult []Order, err error) {
//
//}

func deleteById(id string, dbPath string) (err error) {
	// Use a service account
	ctx := context.Background()
	sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
	app, appErr := firebase.NewApp(ctx, nil, sa)
	if appErr != nil {
		fmt.Println("Failed creating db app.")
		return appErr
	}

	client, clientErr := app.Firestore(ctx)
	if clientErr != nil {
		fmt.Println("Failed creating db client.")
		return clientErr
	}
	defer client.Close()

	_, dbErr := client.Collection("dbPath").Doc(id).Delete(ctx)
	if dbErr != nil {
		fmt.Println("Failed removing date from db.")
		return dbErr
	}
	return nil
}
