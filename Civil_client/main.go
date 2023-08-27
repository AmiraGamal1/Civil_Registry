package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

// citizen data
type Person struct {
	CVname string
	CVid   int32
}

type State struct {
	CVstate string
}

func checkErr(err error) {
	/*
		check error type
	*/
	if err != nil {
		log.Fatal(err)
	}
}

func insert(sock net.Conn) {
	/*
		1- recieve person data from command line
		2- check if data type of data is vaild
		3- check if citizen data are complete (error :  0 : Data is not complete)
		4- store data in struct
		5- encode person data
		6- send data to server
		7- recieve from server insert status(success(1) , fail(2): Duplicate ID )
		8- print status
	*/
	var state, name, f, l, ll, lll string
	var id int32
	fmt.Print("Enter Citizen Data\n")
	fmt.Print("cvID: ")
	_, err := fmt.Scan(&id)
	fmt.Print("cvName (Quadruple name): ")
	fmt.Scan(&f, &l, &ll, &lll) // Quadruple name
	fmt.Print("cvState: ")
	fmt.Scan(&state)
	name = f + " " + l + " " + ll + " " + lll
	if err != nil || len(name) == 0 || len(state) == 0 {
		println("Data is not complete")
	} else {
		//store data in to person struct
		EnPerson := Person{CVname: string(name), CVid: id}
		EnState := State{CVstate: string(state)}

		encode := gob.NewEncoder(sock)
		perr := encode.Encode(EnPerson)
		if perr != nil {
			log.Panic("error to encode person ", perr)
		}
		serr := encode.Encode(EnState)
		if serr != nil {
			log.Panic("error to encode state ", serr)
		}
		// recieved status from server
		var status int
		de := gob.NewDecoder(sock)
		de.Decode(&status)
		if status == 1 {
			println("Success(insert)")
		} else {
			println("fail(insert):Duplicate ID")
		}
	}
}

func delete(sock net.Conn) {
	/*
		recieve id from client command line
		encode id
		send encoded id to server
		recieve delete status sucess(1), fail(2) if id not in database
	*/
	var id int32
	fmt.Print("Enter ID : ")
	fmt.Scan(&id)
	en := gob.NewEncoder(sock)
	en.Encode(&id)

	var status int
	de := gob.NewDecoder(sock)
	de.Decode(&status)
	if status == 1 {
		fmt.Println("Success")
	} else {
		fmt.Println("fail: Not found cv_Id ")
	}
}
func find(sock net.Conn) {
	/*
		recieve id from client
		encode id
		send encoded id to server
		recieve find status sucess(1) fail(2): cannot Find cvID in Civil Registry database
	*/
	var id int32
	fmt.Print("cvID: ")
	_, err := fmt.Scan(&id)
	if err != nil {
		log.Fatal("write id error ", err)
	}
	//send id

	de := gob.NewEncoder(sock)
	de.Encode(&id)

	var status int
	// recived status
	//read status
	dec := gob.NewDecoder(sock)
	dec.Decode(&status)

	if status == 1 {
		println("Success(find)")
		decoder := gob.NewDecoder(sock)
		p := &Person{}
		s := &State{}

		perr := decoder.Decode(p)
		if perr != nil {
			log.Panic("error decode person ", perr)
		}
		serr := decoder.Decode(s)
		if serr != nil {
			log.Panic("error decode state ", serr)
		}
		fmt.Println("cvID: ", p.CVid, "\ncvName: ", p.CVname, "\ncvState: ", s.CVstate)
	} else {
		println("fail(find):Not found ID")
	}

}

func List(sock net.Conn) { // rewite to recieve all database
	/*
		recieve records' data from server
		decode records' data
		print out records
	*/
	cv_de := gob.NewDecoder(sock)
	var n int
	cv_de.Decode(&n)
	fmt.Println("we have ", n, " records")
	fmt.Println("No. ", "cv_ID	", "cv_Name			", "cv_State")
	for i := 0; i < n; i++ {
		var cv_p Person
		var cv_s State
		cv_de.Decode(&cv_p)
		cv_de.Decode(&cv_s)
		fmt.Println(i, " ", cv_p.CVid, "	", cv_p.CVname, "			", cv_s.CVstate)
	}
}

func main() {
	/*
		start connection
		recieve from client what operation want
		send opertion to server
		call opertion function to communicate with server
	*/
	if len(os.Args) != 3 {
		log.Fatal("Uasage : client <hostname> <port>")
	}
	sock, err := net.Dial("tcp", "localhost:8080")
	checkErr(err) // print connection error
	println("select opertion type:\ni : insert\nf: find\nd: delete\nl: list")
	var op string
	fmt.Scanf("%s", &op)
	if len(op) != 1 {
		println("invaild opertion")
	} else {
		// send opertion to server

		_, err := sock.Write([]byte(op))
		checkErr(err)
		// call client opertion
		switch op {
		case "i":
			insert(sock)
		case "f":
			find(sock)
		case "d":
			delete(sock)
		case "l":
			List(sock)
		}
	}
	println("Done")
	sock.Close()
}
