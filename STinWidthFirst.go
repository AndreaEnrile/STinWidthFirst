/*
	Project : Broadcast by waves demo for SDI course
	Author : Guillaume Riondet
	Date : July 2021
*/

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"net"
	"path/filepath"
	//"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

var PORT string = ":30000"

type yamlConfig struct {
	ID         int    `yaml:"id"`
	Address    string `yaml:"address"`
	Neighbours []struct {
		ID         int    `yaml:"id"`
		Address    string `yaml:"address"`
		EdgeWeight int    `yaml:"edge_weight"`
	} `yaml:"neighbours"`
}

func initAndParseFileNeighbours(filename string) yamlConfig {
	fullpath, _ := filepath.Abs("./" + filename)
	yamlFile, err := ioutil.ReadFile(fullpath)

	if err != nil {
		panic(err)
	}

	var data yamlConfig

	err = yaml.Unmarshal([]byte(yamlFile), &data)
	if err != nil {
		panic(err)
	}

	return data
}
func Log(file *os.File, message string) {
	_, err := file.WriteString(message)
	if err != nil {
		panic(err)
	}
	//long++

}

func myLog(localAdress string, message string) {
	fmt.Printf("[%s] : %s\n", localAdress, message)
}

func send(nodeAddress string, neighAddress string, message string) {
	outConn, err := net.Dial("tcp", neighAddress+PORT)
	if err != nil {
		log.Fatal(err)
		return
	}
	outConn.Write([]byte(nodeAddress+message))
	outConn.Close()
}
func sendToAllNeighboursexcept(node yamlConfig, node_addr string , message string) {
	for _, neigh := range node.Neighbours {
		if neigh.Address !=node_addr{
	    go send(node.Address, neigh.Address, message)}
	}
}
func sendToAllNeighbours(node yamlConfig, message string) {
	for _, neigh := range node.Neighbours {
		go send(node.Address, neigh.Address,message)
	}
}
func server(neighboursFilePath string, isStartingPoint bool) {
	var node yamlConfig = initAndParseFileNeighbours(neighboursFilePath)
	filename := "Log-" + node.Address
	file, err := os.Create(filename)
    if err != nil {
		panic(err)
	}
	defer file.Close()
    Log(file, "Parsing done ...\n")
	Log(file, "Server starting ....\n")
	ln, err := net.Listen("tcp", node.Address+PORT)
	if err != nil {
		log.Fatal(err)
		return
	}
	var count int = 0
	var parent_addr string =""
	var fils []string 
	var non_fils []string 
	var non_term bool = true

	//myLog(node.Address, "Neighbours file parsing ...")
	//myLog(node.Address, "Done")

	Log(file, "Starting algorithm ...\n")
	if isStartingPoint {
		Log(file, "I am a proative node ...")
		Log(file, "Sending message to all neighbours...\n")
		go sendToAllNeighbours(node, "M")
	}

	for non_term == true {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n')
		conn.Close()
		remote_addr:= message[0:9]
		msg:= message[9:10]
		Log(file, "Message received : "+msg +" From " + remote_addr + "\n")
		count += 1
		
    
		if parent_addr == "" && msg == "M" {
      // if node have no parent and message is adoption request
			Log(file, "le node n'a pas un parent \n ")
			Log(file, "Sending message P to "+remote_addr+"\n")
			send(node.Address, remote_addr, "P" )
			parent_addr = remote_addr
			Log(file, "Sending message to all neighbours execpt the parent...\n")
			sendToAllNeighboursexcept(node,parent_addr, "M")
		}else if isStartingPoint && msg == "P" {
      // if node is root and message is adoption accepted
			Log(file, "\n")
			fils=append(fils,remote_addr)
		}else if isStartingPoint && msg == "R" {
      // if node is root and message is adoption refused
			Log(file, "\n")
			non_fils=append(fils,remote_addr)
		}else if isStartingPoint && msg == "M" {
      // if node is root and message is adoption request
			Log(file, "C'est le root node \n")
			Log(file, "Sending message R  to "+remote_addr + "\n")
            send(node.Address, remote_addr, "R" )
		}else if parent_addr != "" && msg == "M" {
      // if node have a parent and message is adoption request
			Log(file, "le node a déja un parent \n")
			Log(file, "Sending message R to "+remote_addr + "\n")
			send(node.Address, remote_addr, "R" )
		}else if parent_addr != "" && msg == "P" {
      // if node have a parent and message is adoption accepted
			Log(file, "\n")
			fils=append(fils,remote_addr)
		}else if parent_addr != "" && msg == "R" {
      // if node have a parent and message is adoption refused
			Log(file, "\n")
			non_fils=append(non_fils,remote_addr)
		}
    
    // Check if is root node and the termination condition is fulfilled
		if isStartingPoint && (len(non_fils)+len(fils))==len( node.Neighbours){
			non_term = false
		}
    // Check if is not root node and the termination condition is fulfilled
		if !isStartingPoint && (len(non_fils)+len(fils) + 1)==len( node.Neighbours){
			non_term = false
		}
		
	}
	fmt.Println("La liste des fils de : ",node.Address,"est",fils)
	fmt.Println("La liste des non_fils de : ",node.Address,"est",non_fils)


	
}

func main() {
	//localadress := "127.0.0.1"
	go server("Neighbours/node-2.yaml", false)
	go server("Neighbours/node-3.yaml", false)
	go server("Neighbours/node-4.yaml", false)
	go server("Neighbours/node-5.yaml", false)
	go server("Neighbours/node-6.yaml", false)
	go server("Neighbours/node-7.yaml", false)
	go server("Neighbours/node-8.yaml", false)
	time.Sleep(2 * time.Second) //Waiting all node to be ready
	server("Neighbours/node-1.yaml", true)
	time.Sleep(2 * time.Second) //Waiting all console return from nodes
}
