package main

import (
	"net/http"

	"github.com/micro/go-web"
	"github.com/microhq/message-web/handler"

	"golang.org/x/net/websocket"

	message "github.com/microhq/message-srv/proto/message"
)

func main() {
	service := web.NewService(web.Name("go.micro.web.message"))
	service.Handle("/", http.FileServer(http.Dir("html")))
	service.Handle("/read", http.HandlerFunc(handler.Read))
	service.Handle("/write", http.HandlerFunc(handler.Write))
	service.Handle("/stream", websocket.Handler(handler.Stream))
	service.Init()

	sclient := service.Options().Service

	handler.MessageClient = message.NewMessageService(
		"go.micro.srv.message",
		sclient.Client(),
	)

	service.Run()
}
