package commands

import (
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type FanModeRequest struct {
	layers.BaseLayer
	Manual bool
}

func (this *FanModeRequest) LayerType() gopacket.LayerType {
	return ipmi.LayerTypeMessage
}

func (this *FanModeRequest) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(2)
	if err != nil {
		return err
	}
	bytes[0] = 0x01
	if this.Manual {
		bytes[1] = 0x00
	} else {
		bytes[1] = 0x01
	}
	return nil
}

type FanModeCommand struct {
	Req FanModeRequest
}

func (this *FanModeCommand) Name() string {
	return "Fan Mode"
}

func (this *FanModeCommand) Operation() *ipmi.Operation {
	return &ipmi.Operation{
		Function: ipmi.NetworkFunction(0x30),
		Command:  0x30,
	}
}

func (this *FanModeCommand) RemoteLUN() ipmi.LUN {
	return ipmi.LUNBMC
}

func (this *FanModeCommand) Request() gopacket.SerializableLayer {
	return &this.Req
}

func (this *FanModeCommand) Response() gopacket.DecodingLayer {
	return nil
}
