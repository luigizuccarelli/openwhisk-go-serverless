package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

// A simple golang serverless function to handle IOTData (channels (on/off) and temperature)
// from a custom IOT device

// IOTData schema - we have to define a specific struct for our data
type IOTData struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	Channels     []bool
	Temperatures []float64
	Last         time.Time
}

// Response schema - used as final response to and from our serverless function
type Response struct {
	StatusCode string    `json:"statuscode"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Payload    []IOTData `json:"payload"`
}

func main() {
	// native actions receive one argument, the JSON object as a string
	arg := os.Args[1]

	// unmarshal the string to a JSON object
	var obj map[string]interface{}
	json.Unmarshal([]byte(arg), &obj)

	// some input validation
	ip, ipErr := obj["ip"].(string)
	if !ipErr {
		ip = "localhost"
	}
	db, dbErr := obj["db"].(string)
	if !dbErr {
		db = "test"
	}
	action, actionErr := obj["action"].(string)
	if !actionErr {
		action = "read"
	}

	fmt.Printf("INFO %s %s %s\n", ip, db, action)

	var response Response
	var payload []IOTData
	var tempData IOTData

	// database setup and init
	session, err := mgo.Dial(ip + ":27017")
	if err != nil {
		response = Response{StatusCode: "500", Status: "KO", Message: "Error database connection " + ip, Payload: nil}
		fmt.Println("err connection " + ip)
	} else {
		// if session was ok
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		// Collection iotdata
		c := session.DB(db).C("iotdata")

		if action == "read" {
			iter := c.Find(nil).Limit(50).Sort("-$natural").Iter()
			if iter.Err() != nil {
				response = Response{StatusCode: "500", Status: "KO", Message: "Error iterator", Payload: nil}
			} else {
				for iter.Next(&tempData) {
					payload = append(payload, tempData)
				}
				response = Response{StatusCode: "200", Status: "OK", Message: "Info success", Payload: payload}
				iter.Close()
			}
		} else {
			// check for the payload
			data, dataErr := obj["payload"]
			// report if error
			if !dataErr {
				response = Response{StatusCode: "500", Status: "KO", Message: "Error payload not defined ", Payload: nil}
			} else {
				// convert to interface then []interface and finally make a float64 array
				// for both channels and temperatures
				index, _ := data.(map[string]interface{})
				tmpCh := index["channels"].([]interface{})
				ch := make([]bool, len(tmpCh))
				for i := range tmpCh {
					ch[i] = bool(tmpCh[i].(float64) != 0)
				}
				tmpT := index["temperatures"].([]interface{})
				t := make([]float64, len(tmpT))
				for i := range tmpT {
					t[i] = tmpT[i].(float64)
				}
				// debug output
				fmt.Printf("INFO %v %v\n", ch, t)
				// create with our IOTData schema
				tempData = IOTData{Channels: ch, Temperatures: t, Last: time.Now()}
				err = c.Insert(&tempData)
				if err != nil {
					response = Response{StatusCode: "500", Status: "KO", Message: "Error inserting json data to db ", Payload: append(payload, tempData)}
				} else {
					response = Response{StatusCode: "200", Status: "OK", Message: "Info record inserted successfully ", Payload: append(payload, tempData)}
				}
			}
		}
	}

	// return the response (must be last line of our function)
	res, _ := json.Marshal(response)
	fmt.Println(string(res))

}
