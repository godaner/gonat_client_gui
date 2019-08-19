package _interface

type Conversation interface {
	Monitor()
	Close()
	Heartbeat()
	Send([]byte) error
}

type Safe interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}
