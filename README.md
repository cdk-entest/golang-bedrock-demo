---
title: getting started with amazon bedrock in go
author: haimtran
date: 25/03/2024
---

## Introduction

This repo shows how to get started with Amazon Bedrock in golang through basic examples.

- simple chat and prompt
- query vector database (opensearch)
- simple image analyzing

For learning purpose, it implement these features using only basic concepts and without relying on framework like LangChain, Streamlit, or React.

- basic stream response
- basic css and javascript

## WebServer

Project structure

```go
|--image
   |--demo.jpeg
|--static
   |--claude-haiku.html
   |--claude2.html
   |--image.html
   |--opensearch.html
|--aoss.go
|--bedrock.go
|--constants.go
|--go.mod
|--main.go
```

main.go implement a http server and route request to handlers. bedrock.go and aoss.go are functions to invoke Amazon Bedrock and Amazon OpenSearch Serverless (AOSS), respecitively. static folder contains simple frontend with javascript.

> [!IMPORTANT]  
> To use AOSS, you need create a OpenSearch collection and provide its URL endpoint in constants.go. In addition, you need to setup data access in the AOSS for the running time environment (EC2 profile, ECS taks role, Lambda role, .etc)

## Stream Response

First it is good to create some data structs according to [Amazon Bedrock Claude3 API format]()

```go
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Body struct {
	MaxTokensToSample int       `json:"max_tokens"`
	Temperature       float64   `json:"temperature,omitempty"`
	AnthropicVersion  string    `json:"anthropic_version"`
	Messages          []Message `json:"messages"`
}

// list of messages
messages := []Message{{
	Role:    "user",
	Content: []Content{{Type: "text", Text: promt}},
}}

// form request body
payload := Body{
	MaxTokensToSample: 2048,
	Temperature:       0.9,
	AnthropicVersion:  "bedrock-2023-05-31",
	Messages:          messages,
}
```

Then convert the payload to bytes and invoke Bedrock client

```go
payload := Body{
	MaxTokensToSample: 2048,
	Temperature:       0.9,
	AnthropicVersion:  "bedrock-2023-05-31",
	Messages:          messages,
}

// marshal payload to bytes
payloadBytes, err := json.Marshal(payload)

if err != nil {
	fmt.Println(err)
	return
}

// create request to bedrock
output, error := BedrockClient.InvokeModelWithResponseStream(
	context.Background(),
	&bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	},
)

if error != nil {
	fmt.Println(error)
	return
}
```

Finally, parse the streaming reponse and decode to text. When deploy on a http server, we need to modify the code a bit to stream each chunk of response to client. For example [HERE]()

```go
output, error := BedrockClient.InvokeModelWithResponseStream(
	context.Background(),
	&bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	},
)

if error != nil {
	fmt.Println(error)
	return
}

// parse response stream
for event := range output.GetStream().Events() {
	switch v := event.(type) {
	case *types.ResponseStreamMemberChunk:

		//fmt.Println("payload", string(v.Value.Bytes))

		var resp ResponseClaude3
		err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp.Delta.Text)

	case *types.UnknownUnionMember:
		fmt.Println("unknown tag:", v.Tag)

	default:
		fmt.Println("union is nil or unknown type")
	}
}
```

## Image Analyze

Similarly, for image analyzing using Amazon Bedrock Claude3, we need to create a correct request format. It is possible without explicitly define structs as above and using interface{}

```go
// read image from local file
imageData, error := ioutil.ReadFile("demo.jpeg")

if error != nil {
	fmt.Println(error)
}

// encode image to base64
base64Image := base64.StdEncoding.EncodeToString(imageData)

source := map[string]interface{}{
		"type":       "base64",
		"media_type": "image/jpeg",
		"data":       base64Image,
	}

messages := []map[string]interface{}{{
	"role":    "user",
	"content": []map[string]interface{}{{"type": "image", "source": source}, {"type": "text", "text": "what is in this image?"}},
}}

payload := map[string]interface{}{
	"max_tokens":        2048,
	"anthropic_version": "bedrock-2023-05-31",
	"temperature":       0.9,
	"messages":          messages,
}
```

Then invoke Amazon Bedrock Client like below, and similar for streaming reponse as previous example.

```go
// convert payload struct to bytes
payloadBytes, error := json.Marshal(payload)

if error != nil {
	fmt.Println(error)
}

// invoke bedrock claude3 haiku
output, error := BedrockClient.InvokeModel(
	context.Background(),
	&bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	},
)

if error != nil {
	fmt.Println(error)
}

// response
fmt.Println(string(output.Body))
```
