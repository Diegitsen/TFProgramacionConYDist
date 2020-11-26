package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
	"./RF"
    "net/http"
	"strings"
	"io/ioutil"
)

var direccion_nodo string
var yeah string

const (
	puerto_registro  = 8000
	puerto_notifica  = 8001
	puerto_proceso   = 8002
	puerto_solicitud = 8003
)

var bitacora = []string{"localhost:9002", "localhost:9000"}

type Info struct {
	Tipo     string
	NodeNum  int
	NodeAddr string
}
type MyInfo struct {
	contMsg  int
	first    bool
	nextNum  int
	nextAddr string
}

var puedeIniciar chan bool
var chMiInfo chan MyInfo

var ticket int

func main() {
	direccion_nodo = "localhost:9001"
	yeah = "0"

	rand.Seed(time.Now().UTC().UnixNano())
	ticket = rand.Intn(1000000)
	fmt.Printf("Nro Ticket %d\n", ticket)

	//crear los canales
	puedeIniciar = make(chan bool)
	chMiInfo = make(chan MyInfo)
	//enviar la solicitud inicial
	go func() {
		chMiInfo <- MyInfo{0, true, 1000001, ""}
	}()
	//esperar el inicio de la solicitud
	go func() {
		fmt.Println("Presiona enter para Iniciar la solicitud...")
		bufferIn := bufio.NewReader(os.Stdin)
		msg, _ := bufferIn.ReadString('\n')
		fmt.Println(msg)
		info := Info{"SENDNUM", ticket, direccion_nodo}
		//enviar a todos los nodos de la red
		fmt.Println("paso enviar")
		for _, direccion := range bitacora {
			go enviarSolicitud(direccion, info)
		}
	}()

	escucharSolicitud()
	
}


func escucharSolicitud() {
	//hostname := fmt.Sprintf("%s:%d", direccion_nodo, puerto_solicitud)
	fmt.Println("-----escuchando solicitud------")
	ln, _ := net.Listen("tcp", direccion_nodo)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go manejadorSolicitudes(conn)
	}
}

func manejadorSolicitudes(conn net.Conn) {
	defer conn.Close()
	bufferIn := bufio.NewReader(conn)
	msg, _ := bufferIn.ReadString('\n')
	//leer
	fmt.Println("-----manejando solicitud------")
	var info Info
	json.Unmarshal([]byte(msg), &info)
	fmt.Println("-----imprimiendo info------")
	fmt.Println(info)
	fmt.Println("-----------------")
	switch info.Tipo {
	case "SENDNUM":
		myInfo := <-chMiInfo
		if info.NodeNum < ticket {
			myInfo.first = false
		} else if info.NodeNum < myInfo.nextNum {
			myInfo.nextNum = info.NodeNum
			myInfo.nextAddr = info.NodeAddr
		}
		myInfo.contMsg++
		go func() {
			chMiInfo <- myInfo
		}()
		//autorizacion
		if myInfo.contMsg == len(bitacora) {
			if myInfo.first {
				//atender la seccion critica
				atenderSeccionCritica()
			} else {
				puedeIniciar <- true
			}
		}
	case "START":
		<-puedeIniciar
		//atender seccion critica
		atenderSeccionCritica()
	}
}

func enviarSolicitud(direccion string, msg Info) {
	//envìa la info a todos los nodos
	//remoteHost := strings.TrimSpace(direccion)
	//remoteHost = fmt.Sprintf("%s:%d", remoteHost, puerto_solicitud)
	fmt.Println("-----enviando solicitud------")
	conn, _ := net.Dial("tcp", direccion)
	defer conn.Close()
	bMsg, _ := json.Marshal(msg)
	fmt.Fprintln(conn, string(bMsg))
}
func atenderSeccionCritica() {
	fmt.Println("Iniciando trabajo en seccion critica")
	myInfo := <-chMiInfo
	if myInfo.nextAddr == "" {
		fmt.Println("Soy el proceso unico")
	} else {
		fmt.Println("Finalizando trabajo en la seccion critica")
		executeProcess()
		fmt.Printf("Siguiente a procesar es %s con el ticket %d\n", myInfo.nextAddr, myInfo.nextNum)
		msg := Info{Tipo: "START"}
		enviarSolicitud(myInfo.nextAddr, msg) // se esta comunicando al nodo su orden de acceso a la SC
	}
}

