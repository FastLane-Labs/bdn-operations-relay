package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fasthttp/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/valyala/fastjson"
)

// ContextWithSignal returns a context that is cancelled when the process receives the given termination signal.
func ContextWithSignal(parent context.Context, s ...os.Signal) context.Context {
	if len(s) == 0 {
		s = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}
	ctx, cancel := context.WithCancel(parent)
	c := make(chan os.Signal, 1)
	signal.Notify(c, s...)

	go func() {
		// wait for either the signal, or for the context to be cancelled
		select {
		case <-c:
		case <-parent.Done():
		}

		// cancel the context
		// and stop waiting for more signals (next sigterm will terminate the process)
		cancel()
		signal.Stop(c)
	}()

	return ctx
}

func randomID() jsonrpc2.ID {
	return jsonrpc2.ID{Str: strconv.FormatUint(rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), 10), IsString: true}
}

func main() {
	userAndSolverOperations()
}

func userAndSolverOperations() {
	ctx := ContextWithSignal(context.Background())
	go func() {
		err := solver(ctx)
		if err != nil {
			panic(err)
		}
	}()

	req := &types.UserOperationWithHintsRaw{
		ChainId: (*hexutil.Big)(big.NewInt(56)),
		UserOperation: &types.UserOperationRaw{
			From:         NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
			To:           NewEthAddress("0xCbe321c620071307Ba5d0381c886B7359763735E"),
			Value:        (*hexutil.Big)(big.NewInt(2)),
			Gas:          (*hexutil.Big)(big.NewInt(3)),
			MaxFeePerGas: (*hexutil.Big)(big.NewInt(4)),
			Nonce:        (*hexutil.Big)(big.NewInt(5)),
			Deadline:     (*hexutil.Big)(big.NewInt(6)),
			Dapp:         NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
			Control:      NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
			CallConfig:   (*hexutil.Big)(big.NewInt(76)),
			SessionKey:   NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
			Data:         (hexutil.Bytes)([]byte("data")),
			Signature:    (hexutil.Bytes)([]byte("signature")),
		},
		Hints: []common.Address{
			NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
			NewEthAddress("0xCbe321c620071307Ba5d0381c886B7359763735E"),
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("failed to marshal user operation partial: %v", err)
	}

	resp, err := http.DefaultClient.Post("http://localhost:9080/userOperation", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to submit intent: %v", err)
	}

	var m map[string]string
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		log.Fatalf("failed to decode response: %v", err)
	}
	resp.Body.Close()

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		resp, err = http.DefaultClient.Get(fmt.Sprintf("http://localhost:9080/solverOperations?intent_id=%s", m["intent_id"]))
		if err != nil {
			log.Fatalf("failed to get intent solutions: %v", err)
		}

		var m2 []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&m2)
		if err != nil {
			log.Fatalf("failed to decode response: %v", err)
		}

		log.Printf("intent solutions: %v", m2)

		if len(m2) > 0 {
			break
		}
	}

	time.Sleep(10000 * time.Second)
}

func solver(ctx context.Context) error {
	addr := "localhost:9080"

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/solver"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer c.Close()

	wg := sync.WaitGroup{}
	wg.Add(3)

	readMsgChan := make(chan []byte, 1000)
	writeMsgChan := make(chan *jsonrpc2.Request, 1000)

	go func() {
		defer wg.Done()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			log.Printf("recv: %s", message)

			readMsgChan <- message
		}
	}()

	go func() {
		defer wg.Done()

		params := json.RawMessage(`{"subscription_type": "intent"}`)

		// subscribe
		req := &jsonrpc2.Request{
			Method: "subscribe",
			ID:     randomID(),
			Params: &params,
		}
		err = c.WriteJSON(req)
		if err != nil {
			log.Println("write:", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-writeMsgChan:
				err = c.WriteJSON(msg)
				if err != nil {
					log.Println("write:", err)
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case message := <-readMsgChan:
				err := handleMessage(message, writeMsgChan)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

func handleMessage(message []byte, writeMsgChan chan *jsonrpc2.Request) error {
	var p fastjson.Parser
	v, err := p.ParseBytes(message)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	sID := string(v.GetStringBytes("params", "subscription_id"))
	if sID != "" {
		return nil
	}

	intentID := string(v.GetStringBytes("params", "intentID"))
	if intentID == "" {
		return nil
	}
	intent := v.GetStringBytes("params", "intent")

	out := make([]byte, base64.StdEncoding.DecodedLen(len(intent)))
	base64.StdEncoding.Decode(out, intent)

	var uopr types.UserOperationPartialRaw
	err = json.Unmarshal(out, &uopr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal intent into user operation partial: %w", err)
	}

	op := types.SolverOperationRaw{
		From:         uopr.From,
		To:           uopr.To,
		Value:        (*hexutil.Big)(big.NewInt(10000)),
		Gas:          uopr.Gas,
		MaxFeePerGas: uopr.MaxFeePerGas,
		Deadline:     uopr.Deadline,
		Solver:       NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
		Control:      uopr.Control,
		UserOpHash:   uopr.UserOpHash,
		BidToken:     NewEthAddress("0x0087c5900b9bbc051b5f6299f5bce92383273b28"),
		BidAmount:    (*hexutil.Big)(big.NewInt(1)),
	}

	data, err := json.Marshal(op)
	if err != nil {
		return fmt.Errorf("failed to marshal solver operation: %w", err)
	}

	params := json.RawMessage(fmt.Sprintf(`{"intent_id": "%s", "intent_solution": %s}`, intentID, data))

	req := &jsonrpc2.Request{
		Method: "submitSolverOperation",
		ID:     randomID(),
		Params: &params,
	}

	writeMsgChan <- req

	return nil
}

// NewEthAddress is a one-line utility for deserializing strings into Ethereum addresses
func NewEthAddress(s string) common.Address {
	var address common.Address
	addressBytes, _ := hexutil.Decode(s)
	copy(address[:], addressBytes)
	return address
}
