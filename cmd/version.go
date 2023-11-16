package cmd

import (
	"runtime"

	"github.com/byteplus-sdk/byteplus-go-sdk/byteplus/request"
)

var clientVersionAndUserAgentHandler = request.NamedHandler{
	Name: "ByteplusCliUserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler(clientName, clientVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH),
}

const clientName = "byteplus-cli"
const clientVersion = "1.0.0"
