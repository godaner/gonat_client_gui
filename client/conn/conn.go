package conn

import (
	"context"
	"encoding/binary"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/soyum2222/slog"
	"gonat_ui/client/config"
	"gonat_ui/interface"
	"gonat_ui/proto"
	"gonat_ui/safe"
	"net"
	"time"
)

func Start(stop_signal context.Context, window fyne.Window) {

	content := window.Content()
	//box_v := *content.(*widget.Box).Children[0].(*widget.Form)

	temp := make([]fyne.CanvasObject, len(content.(*widget.Box).Children))
	copy(temp, content.(*widget.Box).Children)

	defer func() {
		content.(*widget.Box).Children[0] = temp[0]
	}()
	content.(*widget.Box).Children[0] = widget.NewLabelWithStyle("connecting...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

label:
	fmt.Println(config.Remote_ip)
	remote_conn, err := net.Dial("tcp", config.Remote_ip)
	if err != nil {
		slog.Logger.Error(err)
		time.Sleep(5 * time.Second)

		select {
		case <-stop_signal.Done():
			remote_conn.Close()
		default:
			goto label
		}
	}
	content.(*widget.Box).Children[0] = widget.NewLabelWithStyle("connection succeeded", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-stop_signal.Done():
			remote_conn.Close()
			return
		}
	}()
	start_conversation(remote_conn)
	cancel()

}

func start_conversation(remote_conn net.Conn) {

	rc := remote_conversation{}
	rc.close_chan = make(chan struct{}, 1)
	rc.remote_conn = remote_conn
	rc.server_conversation_map = make(map[uint32]_interface.Conversation)
	rc.crypto_handler = safe.GetSafe(config.Crypt, config.CryptKey)

	port := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(port, uint32(config.Remote_port))
	p := proto.Proto{
		Kind:           proto.TCP_SEND_PROTO,
		ConversationID: 0,
		Body:           append([]byte("gonat_port:"), port...),
	}
	_, err := rc.remote_conn.Write(p.Marshal(rc.crypto_handler))
	if err != nil {
		slog.Logger.Error(err)
		return
	}
	go rc.Heartbeat()
	rc.Monitor()

}
