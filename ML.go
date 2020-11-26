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
	"io"
	"log"
)

var direccion_nodo string
var persons [][]interface{}
var prediccion string
var personss []Person

const (
	puerto_registro  = 8000
	puerto_notifica  = 8001
	puerto_proceso   = 8002
	puerto_solicitud = 8003
)

///

func handleResquest() {
	http.HandleFunc("/prediction", machineLearningProcess2)
	http.HandleFunc("/list", machineLearningList)
	http.HandleFunc("/persons3", paralelism)
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func personList(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")
	//serializacion
	jsonBytes, _ := json.MarshalIndent(personss, "", " ")
	
	io.WriteString(res, string(jsonBytes))
}

type Person struct {
	Age string `json:"age"`
	Sex string `json:"sex"`
	Resting_Blood_Pressure string `json:"blood_Pressure"`
	Serum_Cholestrol string `json:"cholestrol"`
}

func paralelism(res http.ResponseWriter, req *http.Request){
	res.Header().Set("Content-Type", "application/json")
	direccion_nodo = "localhost:9002"

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


///

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


	personss = []Person{
		{"21", "1", "1", "150"},
		{"19", "0", "4", "120"},
		{"45", "1", "2", "134"},}

	handleResquest()
	
	
	/*direccion_nodo = "localhost:9001"

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

	escucharSolicitud()*/

	
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
		fmt.Println("---REQUESTS---")
		fmt.Println("Prediccion: ", prediccion)
		fmt.Println("Lista de personas ", persons)
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
		fmt.Println("\n---------MACHINE LEARNING------------")
		machineLearningProcess()
	}else if(direccion_nodo == "localhost:9001"){
		fmt.Println("\n---------LIST------------")
		getPersons()
	}else{
		fmt.Println("\n---------MACHINE LEARNING------------")
		machineLearningProcess()
	}
	
}


func machineLearningProcess2(res http.ResponseWriter, req *http.Request){

	fmt.Println("Ingreso!!")
	//age, sex, blood, cholesterol, depression
	res.Header().Set("Content-Type", "application/json")

	age := req.FormValue("age") 
	//age1, _ := strconv.Atoi(age) 
	sex := req.FormValue("sex") 
	//sex1, _ := strconv.Atoi(sex) 
	blood := req.FormValue("blood") 
	//blood1, _ := strconv.Atoi(blood) 
	cholesterol := req.FormValue("cholesterol") 
	//cholesterol1, _ := strconv.Atoi(cholesterol) 
	depression := req.FormValue("depression") 
	//depression1, _ := strconv.Atoi(depression) 


	resp, err := http.Get("https://github.com/Diegitsen/TFProgramacionConYDist/raw/main/cleveland_dataset.data.csv")    
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

	//67,1,4,   160,        286,              0,2,108,1,1.5,2,3,3,1	
	//Age	Sex	Chest_pain	Resting_Blood_Pressure	Serum_Cholestrol	Fasting_Blood_Sugar	Resting_ECG	Max_heart_rate_achieved	Exercise_induced_angina	ST_depression	Peak_exercise	Number_of_major_vessels	Thal	Diagnosis_of_heart_disease
	//age, sex, blood, cholesterol, depression
	atup := [14]string{age, sex, "4","160", blood, cholesterol,"2","108","1","1",depression,"3","3","1"}
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
	var prediction string

	for i:=0;i<len(test_inputs);i++{
		output := forest.Predicate(test_inputs[i])
		expect := test_targets[i]
		fmt.Printf("Se predijo: ", output)
		prediction = output
		if(output == "1"){
			prediccion = "true"
		}else{
			prediccion = "false"
		}
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
	//json := [14]string{age, sex, "4","160", blood, cholesterol,"2","108","1","1",depression,"3","3","1"}
	fmt.Println("\nprecision ", precision)
	fmt.Println("recall",recall)
	fmt.Println("f1 ", f1)
	fmt.Println("accurancy", accurancy)
	fmt.Println("\nfin")

	patients := []Patient{
		{age, sex, blood, cholesterol, depression,prediction}}

	jsonBytes, _ := json.MarshalIndent(patients, "", " ")
	io.WriteString(res, string(jsonBytes))
}

func machineLearningList(res http.ResponseWriter, req *http.Request){

	fmt.Println("Ingreso!!")
	//age, sex, blood, cholesterol, depression
	res.Header().Set("Content-Type", "application/json")

	resp, err := http.Get("https://github.com/Diegitsen/TFProgramacionConYDist/raw/main/cleveland_dataset.data.csv")    
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
			test_inputs = append(test_inputs, x)
		}else{
			train_inputs = append(train_inputs, x)
		}
	}

	for i,y := range targets{
		if i%8==1{
			test_targets = append(test_targets,y)
		}else{
			train_targets = append(train_targets,y)
		}
	}





	forest := RF.BuildForest(inputs,targets,10,2500,len(train_inputs[0]))

	test_inputs = test_inputs
	test_targets = test_targets
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
		//fmt.Printf("Se predijo: ", output)
		if(output == "1"){
			prediccion = "true"
		}else{
			prediccion = "false"
		}
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
	//json := [14]string{age, sex, "4","160", blood, cholesterol,"2","108","1","1",depression,"3","3","1"}
	fmt.Println("\nprecision ", precision)
	fmt.Println("recall",recall)
	fmt.Println("f1 ", f1)
	fmt.Println("accurancy", accurancy)
	fmt.Println("\nfin")

	//patients := []Patient{
	//	{age, sex, blood, cholesterol, depression,prediction}}

    jsonBytes, _ := json.MarshalIndent(test_inputs, "", " ")
    io.WriteString(res, string(jsonBytes))
}

type Patient struct {
	Age                  string `json:"age"`
	Sex                  string `json:"sex"`
	RestingBloodPressure string `json:"blood_Pressure"`
	SerumCholestrol      string `json:"cholestrol"`
	Depression      string `json:"depression"`
	Prediction      string `json:"prediction"`
}

func machineLearningProcess(){

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

	
	atup := [5]string{"jovem", "secundaria", "FREPAP", "JUNIN", "1"}
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
		if(output == "1"){
			prediccion = "true"
		}else{
			prediccion = "false"
		}
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

func getPersons(){

	resp, err := http.Get("https://github.com/Diegitsen/TFProgramacionConYDist/raw/main/cleveland_dataset.data.csv")    
    if err != nil {
        print(err)
	}
	defer resp.Body.Close()
	content,_ := ioutil.ReadAll(resp.Body)
	s_content := string(content)
	lines := strings.Split(s_content,"\n")
	//persons := make([][]interface{},0)
	for _,line := range lines{

		line = strings.TrimRight(line,"\r\n")
		if len(line)==0{
			continue
		}
		tup := strings.Split(line,",")

		X := make([]interface{},0)
		for _,x := range tup{
			X = append(X,x)	
		}

		persons = append(persons,X)
	}

	
}