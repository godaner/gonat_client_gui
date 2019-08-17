package conn

import (
	"encoding/binary"
	"github.com/soyum2222/slog"
	"gonat_ui/client/config"
	"gonat_ui/interface"
	"gonat_ui/proto"
	"io"
	"net"
	"sync"
	"time"
)

type remote_conversation struct {
	crypto_handler          _interface.Safe
	remote_conn             net.Conn
	server_conversation_map map[uint32]_interface.Conversation
	close_chan              chan struct{}
	close_mu                sync.Mutex
}

func (rc *remote_conversation) Monitor() {
	l := make([]byte, 4, 4)
	p := proto.Proto{}

	for {

		select {

		case <-rc.close_chan:
			return

		default:

			_, err := io.ReadFull(rc.remote_conn, l)
			if err != nil {
				slog.Logger.Error(err)
				rc.Close()
				time.Sleep(time.Second * 2)
				return
			}

			data_len := binary.BigEndian.Uint32(l)

			data := make([]byte, data_len, data_len)

			_, err = io.ReadFull(rc.remote_conn, data)
			if err != nil {
				slog.Logger.Error(err)
				rc.Close()
				return
			}

			p.Unmarshal(data, rc.crypto_handler)

			switch p.Kind {

			case proto.TCP_CREATE_CONN:
				server_con, err := net.Dial("tcp", config.Server_ip)
				if err != nil {
					slog.Logger.Error(err)
					p.Kind = proto.TCP_DIAL_ERROR
					data := p.Marshal(rc.crypto_handler)
					rc.Send(data)
					rc.remote_conn.Close()
					close(rc.close_chan)
					return

				}
				sc := server_conversation{}
				sc.server_conn = server_con
				sc.remote_conn = rc.remote_conn
				sc.close_chan = make(chan struct{}, 1)
				sc.id = p.ConversationID
				sc.crypto_handler = rc.crypto_handler
				go sc.Monitor()
				rc.server_conversation_map[p.ConversationID] = &sc

			case proto.TCP_COMM:
				err := rc.server_conversation_map[p.ConversationID].Send(p.Body)
				if err != nil {
					slog.Logger.Error(err)
					rc.server_conversation_map[p.ConversationID].Close()
					continue
				}

				slog.Logger.Debug("send server len:", len(p.Body))
				slog.Logger.Debug("send server :", string(p.Body))

			case proto.TCP_SEND_PROTO:
				slog.Logger.Info("remote port :", string(p.Body))

			case proto.TCP_PORT_BIND_ERROR:
				slog.Logger.Info("remote port already bound please replace remote_port value")

			}
		}

	}

}

func (rc *remote_conversation) Close() {
	for _, v := range rc.server_conversation_map {
		v.Close()
	}
	rc.close_mu.Lock()
	defer rc.close_mu.Unlock()
	rc.remote_conn.Close()
	close(rc.close_chan)
}

func (rc *remote_conversation) Send(b []byte) error {
	_, err := rc.remote_conn.Write(b)
	return err
}