func executeProcess(){
	if(direccion_nodo == "localhost:9000"){ 
		yeah = "1"
	}else if(direccion_nodo == "localhost:9001"){
		yeah = "2"
	}else{
		yeah = "3"
	}
	fmt.Println("see " , yeah)
	resp, err := http.Get("https://raw.githubusercontent.com/Diegitsen/progr_conc_ta2_data/master/encuesta_2_variacion.data")
    if err != nil {
        print(err)
	}
	defer resp.Body.Close()
	content,_ := ioutil.ReadAll(resp.Body)
	s_content := string(content)
	lines := strings.Split(s_content,"\n")
	inputs := make([][]interface{},0)

	targets := make([]string,0)
	for _,line := range lines{

		line = strings.TrimRight(line,"\r\n")
		if len(line)==0{
			continue
		}
		tup := strings.Split(line,",")

		pattern := tup[:len(tup)-1]

		target := tup[len(tup)-1]

		X := make([]interface{},0)
		for _,x := range pattern{
			X = append(X,x)	
		}

		inputs = append(inputs,X)
	
		targets = append(targets,target)

	}

	train_inputs := make([][]interface{},0)
	train_targets := make([]string,0)

	test_inputs := make([][]interface{},0)
	test_targets := make([]string,0)

	for i,x := range inputs{
		if i%8==1{
			train_inputs = append(test_inputs, x)
		}else{
			train_inputs = append(train_inputs, x)
		}
	}

	for i,y := range targets{
		if i%8==1{
			train_targets = append(test_targets,y)
		}else{
			train_targets = append(train_targets,y)
		}
	}

	atup := [5]string{"joven", "secundaria", "FREPAP", "JUNIN", "1"}
	//atup := "[joven secundaria FREPAP JUNIN 1]"
	apattern  := atup[:len(atup)-1]
	atarget  := atup[len(atup)-1]

	aX := make([]interface{},0)
	for _,ax := range apattern{
		aX = append(aX,ax)	
	}

	ainputs := make([][]interface{},0)
	atargets := make([]string,0)
	
	ainputs = append(ainputs,aX)
	atargets = append(atargets,atarget)



	forest := RF.BuildForest(inputs,targets,10,2500,len(train_inputs[0]))

	test_inputs = ainputs
	test_targets = atargets
	fmt.Println("input ", ainputs)
	fmt.Println("targets ", atargets)
	true_positive := 0.0
	false_positive := 0.0
	true_negative := 0.0
	false_negative := 0.0

	precision := 0.0
	recall := 0.0
	f1 := 0.0
	accurancy := 0.0

	for i:=0;i<len(test_inputs);i++{
		output := forest.Predicate(test_inputs[i])
		expect := test_targets[i]
		fmt.Printf("Se predijo: ", output)
		if output == "1" && expect == "1"{
			true_positive += 1
		}else if output == "1" && expect == "0"{
			false_negative += 1
		}else if output == "0" && expect == "0"{
			true_negative += 1
		}else if output == "0" && expect == "1"{
			false_positive += 1
		}
	}
	precision = true_positive/(true_positive+false_positive)
	recall = true_positive/(true_positive+false_negative)
	f1 = 2*((precision*recall)/(precision+recall))
	accurancy = (true_positive + true_negative)/(true_positive + true_negative + false_positive + false_negative)
	fmt.Println("\nprecision ", precision)
	fmt.Println("recall",recall)
	fmt.Println("f1 ", f1)
	fmt.Println("accurancy", accurancy)
	fmt.Println("\nfin")
}