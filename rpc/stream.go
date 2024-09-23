package rpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"google.golang.org/protobuf/proto"
)

func NewBidiStream(call CallType, newRecv func() proto.Message) *BidiStream {
	return &BidiStream{call: call, newRecv: newRecv}
}

type BidiStream struct {
	reader  io.Reader
	writer  io.Writer
	ready   chan struct{}
	newRecv func() proto.Message
	rspErr  Error
	rspMsg  proto.Message
	call    CallType
}

func (b *BidiStream) SetReader(r io.Reader) {
	b.reader = r
}

func (b *BidiStream) SetWriter(w io.Writer) {
	b.writer = w
}

func (b *BidiStream) SetReady(ready chan struct{}) {
	b.ready = ready
}

func (b *BidiStream) Send(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	payload := &ProtoPayload{
		Type: b.call,
		Data: data,
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return err
	}

	total := uint32(len(payloadBytes))
	totalBytes := [4]byte{}
	binary.LittleEndian.PutUint32(totalBytes[:], total)

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	buffer.Write(totalBytes[:])
	buffer.Write(payloadBytes)
	_, err = b.writer.Write(buffer.Bytes())
	return err
}

func (b *BidiStream) Recv() (proto.Message, error) {
	if b.ready != nil {
		<-b.ready
	}

	totalBytes := [4]byte{}
	err := binary.Read(b.reader, binary.LittleEndian, &totalBytes)
	if err != nil {
		return nil, err
	}

	total := binary.LittleEndian.Uint32(totalBytes[:])
	buf := make([]byte, int(total))
	_, err = io.ReadFull(b.reader, buf)
	if err != nil {
		return nil, err
	}

	payload := &ProtoPayload{}
	err = proto.Unmarshal(buf, payload)
	if err != nil {
		return nil, err
	}

	var msg proto.Message
	if len(payload.GetData()) != 0 {
		msg = b.newRecv()
		err = proto.Unmarshal(payload.GetData(), msg)
	}

	if payload.Type == CallType_Response {
		if payload.GetErrCode() == 0 {
			b.rspMsg = msg
			return nil, io.EOF
		}
		b.rspErr = NewError(payload.GetErrCode(), errors.New(payload.GetErrMsg()))
		return nil, b.rspErr
	}

	return msg, err
}

func (b *BidiStream) RecvResponse() (proto.Message, error) {
	_, err := b.Recv()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return b.rspMsg, b.rspErr
}

func (b *BidiStream) Close() {
	b.CloseReader()
	b.CloseWriter()
}

func (b *BidiStream) CloseReader() {
	if b.reader != nil {
		if closer, ok := b.reader.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

func (b *BidiStream) CloseWriter() {
	if b.writer != nil {
		if closer, ok := b.writer.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

func (b *BidiStream) SendResponse(msg proto.Message, rpcErr Error) error {
	payload := &ProtoPayload{
		Type: CallType_Response,
	}

	if msg != nil {
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		payload.Data = data
	}

	if rpcErr != nil {
		payload.ErrCode = rpcErr.Code()
		payload.ErrMsg = rpcErr.Error()
	}

	payloadBytes, perr := proto.Marshal(payload)
	if perr != nil {
		return perr
	}

	total := uint32(len(payloadBytes))
	totalBytes := [4]byte{}
	binary.LittleEndian.PutUint32(totalBytes[:], total)

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	buffer.Write(totalBytes[:])
	buffer.Write(payloadBytes)
	_, werr := b.writer.Write(buffer.Bytes())
	return werr
}
