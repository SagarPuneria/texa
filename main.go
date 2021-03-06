package main

import (
	"crypto/md5"
	"encoding/json"
	"strings"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	ut "texa/util"

	//Import this by exec in CLI: `go get -u github.com/TexaProject/texalib`
	"github.com/TexaProject/texajson"
	"github.com/TexaProject/texalib"
	mgo "gopkg.in/mgo.v2"
)

// AIName exports form value from /welcome globally
var AIName string

// IntName exports form value from /texa globally
var IntName string

type structChatHistory struct {
	You   string
	Eliza string
}

type mongoDB struct {
	session                *mgo.Session
	collection             *mgo.Collection
	port, user, pass, host string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in rootHandler function, Error Info: ", errD)
		}
	}()
	http.Redirect(w, r, "/welcome", 301)
}

func (mdb *mongoDB) texaHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in texaHandler method, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// fmt.Printf("%+v\n", r.Form)
		fmt.Fprint(w, "<html><head><link rel=\"stylesheet\" href=\"http://localhost:3030/css/bootstrap.min.css\"><title>File Ack | TEXA Project</title></head><body>ACKNOWLEDGEMENT: Received the scores. <br /><br />Info:<br />")
		fmt.Fprint(w, "<br /><br />VISIT: /result for interrogation.")
		fmt.Fprintf(w, "<br /><br /><input type=\"button\" class=\"btn info\" onclick=\"location.href='http://localhost:3030/result';\" value=\"Visit /result\" /></body></html>")

		fmt.Println("--INTERROGATION FORM DATA--")
		IntName = r.Form.Get("IntName")
		QSA := r.Form.Get("scoreArray")
		SlabName := r.Form.Get("SlabName")
		slabSequence := r.Form.Get("slabSequence")
		chatHistory := r.Form.Get("chatHistory")

		fmt.Println("###", AIName)
		fmt.Println("###", IntName)
		fmt.Println("###", QSA)
		fmt.Println("###", SlabName)
		fmt.Println("###", slabSequence)

		// LOGIC
		re := regexp.MustCompile("[0-1]+")
		array := re.FindAllString(QSA, -1)

		SlabNameArray := regexp.MustCompile("[,]").Split(SlabName, -1)
		slabSeqArray := regexp.MustCompile("[,]").Split(slabSequence, -1)
		chatHistoryArray := regexp.MustCompile("[,]").Split(chatHistory, -1)

		fmt.Println("###chatHistoryArray:", chatHistoryArray)
		for i := 0; i <= len(chatHistoryArray)-1; i = i + 2 {
			var c structChatHistory
			c.You = chatHistoryArray[i]
			j := i + 1
			if j <= len(chatHistoryArray)-1 {
				c.Eliza = chatHistoryArray[j]
			}
			err := mdb.collection.Insert(&c)
			if err != nil {
				fmt.Println("c.Insert error info:", err)
			}
		}

		fmt.Println("###Resulting Array:")
		for x := range array {
			fmt.Println(array[x])
		}

		fmt.Println("###SlabNameArray: ")
		fmt.Println(SlabNameArray)

		fmt.Println("###slabSeqArray: ")
		fmt.Println(slabSeqArray)

		ArtiQSA := texalib.Convert(array)
		fmt.Println("###ArtiQSA:")
		fmt.Println(ArtiQSA)

		HumanQSA := texalib.SetHumanQSA(ArtiQSA)
		fmt.Println("###HumanQSA:")
		fmt.Println(HumanQSA)

		TSA := texalib.GetTransactionSeries(ArtiQSA, HumanQSA)
		fmt.Println("###TSA:")
		fmt.Println(TSA)

		ArtiMts := texalib.GetMeanTestScore(ArtiQSA)
		HumanMts := texalib.GetMeanTestScore(HumanQSA)

		fmt.Println("###ArtiMts: ", ArtiMts)
		fmt.Println("###HumanMts: ", HumanMts)

		PageArray := texajson.GetPages()
		fmt.Println("###PageArray")
		fmt.Println(PageArray)
		for _, p := range PageArray {
			fmt.Println(p)
		}

		newPage := texajson.ConvtoPage(AIName, IntName, ArtiMts, HumanMts)

		PageArray = texajson.AddtoPageArray(newPage, PageArray)
		fmt.Println("###AddedPageArray")
		fmt.Println(PageArray)

		JsonPageArray := texajson.ToJson(PageArray)
		fmt.Println("###jsonPageArray:")
		fmt.Println(JsonPageArray)

		////
		fmt.Println("### SLAB LOGIC")

		slabPageArray := texajson.GetSlabPages()
		fmt.Println("###slabPageArray")
		fmt.Println(slabPageArray)

		slabPages := texajson.ConvtoSlabPage(ArtiQSA, SlabNameArray, slabSeqArray)
		fmt.Println("###slabPages")
		fmt.Println(slabPages)
		for z := 0; z < len(slabPages); z++ {
			slabPageArray = texajson.AddtoSlabPageArray(slabPages[z], slabPageArray)
		}
		fmt.Println("###finalslabPageArray")
		fmt.Println(slabPageArray)

		JsonSlabPageArray := texajson.SlabToJson(slabPageArray)
		fmt.Println("###JsonSlabPageArray: ")
		fmt.Println(JsonSlabPageArray)

		////
		fmt.Println("### CAT LOGIC")

		CatPageArray := texajson.GetCatPages()
		fmt.Println("###CatPageArray")
		fmt.Println(CatPageArray)

		CatPages := texajson.ConvtoCatPage(AIName, slabPageArray, SlabNameArray)
		fmt.Println("###CatPages")
		fmt.Println(CatPages)
		CatPageArray = texajson.AddtoCatPageArray(CatPages, CatPageArray)

		// for z := 0; z < len(CatPages); z++ {
		// 	CatPageArray = texajson.AddtoCatPageArray(CatPages[z], CatPageArray)
		// }
		fmt.Println("###finalCatPageArray")
		fmt.Println(CatPageArray)

		JsonCatPageArray := texajson.CatToJson(CatPageArray)
		fmt.Println("###JsonCatPageArray: ")
		fmt.Println(JsonCatPageArray)
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in welcomeHandler function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/welcome.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
	}
}

