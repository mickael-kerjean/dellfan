package commands

import (
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type FanSpeedRequest struct {
	layers.BaseLayer
	Speed byte
}

func (this *FanSpeedRequest) LayerType() gopacket.LayerType {
	return ipmi.LayerTypeMessage
}

func (this *FanSpeedRequest) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(3)
	if err != nil {
		return err
	}
	bytes[0] = 0x02
	bytes[1] = 0xff
	bytes[2] = this.Speed
	return nil
}

type FanSpeedCommand struct {
	Req FanSpeedRequest
}

func (this *FanSpeedCommand) Name() string {
	return "Fan Control"
}

func (this *FanSpeedCommand) Operation() *ipmi.Operation {
	return &ipmi.Operation{
		Function: ipmi.NetworkFunction(0x30),
		Command:  0x30,
	}
}

func (this *FanSpeedCommand) RemoteLUN() ipmi.LUN {
	return ipmi.LUNBMC
}

func (this *FanSpeedCommand) Request() gopacket.SerializableLayer {
	return &this.Req
}

func (this *FanSpeedCommand) Response() gopacket.DecodingLayer {
	return nil
}
