package rpc

import (
	"bytes"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
	"sparrow/network"
	"sparrow/utils"
)

type Codec struct {
}

func (c *Codec) HandleWrite(ctx network.WriteContext, message any) {
	payload := message.(*ProtoPayload)
	payloadBytes, err := proto.Marshal(payload)
	utils.Assert(err)

	total := uint32(len(payloadBytes))
	buffer := utils.ByteBufferPool.Get().(*bytes.Buffer)
	buffer.Grow(int(total))
	defer func() {
		buffer.Reset()
		utils.ByteBufferPool.Put(buffer)
	}()
	totalBytes := [4]byte{}
	binary.LittleEndian.PutUint32(totalBytes[:], total)
	buffer.Write(totalBytes[:])
	buffer.Write(payloadBytes)

	ctx.HandleWrite(buffer.Bytes())
}

func (c *Codec) HandleRead(ctx network.ReadContext, message any) {
	reader := message.(io.Reader)
	totalBytes := [4]byte{}
	err := binary.Read(reader, binary.LittleEndian, &totalBytes)
	utils.Assert(err)

	total := binary.LittleEndian.Uint32(totalBytes[:])
	buffer := make([]byte, total)
	_, err = io.ReadFull(reader, buffer)
	utils.Assert(err)

	payload := &ProtoPayload{}
	err = proto.Unmarshal(buffer, payload)
	utils.Assert(err)

	ctx.HandleRead(payload)
}
