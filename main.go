// SPDX-License-Identifier: CC0-1.0

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	bind   string
	opts   *mqtt.ClientOptions
	prefix string
	qos    byte
)

func init() {
	flag.StringVar(&bind, "bind", "[::1]:8080", "listen on this address")
	broker := flag.String("broker", "tcp://[::1]:1883", "mqtt server to which to connect")
	flag.StringVar(&prefix, "prefix", "/", "topic prefix to bridge")
	qos_ := flag.Int("qos", 0, "QoS for bridged messages")
	clientId := flag.String("client-id", "", "mqtt client-id")
	username := flag.String("username", "", "mqtt username")
	pwfile := flag.String("password-file", "", "file containing mqtt user password")
	flag.Parse()

	qos = byte(*qos_)

	opts = mqtt.NewClientOptions()
	opts.AddBroker(*broker)
	opts.SetConnectRetry(true)
	opts.SetClientID(*clientId)
	opts.SetUsername(*username)

	if *pwfile != "" {
		password_, err := os.ReadFile(os.ExpandEnv(*pwfile))
		if err != nil {
			log.Fatal(err)
		}
		opts.SetPassword(string(password_))
	}
}

type handler struct {
	client mqtt.Client
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "only POST and PUT are supported", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while reading body", http.StatusBadRequest)
		return
	}

	topic := path.Clean(path.Join(prefix, r.URL.Path))
	topic = strings.TrimPrefix(topic, "/")

	if topic == "" {
		http.Error(w, "got empty path without prefix", http.StatusBadRequest)
		return
	}

	mime := http.DetectContentType(body)
	fmt.Printf("[MSG] topic=\"%s\" length=%d mime=\"%s\"\n", topic, len(body), mime)

	// TODO handle publish errors; queue qos>1 ?
	h.client.Publish(topic, qos, false, body).Wait()
}

func main() {
	mqtt.ERROR = log.New(os.Stderr, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stderr, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stderr, "[WARN] ", 0)

	c := mqtt.NewClient(opts)
	c.Connect().Wait()
	defer c.Disconnect(250) // ms

	log.Fatal(http.ListenAndServe(bind, &handler{client: c}))
}
