package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/pborman/uuid"
	"github.com/yosssi/ace"
	"golang.org/x/net/context"

	"golang.org/x/net/websocket"
	//"github.com/gorilla/mux"
	message "github.com/microhq/message-srv/proto/message"
)

var (
	templateDir = "templates"
	opts        *ace.Options

	MessageClient message.MessageService
	Namespace     = "default"
)

func init() {
	opts = ace.InitializeOptions(nil)
	opts.BaseDir = templateDir
	opts.DynamicReload = true
	opts.FuncMap = template.FuncMap{
		"TimeAgo": func(t int64) string {
			return timeAgo(t)
		},
		"Colour": func(s string) string {
			return colour(s)
		},
	}
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	basePath := hostPath(r)

	opts.FuncMap["URL"] = func(path string) string {
		return filepath.Join(basePath, path)
	}

	tpl, err := ace.Load("layout", tmpl, opts)
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	if err := tpl.Execute(w, data); err != nil {
		http.Redirect(w, r, "/", 302)
	}
}

func Read(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	channel := r.Form.Get("channel")

	if len(channel) == 0 {
		channel = "default"
	}

	rsp, err := MessageClient.Search(context.TODO(), &message.SearchRequest{
		Namespace: Namespace,
		Reverse:   true,
		// select channel based on
		Channel: channel,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sort.Sort(sortedEvents{rsp.Events})

	b, err := json.Marshal(rsp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprint(w, string(b))
}

func Write(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	channel := r.Form.Get("channel")
	from := r.Form.Get("from")
	text := r.Form.Get("text")

	if len(channel) == 0 {
		channel = "default"
	}

	if len(text) == 0 {
		http.Error(w, "text required", 500)
		return
	}

	_, err := MessageClient.Create(context.TODO(), &message.CreateRequest{
		Event: &message.Event{
			Id:        uuid.NewUUID().String(),
			Namespace: Namespace,
			Channel:   channel,
			Text:      text,
			From:      from,
			Type:      "message",
		},
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprint(w, `{}`)
}

func Stream(ws *websocket.Conn) {
	var m map[string]interface{}

	if err := websocket.JSON.Receive(ws, &m); err != nil {
		fmt.Println("err", err)
		return
	}

	stream, err := MessageClient.Stream(context.TODO(), &message.StreamRequest{
		Namespace: Namespace,
		Channel:   m["channel"].(string),
	})
	if err != nil {
		fmt.Println("err", err)
		return
	}
	defer stream.Close()

	for {
		msg, err := stream.Recv()
		if err != nil {
			fmt.Println("err", err)
			return
		}

		if err := websocket.JSON.Send(ws, msg.Event); err != nil {
			fmt.Println("err", err)
			return
		}
	}
}
