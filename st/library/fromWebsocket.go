package library

import (
    "github.com/nytlabs/streamtools/st/blocks" // blocks
    "github.com/gorilla/websocket"
    "encoding/json"
    "time"
    "net/http"
    "io"
    "io/ioutil"
)

// specify those channels we're going to use to communicate with streamtools
type FromWebsocket struct {
    blocks.Block
    queryrule chan chan interface{}
    inrule    chan interface{}
    inpoll    chan interface{}
    in        chan interface{}
    out       chan interface{}
    quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromWebsocket() blocks.BlockInterface {
    return &FromWebsocket{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromWebsocket) Setup() {
    b.Kind = "FromWebsocket"
    b.inrule = b.InRoute("rule")
    b.queryrule = b.QueryRoute("rule")
    b.quit = b.Quit()
    b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromWebsocket) Run() {
    var ws *websocket.Conn
    var URL string
    var handshakeDialer = &websocket.Dialer{
        Subprotocols:    []string{"p1", "p2"},
    }
    wsHeader := http.Header{"Origin": {"http://localhost/"}}

    loop := time.NewTicker(time.Millisecond * 10)
    for {
        select {
        case <-loop.C:
        case ruleI := <-b.inrule:
            var err error
            // set a parameter of the block
            r, ok := ruleI.(map[string]interface{})
            if !ok {
                b.Error("bad rule")
                break
            }

            url, ok := r["url"]
            if !ok {
                b.Error("no url specified")
                break
            }
            surl, ok := url.(string)
            if !ok {
                b.Error("error reading url")
                break
            }

            ws, _, err = handshakeDialer.Dial(surl, wsHeader)          
            if err != nil {
                b.Error("could not connect to url")
                break
            }
            ws.SetReadDeadline(time.Time{})  
            URL = surl
        case <-b.quit:
            // quit the block
            return
        case o := <-b.queryrule:
            o <- map[string]interface{}{
                "url": URL,
            }
        }
        if ws != nil {

            for {
                var r io.Reader
                var err error
                _, r, err = ws.NextReader()
                if err != nil {
                    b.Error(err)
                    break
                }
                p, err := ioutil.ReadAll(r)
                if err != nil {
                    break
                }

                var outMsg interface{}
                err = json.Unmarshal(p, &outMsg)
                if err != nil {
                    break
                }
                b.out <- outMsg
            }
        }
    }
}
