package main

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type anyInterface interface{}

type idData struct {
	ID anyInterface `json:"id,omitempty"`
}

type errorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonReply struct {
	idData
	Result string `json:"result"`
}

type jsonReplyErr struct {
	// "id": null - correct for wrong json
	ID    anyInterface `json:"id"`
	Error errorData    `json:"error"`
}

type jsonRequest struct {
	idData
	Method string    `json:"method"`
	Params [1]string `json:"params,omitempty"`
}

type jsonALL struct {
	idData
	Result string       `json:"result"`
	Error  *errorData   `json:"error"`
	Method string       `json:"method"`
	Params anyInterface `json:"params,omitempty"`
}

func hashMatch(token string, hash string) bool {
	shaHasher := sha512.New()
	shaHasher.Write([]byte(token))
	tokenHash := fmt.Sprintf("%x", shaHasher.Sum(nil))
	return tokenHash == hash
}

func makeRequest(method string, param string, id anyInterface) (string, bool) {
	request := jsonRequest{Method: method}
	request.ID = id
	if len(param) > 0 {
		request.Params = [1]string{param}
	}
	var result []byte
	var err error
	if result, err = json.Marshal(request); err != nil {
		fmt.Printf("Internal makeRequest error: %v\n", err)
		return "", false
	}
	return string(result), id != nil

}

func makeError(code int, message string, id anyInterface) (string, bool) {
	requestError := jsonReplyErr{Error: errorData{code, message}}
	requestError.ID = id
	var result []byte
	var err error
	if result, err = json.Marshal(requestError); err != nil {
		fmt.Printf("Internal makeError error: %v\n", err)
		return "", false
	}
	return string(result), id != nil
}

func makeReply(result string, id anyInterface) (string, bool) {
	reply := jsonReply{Result: result}
	reply.ID = id
	var resulT []byte
	var err error
	if resulT, err = json.Marshal(reply); err != nil {
		fmt.Printf("Internal makeReply error: %v\n", err)
		return "", false
	}
	return string(resulT), id != nil
}

func parseAny(data string) (jsonALL, error) {
	response := jsonALL{}
	return response, json.Unmarshal([]byte(data), &response)
}

func makePingRequest() (string, bool) {
	return makeRequest("ping", fmt.Sprintf("%f", timeTime()), "pong")
}

// Parser class
type Parser struct {
	token string
	stage uint
}

// Parse parser
func (parser *Parser) Parse(line string) (string, bool) {
	fmt.Printf("recv <- %s\n", line)
	if parser.stage > 2 {
		return "", false
	}
	if any, err := parseAny(line); err != nil {
		message := fmt.Sprintf("Wrong JSON: %v", err)
		fmt.Println(message)
		reply, _ := makeError(-32700, message, nil)
		return reply, len(reply) > 2
	} else if any.Error != nil {
		handleError(any.Error, any.ID)
	} else if any.Result != "" {
		if parser.stage == 2 {
			handleResult(any.Result, any.ID)
		} else {
			return makeError(100, "forbidden: authorization is necessary", any.ID)
		}
	} else if any.Method != "" {
		param := ""
		if list, ok := any.Params.([]interface{}); ok && len(list) > 0 {
			if str, ok := list[0].(string); ok {
				param = str
			}
		}
		return parser.handleMethod(any.Method, param, any.ID)
	} else {
		fmt.Print("Broken JSON-RPC: ")
		fmt.Println(any)
	}
	return "", false
}

func (parser *Parser) makeWrongMethodReply(wait string, method string, id anyInterface) (string, bool) {
	msg := fmt.Sprintf("Wrong method \"%s\", i wait \"%s\" in stage %d", method, wait, parser.stage)
	return makeError(-32600, msg, id)
}

func (parser *Parser) handleMethod(method string, param string, id anyInterface) (string, bool) {
	switch parser.stage {
	case 0:
		if method != "authorization" {
			return parser.makeWrongMethodReply("authorization", param, id)
		}
		var reply string
		var ok bool
		if parser.token == "" || hashMatch(parser.token, param) {
			parser.stage++
			reply, ok = makeReply("authorized", id)
		} else {
			reply, ok = makeError(102, "forbidden: wrong hash", id)
		}
		return reply, ok
	case 1:
		if method != "upgrade duplex" {
			return parser.makeWrongMethodReply("upgrade duplex", param, id)
		}
		parser.stage++
		return makeReply("upgraded", id)
	case 2:
		if method == "ping" {
			return makeReply(param, id)
		}
	}
	return "", false
}

func handleError(data *errorData, id anyInterface) {
	fmt.Printf("Error recive: %d:%s, id: %v", data.Code, data.Message, id)
}

func handleResult(result string, anyID anyInterface) {
	if id, ok := anyID.(string); !ok {
		return
	} else if id == "pong" {
		currentTime := timeTime()
		if oldTime, err := strconv.ParseFloat(result, 64); err == nil {
			// считаем пинг
			diff := (currentTime - oldTime) * 1000
			var diffStr string
			if diff < 2 {
				diffStr = fmt.Sprintf("%.2f", diff)
			} else {
				diffStr = fmt.Sprintf("%.0f", diff)
			}
			fmt.Printf("ping %s ms\n", diffStr)
		}
	}
}

func timeTime() float64 {
	return float64(time.Now().UnixNano()) / 1000000000
}
