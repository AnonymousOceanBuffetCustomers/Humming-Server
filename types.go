package main

import (
	"fmt"
	"time"
)

const (
	DRONEONLY int = 0;
	ROBOTONLY int = 1;
	BOTH int = 2;

	NOTORDERED int = 0;
	ORDERED int = 1;
	STARTED int = 2;
	PICKEDUP int = 3
	FINISHED int = 4;
	EXPIRED int = 5;

	NOTINSTATION int = 0;
	STATION0 int = 1;
	STATION1 int = 2;
	STATION2 int = 3;
)

type Signature []byte

// weight
type Query struct {
	StartingPoint Location   `json:"starting_point"`
	Destination   Location   `json:"destination"`
	MachineType   int  		 `json:"machine_type"`
	Time 		  time.Time	 `json:"time"`
	Weight   	  float64	 `json:"weight"`
}

// starting point
// destination
// weight
type Solution struct {
	StartingPoint Location   	`json:"starting_point"`
	Destination   Location   	`json:"destination"`
	//QueryTime	  time.Time		`json:"query_time"` // 5-4
	StartTime	  time.Time		`json:"start_time"`
	PickUpTime	  time.Time		`json:"estimated_pickup_time"`
	DropOffTime	  time.Time		`json:"estimated_dropoff_time"`
	EndTime       time.Time		`json:"end_time"`
	Station 	  int			`json:"station"`
	MachineType   int  			`json:"machine_type"`
	Price 		  float64		`json:"price"`
	Weight   	  float64		`json:"weight"`
}

type Coordinate struct {
	Lat     float64   `json:"lon"`
	Lon     float64   `json:"lat"`
}

type Location struct {
	Coordinate Coordinate `json:"coordinate"`
	Address    string  	  `json:"address"`
}

type Order struct {
	UserId string
	StartingPoint Location
	Destination Location
	PlacingTime time.Time
	StartTime time.Time
	PickUpTime time.Time
	DropOffTime time.Time
	EndTime time.Time
	MachineType int
	Weight float64
	Price float64
	MachineId string
	Status int
}

type User struct {
	Username string
	ImageUrl string
	CreditCards []string
	Addresses []string
}

type Machine struct {
	MachineType int
	Coordinate Coordinate
	Station int
	Intervals string
	Electricity float64
}

func (location Location) ToString() string {
	return fmt.Sprintf("lon %v; lat %v; address %v", location.Coordinate.Lon, location.Coordinate.Lat, location.Address)
}

func (solution Solution) ToString() string {
	return fmt.Sprintf("StartingPoint %v; Destination %v; StartTime %v; PickUpTime %v; DropOffTime %v; EndTime %v; Station %v; MachineType %v; Price %v;  Weight %v",
						solution.StartingPoint, solution.Destination, solution.StartTime, solution.PickUpTime, solution.DropOffTime, solution.EndTime, solution.Station, solution.MachineType, solution.Price, solution.Weight)
}

// 5-5
//func (solution Solution) ToString() string {
//	return fmt.Sprintf("StartingPoint %v; Destination %v; QueryTime %v; StartTime %v; PickUpTime %v; DropOffTime %v; EndTime %v; Station %v; MachineType %v; Price %v;  Weight %v",
//		solution.StartingPoint, solution.Destination, solution.QueryTime,solution.StartTime, solution.PickUpTime, solution.DropOffTime, solution.EndTime, solution.Station, solution.MachineType, solution.Price, solution.Weight)
//}

type OrderRequest struct {
	Solution Solution   `json:"solution"`
	Signature Signature `json:"signature"`
}

type PaymentRequest struct {
	OrderId string	`json:"order_id"`
}