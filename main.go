package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
)

// opensearch severless client
var AOSSClient *opensearch.Client

// bedrock client
var BedrockClient *bedrockruntime.Client

// create an init function to initializing opensearch client
func init() {

	//
	fmt.Println("init and create an opensearch client")

	// load aws credentials from profile demo using config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// create a aws request signer using requestsigner
	signer, err := requestsigner.NewSignerWithService(awsCfg, "aoss")

	if err != nil {
		log.Fatal(err)
	}

	// create an opensearch client using opensearch package
	AOSSClient, err = opensearch.NewClient(opensearch.Config{
		Addresses: []string{AOSS_ENDPOINT},
		Signer:    signer,
	})

	if err != nil {
		log.Fatal(err)
	}

	// create bedorck runtime client
	BedrockClient = bedrockruntime.NewFromConfig(awsCfg)

}

func main() {

	// create handler multiplexer
	mux := http.NewServeMux()

	// hello message
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte("Hello"))
		http.ServeFile(w, r, "./static/claude-haiku.html")
	})

	// response simple json
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			Message string `json:"Message"`
		}{Message: "Hello"})
	})

	// handle aoss frontend
	mux.HandleFunc("/aoss", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			name := r.FormValue("name")
			w.Write([]byte(fmt.Sprintf("Hello %s", name)))
		}

		if r.Method == "GET" {
			// w.Write([]byte("Hello"))
			http.ServeFile(w, r, "./static/opensearch.html")
		}
	})

	// handle query to aoss
	mux.HandleFunc("/query", HandleAOSSQuery)

	// handle chat with bedrock
	mux.HandleFunc("/bedrock-stream", HandleBedrockClaude2Chat)

	// bedrock chat frontend
	mux.HandleFunc("/claude2", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/claude2.html")
	})

	// handle chat with bedrock
	mux.HandleFunc("/bedrock-haiku", HandleBedrockClaude3HaikuChat)

	// bedrock chat frontend
	mux.HandleFunc("/haiku", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/claude-haiku.html")
	})

	// bedrock frontend for image analyzer 
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/image.html")
	})

	// bedrock backend to analyze image 
	mux.HandleFunc("/claude-haiku-image", HandleHaikuImageAnalyzer)

	// create a http server using http
	server := http.Server{
		Addr:           ":3000",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server.ListenAndServe()

}
