package main

import (
	"bitbucket.org/ckvist/twilio/twiml"
	"fmt"
)

type TwimlCall struct {
	CallStatus string
	DialStatus string
	CalledVia string
	Digits string

	Msg string `formam:"msg"`

	Called string
	CalledZip string
	CalledCity string
	CalledState string
	CalledCountry string

	Caller string
	CallerState string
	CallerCity string
	CallerZip string
	CallerCountry string

	ApiVersion string

	Duration int
	DialCallDuration int
	CallDuration int

	CallSid string
	CallGuid string

	AccountSid string
	AccountGuid string

	SegmentGuid string
	CallSegmentGuid string
}

func NewScreen () *twiml.Response {
	screen := twiml.NewResponse()
	screen.Gather(
		twiml.Gather{
			Action: "/complete",
		},
		twiml.Pause{
			Length: 2,
		},
		twiml.Say{
			Text: "Press any key to accept",
		})
	screen.Response = append(screen.Response, twiml.Hangup{})

	return screen
}

func NewComplete () *twiml.Response {
	complete := twiml.NewResponse()
	complete.Action(
		twiml.Say{
			Text:"Connecting",
		})

	return complete
}

func NewDial (number string) *twiml.Response {
	call := twiml.NewResponse()
	call.Dial(
		twiml.Dial{
			Action: fmt.Sprintf("/call"),
		},
		twiml.Number{
			Url:"/screen",
			Number: number,
		})

	return call
}

func NewHangup () *twiml.Response {
	hangup := twiml.NewResponse()
	hangup.Action(
		twiml.Hangup{
		})

	return hangup
}

func NewNoNumbers () *twiml.Response {
	hangup := twiml.NewResponse()
	hangup.Action(
		twiml.Say{Text: "No available coordinators"})
	hangup.Response = append(hangup.Response, twiml.Hangup{})

	return hangup
}
