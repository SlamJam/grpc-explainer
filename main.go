package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/k0kubun/pp"
	"github.com/protocolbuffers/protoscope"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func compressBtyToString(compress byte) string {
	switch compress {
	case 0:
		return "no"
	case 1:
		return "yes"
	default:
		return fmt.Sprintf("unknown(0x%X)", compress)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("[%v] Request handled at path: %v", r.Proto, r.URL.Path)

	logrus.Debug("Headers:\n", pp.Sprint(r.Header))

	grpcHeader := make([]byte, 5)

	_, err := io.ReadFull(r.Body, grpcHeader)
	if err != nil {
		logrus.Fatalln(err)
	}

	compress := grpcHeader[0]
	length := binary.BigEndian.Uint32(grpcHeader[1:5])

	logrus.Debugf("gRPC with compess: %s, length: %v", compressBtyToString(compress), length)

	rawRequest := make([]byte, length)
	_, err = io.ReadFull(r.Body, rawRequest)
	if err != nil {
		logrus.Fatalln(err)
	}

	outBytes := (protoscope.Write(rawRequest, protoscope.WriterOptions{
		// NoQuotedStrings:        true,
		// AllFieldsAreMessages:   true,
		ExplicitWireTypes:      true,
		ExplicitLengthPrefixes: true,
	}))

	logrus.Debug("Message in protoscope:\n", outBytes)

	// Возвращаем статус Canceled
	w.Header().Add("grpc-status", "1")
	// Возвращем кастомное сообщение об ошибке
	w.Header().Add("grpc-message", "Thank you for your visit!")
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	h2s := &http2.Server{}

	h1s := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(http.HandlerFunc(handler), h2s),
	}

	fmt.Println("Ready to start")
	log.Fatal(h1s.ListenAndServe())
}
