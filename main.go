// todo: query - call algorithm and return solutions
// todo: placing order - check if you can place the order; lock machine
// todo: payment - periodically check and remove order; unlock machine
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/pborman/uuid"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ORDERS_COLLECTION string = "orders"
	ORDER_EXPIRE_SECONDS int = 30
	//SOLUTION_EXPIRE_SECONDS int = 10 * 60 // 5-1
	SECOND_TO_NANOSECOND int = 1000000000
)

var (
	ed25519PublicKey []byte = make([]byte, 32)
	ed25519PrivateKey []byte = make([]byte, 64)
	orderMap map[string]Order = make(map[string]Order)
	chanMap map[string][]chan bool = make(map[string][]chan bool)
)

func main() {
	fmt.Println("started-service")
	generateKey()
	r := mux.NewRouter()
	r.Handle("/query", http.HandlerFunc(handlerQuery)).Methods("POST", "OPTIONS")
	r.Handle("/order", http.HandlerFunc(handlerPlacingOrder)).Methods("POST", "OPTIONS")
	r.Handle("/pay", http.HandlerFunc(handlerPayment)).Methods("POST", "OPTIONS")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func handlerQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one query request")

	// Parse from body of request to get a json object.
	decoder := json.NewDecoder(r.Body)
	var query Query
	if err := decoder.Decode(&query); err != nil {
		http.Error(w, "Failed decoding query data.", http.StatusBadRequest)
		return
	}
	query.Time = time.Now()

	// handle query
	// solutions = generateSolutions(*query)
	// ------hard code------
	//solution1 := Solution{MachineType: 1, Price: 100, QueryTime: query.Time} // 5-2
	solution1 := Solution{MachineType: 1, Price: 100}
	solution2 := Solution{MachineType: 0}
	var solutions []Solution
	solutions = append(solutions, solution1, solution2)
	// ------hard code------

	var signatures []Signature
	for _, val := range solutions {
		signatures = append(signatures, ed25519.Sign(ed25519PrivateKey, []byte(val.ToString())))
	}

	response := map[string]interface{}{"solutions": solutions, "signatures": signatures}
	js, passErr := json.Marshal(response)
	if passErr != nil {
		http.Error(w, "Failed to pass response into JSON object", http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func handlerPlacingOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one order request")

	// verify token
	idToken := r.Header.Get("Authorization")
	fmt.Printf("%v\n", idToken)
	authToken, verifyErr := verifyToken(idToken)
	if verifyErr != nil {
		http.Error(w, "Failed verifying identity.", http.StatusUnauthorized)
		return
	}

	// Parse from body of request to get a json object.
	decoder := json.NewDecoder(r.Body)
	var orderRequest OrderRequest
	if err := decoder.Decode(&orderRequest); err != nil {
		http.Error(w, "Failed decoding order data.", http.StatusBadRequest)
		return
	}
	fmt.Println(orderRequest)

	//check solution and signature
	solution := orderRequest.Solution
	if !ed25519.Verify(ed25519PublicKey, []byte(solution.ToString()), orderRequest.Signature) {
		http.Error(w, "Invalid order.", http.StatusBadRequest)
		return
	}

	// check if solution is expired, 5-3
	//if time.Since(solution.QueryTime) > time.Duration(SOLUTION_EXPIRE_SECONDS * SECOND_TO_NANOSECOND) {
	//	http.Error(w, "Expired solution.", http.StatusBadRequest)
	//	return
	//}

	// get machine id and lock machine
	// machineId := assignMachine(*solution)

	// generate order
	orderId := uuid.New()
	order := Order{ authToken.UID,
					solution.StartingPoint,
					solution.Destination,
					time.Now(),
					solution.StartTime,
					solution.PickUpTime,
					solution.DropOffTime,
					solution.EndTime,
					solution.MachineType,
					solution.Weight,
					solution.Price,
					"machine_id",
					NOTORDERED,
	}

	// store to db and memory
	createWithId(orderId, order, ORDERS_COLLECTION)
	orderMap[orderId] = order

	//go routine
	// create corresponding channels
	chanMap[orderId] = []chan bool{make(chan bool), make(chan bool)}
	go func(orderId string) {
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()
		chans := chanMap[orderId]
		for {
			select {
				case <-chans[0]:
					if success := <-chans[1]; success {
						updateStatus(orderId, ORDERED)
						return
					}
				case <-ticker.C:
					if time.Since(orderMap[orderId].PlacingTime) > time.Duration(ORDER_EXPIRE_SECONDS * SECOND_TO_NANOSECOND) {
						updateStatus(orderId, EXPIRED)
						return
					}
			}
		}
	}(orderId)

	// response
	fmt.Printf("Order is created.\n")
	js, err := json.Marshal(map[string]string{"order_id": orderId})
	if err != nil {
		http.Error(w, "Failed to pass orderId into JSON object", http.StatusInternalServerError)
		return
	}

	w.Write(js)
	//fmt.Fprintf(w, "\nCongradulations! Your order has been created succeesfully!")
}

func updateStatus(orderId string, status int) {
	chans := chanMap[orderId]
	delete(orderMap, orderId)
	delete(chanMap, orderId)
	close(chans[0])
	close(chans[1])
	succeed, dbErr := updateById(orderId, map[string]int{"Status": status}, ORDERS_COLLECTION)
	if !succeed || dbErr != nil {
		fmt.Printf("\nFailed update status to db: %v", dbErr)
	}
}

func handlerPayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one payment request")

	// verify token
	idToken := r.Header.Get("Authorization")
	fmt.Printf("%v\n", idToken)
	authToken, verifyErr := verifyToken(idToken)
	if verifyErr != nil {
		http.Error(w, "Failed verifying identity.", http.StatusUnauthorized)
		return
	}

	// parse orderId
	decoder := json.NewDecoder(r.Body)
	var paymentRequest PaymentRequest
	if err := decoder.Decode(&paymentRequest); err != nil {
		http.Error(w, "Failed decoding payment data.", http.StatusBadRequest)
		return
	}
	orderId := paymentRequest.OrderId

	// get order and check if user_token fits the corresponding order
	order, exist := orderMap[orderId]
	if !exist {
		http.Error(w, "Order is nonexistent or expired .", http.StatusUnauthorized)
		return
	}
	if order.UserId != authToken.UID {
		http.Error(w, "Operation is unauthorized.", http.StatusUnauthorized)
		return
	}

	chanMap[orderId][0] <- true

	// call stripe
	stripe.Key = STRIPEPRIVATEKEY

	price := order.Price // in cents
	params := &stripe.ChargeParams{
		Amount: stripe.Int64(int64(price)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String("Humming Product"),
	}

	paymentToken := r.Header.Get("Payment")
	params.SetSource(paymentToken)
	_, chargeErr := charge.New(params)
	if chargeErr != nil {
		if stripeErr, ok := chargeErr.(*stripe.Error); ok {
			http.Error(w, stripeErr.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, chargeErr.Error(), http.StatusBadRequest)
		}
		chanMap[orderId][1] <- false
		return
	}

	chanMap[orderId][1] <- true

	// update order in db
	//order.Status = ORDERED
	//succeed, dbErr := updateById(orderId, order, ORDERS_COLLECTION)
	//if !succeed || dbErr != nil {
	//	updateById(orderId, order, ORDERS_COLLECTION)
	//}
	// clear order in memory
	//fmt.Fprintf(w, "Congradulations! Your order has been created succeesfully!")
}

func verifyToken(idToken string) (*auth.Token, error) {
	ctx := context.Background()
	sa := option.WithCredentialsFile(FIREBASEPRIVATEKEYPATH)
	app, appErr := firebase.NewApp(ctx, nil, sa)
	if appErr != nil {
		fmt.Println("Failed creating auth app.")
		return nil, appErr
	}

	client, clientErr := app.Auth(ctx)
	if clientErr != nil {
		fmt.Printf("Failed getting Auth client: %v\n", clientErr)
		return nil, clientErr
	}

	token, verificationErr := client.VerifyIDToken(ctx, idToken)
	if verificationErr != nil {
		fmt.Printf("Failed verifying ID token: %v\n", verificationErr)
		return nil, verificationErr
	}

	fmt.Printf("Verified ID token: %v\n", token)
	return token, nil
}

func generateKey(){
	for index, val := range strings.Fields(ED25519PUBLICKEYSTRING) {
		num, _ := strconv.Atoi(val)
		ed25519PublicKey[index] = byte(num)
	}

	for index, val := range strings.Fields(ED25519PRIVATEKEYSTRING) {
		num, _ := strconv.Atoi(val)
		ed25519PrivateKey[index] = byte(num)
	}
}