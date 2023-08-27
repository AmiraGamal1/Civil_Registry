package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Store citizen data into Struct
type Person struct {
	CVname string
	CVid   int32
}
type State struct {
	CVstate string
}

func insert(sock net.Conn) {
	/*	recieve client encoded struct data
		decode data
		if id in csv file: send fail(0)
		else
			1- store data into buffer to increase concurancy
		    2 - store data into csv file
			3 - send sucess(1) to client
	*/
	var i_file, _ = os.OpenFile("CivilRegistry.csv", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644) // open database
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
	reader := csv.NewReader(i_file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}
	var status int
	//test if record exist
	//send fail (0)
	status = 1
	for _, record := range records {
		id, _ := strconv.Atoi(record[0])
		if int32(id) == p.CVid {
			status = 0 // fail duplicated id
			break
		}
	}

	//fmt.Println("Hello ", p.CVname, ", Your ID is ", p.CVid, " state ", s.CVstate)
	if status == 1 {
		writer := csv.NewWriter(i_file)
		defer writer.Flush()
		cv_record := []string{strconv.Itoa(int(p.CVid)), p.CVname, s.CVstate} // store data as map to dynamicly manipulation

		error := writer.Write(cv_record)
		if error != nil {
			log.Fatalln("error writing record to file ", error)
		}
		//send success (1)
	}
	en := gob.NewEncoder(sock)
	en.Encode(status)
}

func delete(sock net.Conn) {
	/*
		1- recieve id
		2- search if id in file
				if not send fail(2)
		3- set Write lock
		4- rewite recode with empty string
		5- send sucess(1)
	*/
	var id int32
	de := gob.NewDecoder(sock)
	de.Decode(&id)

	var d_file, _ = os.OpenFile("CivilRegistry.csv", os.O_RDWR, 0644) // open database in read write mode

	reader := csv.NewReader(d_file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	var status int
	//test if record exist
	//send fail (0)
	status = 0
	// search if id in csv file
	var index int // index of deleted record
	for i, record := range records {
		ID, _ := strconv.Atoi(record[0])
		if int32(ID) == id {
			index = i
			status = 1 // find id
			break
		}
	}
	if status == 1 { // if record if found
		file, err := ioutil.ReadFile("CivilRegistry.csv")
		if err != nil {
			panic(err)
		}

		info, _ := os.Stat("CivilRegistry.csv")
		mode := info.Mode()

		array := strings.Split(string(file), "\n")                                     // store database in array of string
		array = append(array[:index], array[index+1:]...)                              // ignore deleted record
		ioutil.WriteFile("CivilRegistry.csv", []byte(strings.Join(array, "\n")), mode) // rewrite file
	}

	en := gob.NewEncoder(sock)
	en.Encode(status)
}

func find(sock net.Conn) {
	/*	recieve id
		read all csv file
		search id in to col 0
			if id Existing :
				1 - send sucess(1)
			  	2 - send encoded struct data
			else :
				send fail(2)
	*/
	//recieve id
	var id int32
	de := gob.NewDecoder(sock)
	de.Decode(&id)
	var f_file, _ = os.OpenFile("CivilRegistry.csv", os.O_RDONLY, 0644) // open database in readonly

	reader := csv.NewReader(f_file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("error read", err)
	}
	var status int
	status = 0 // fail to find id
	for _, record := range records {
		ID, _ := strconv.Atoi(record[0])
		if int32(ID) == id {
			status = 1

			en := gob.NewEncoder(sock)
			en.Encode(status)

			EnPerson := Person{CVname: record[1], CVid: int32(ID)}
			EnState := State{CVstate: record[2]}

			encode := gob.NewEncoder(sock)
			perr := encode.Encode(EnPerson)
			if err != nil {
				log.Panic("error to encode person ", perr)
			}
			serr := encode.Encode(EnState)
			if err != nil {
				log.Panic("error to encode state ", serr)
			}

			break
		}
	}
	if status == 0 {
		en := gob.NewEncoder(sock)
		en.Encode(status)
	}

}

func List(sock net.Conn) {
	/*
		1- set Read Lock
		2- read all data into csv file
		3- send records data
	*/
	var l_file, _ = os.OpenFile("CivilRegistry.csv", os.O_RDONLY, 0644) // open database in readonly mode

	reader := csv.NewReader(l_file)
	cv_records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("error read", err)
	}

	cv_en := gob.NewEncoder(sock)

	n := len(cv_records)
	cv_en.Encode(&n)

	for _, r := range cv_records {
		id, _ := strconv.Atoi(r[0])
		cv_p := Person{CVname: r[1], CVid: int32(id)}
		cv_s := State{CVstate: r[2]}
		cv_en.Encode(&cv_p)
		cv_en.Encode(&cv_s)
	}
}

func handelClient(sock net.Conn) {
	/*
		handel client connection
		recieve what client opertion want (insert, delete, find, list)
		call opertion function
	*/
	fmt.Println("New client connected")
	op := make([]byte, 1)
	_, err := sock.Read(op) //read operation

	if err != nil {
		log.Println(err)
	}

	switch string(op) {
	case "i":
		insert(sock)
	case "f":
		find(sock)
	case "d":
		delete(sock)
	case "l":
		List(sock)
	}
	sock.Close()
	fmt.Println("one client disconnected.")
}

func main() {
	/*	listen to client connection
		make recive / send socket
		use go rotuine to handel new connection
	*/
	psock, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Cannot open port\n")
	}
	defer psock.Close()
	for {
		sock, err := psock.Accept()
		if err != nil {
			log.Fatal("Cannot accept a new connection")
		}
		go handelClient(sock)
	}
}
