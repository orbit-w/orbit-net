package transport

import "github.com/orbit-w/golib/bases/packet"

/*
   @Author: orbit-w
   @File: framer
   @2023 11月 周日 12:59
*/

type Framer struct{}

type Frame struct {
	Type     int8
	End      bool
	StreamId int64
	Data     packet.IPacket
}

func (f *Framer) Encode(frame *Frame) packet.IPacket {
	w := packet.Writer()
	w.WriteInt8(frame.Type)
	w.WriteBool(frame.End)
	w.WriteInt64(frame.StreamId)
	if data := frame.Data; data != nil {
		frame.Data = nil
		w.Write(data.Remain())
		data.Return()
	}
	return w
}

func (f *Framer) Decode(data packet.IPacket) (Frame, error) {
	defer data.Return()
	frame := Frame{}
	ft, err := data.ReadInt8()
	if err != nil {
		return frame, err
	}

	end, err := data.ReadBool()
	if err != nil {
		return frame, err
	}

	sId, err := data.ReadUint64()
	if err != nil {
		return frame, err
	}

	frame.StreamId = int64(sId)
	frame.Type = ft
	frame.End = end
	if len(data.Remain()) > 0 {
		reader := packet.Reader(data.Remain())
		frame.Data = reader
	}
	return frame, nil
}