// upload logic
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in uploadHandler function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("login.html")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		handler.Filename = "elizadata.js"
		AIName = r.FormValue("AIName")
		fmt.Println(AIName)
		defer file.Close()

		fmt.Fprint(w, "<html><head><link rel=\"stylesheet\" href=\"http://localhost:3030/css/bootstrap.min.css\"><title>File Ack | TEXA Project</title></head><body>ACKNOWLEDGEMENT: Uploaded the file. <br /><br />Header Info:<br />")
		fmt.Fprintf(w, "%v", handler.Header)
		fmt.Fprintf(w, "<br /><br />Saved As: www/js/"+handler.Filename)
		fmt.Fprint(w, "<br /><br />VISIT: /texa for interrogation.")
		fmt.Fprintf(w, "<br /><br /><input type=\"button\" class=\"btn info\" onclick=\"location.href='http://localhost:3030/texa';\" value=\"Visit /texa\" /></body></html>")
		f, err := os.OpenFile("./www/js/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Selected file: ", handler.Filename)
		defer f.Close()
		io.Copy(f, file)
		// http.Redirect(w, r, "/texa", 301)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in resultHandler function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/result.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
	}
}

func (mdb *mongoDB) getchathistory(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in getchathistory method, Error Info: ", errD)
		}
	}()
	fmt.Println("getchathistory method:", r.Method)
	iter := mdb.collection.Find(nil).Iter()
	var arrayStructChatHistory []structChatHistory
	var result structChatHistory
	for iter.Next(&result) {
		arrayStructChatHistory = append(arrayStructChatHistory, result)
	}
	if err := iter.Close(); err != nil {
		fmt.Println("iter.Close error info:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		os.Exit(1)
	}

	response, err := json.Marshal(arrayStructChatHistory)
	if err != nil {
		fmt.Println("json.Marshal error info:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		os.Exit(1)
	}
	fmt.Println("###response:", string(response))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		fmt.Println("w.Write error info:", err)
		os.Exit(1)
	}
}

func getCatJSON(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in getCatJSON function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	catPages := texajson.GetCatPages()
	bs, err := json.Marshal(catPages)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func getMtsJSON(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in getMtsJSON function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	mtsPage := texajson.GetPages()
	bs, err := json.Marshal(mtsPage)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func getSlabJSON(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in getSlabJSON function, Error Info: ", errD)
		}
	}()
	fmt.Println("method:", r.Method) //get	request	method
	slabPages := texajson.GetSlabPages()
	bs, err := json.Marshal(slabPages)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func (mdb *mongoDB) initiateMongoDB() {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in initiateMongoDB method, Error Info: ", errD)
		}
	}()
	var url string
	if mdb.user == "" || mdb.pass == "" {
		url = fmt.Sprintf("mongodb://%s:%s", mdb.host, mdb.port)
	} else {
		url = fmt.Sprintf("mongodb://%s:%s@%s:%s", mdb.user, mdb.pass, mdb.host, mdb.port)
	}

	var err error
	fmt.Println("Starting connect mongoDB....")
	mdb.session, err = mgo.Dial(url)
	if err != nil {
		fmt.Println("session.DB error info::", err)
		os.Exit(1)
	}
	mdb.session.SetMode(mgo.Monotonic, true)
	mdb.collection = mdb.session.DB("texa").C("chatHistory")
}

func main() {
	defer func() {
		if errD := recover(); errD != nil {
			fmt.Println("Exception occurred at ", ut.RecoverExceptionDetails(ut.FunctionName()), " and recovered in main function, Error Info: ", errD)
		}
	}()
	mdb := &mongoDB{}
	mdb.user = strings.TrimSpace(os.Args[1])
	mdb.pass = strings.TrimSpace(os.Args[2])
	mdb.host = strings.TrimSpace(os.Args[3])
	mdb.port = strings.TrimSpace(os.Args[4])
	mdb.initiateMongoDB()
	defer mdb.session.Close()

	fmt.Println("--TEXA SERVER--")
	fmt.Println("STATUS: INITIATED")
	fmt.Println("ADDR: http://127.0.0.1:3030")
	fsc := http.FileServer(http.Dir("www/css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsc))
	fsj := http.FileServer(http.Dir("www/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsj))
	fsd := http.FileServer(http.Dir("www/data"))
	http.Handle("/data/", http.StripPrefix("/data/", fsd))

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/texa", mdb.texaHandler)
	http.HandleFunc("/texachathistory", mdb.getchathistory)
	http.HandleFunc("/result", resultHandler)
	http.HandleFunc("/cat", getCatJSON)
	http.HandleFunc("/mts", getMtsJSON)
	http.HandleFunc("/slab", getSlabJSON)

	http.ListenAndServe(":3030", nil)
}
