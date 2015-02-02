package main

import (
	"bitbucket.org/ckvist/twilio/twiml"
	"fmt"
	"os"
	"net/http"
	"github.com/op/go-logging"
	//"github.com/ajg/form"
	"github.com/monoculum/formam"
)

var log = logging.MustGetLogger("cf")
var format = logging.MustStringFormatter(
    "%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

func NewScreen () *twiml.Response {
	screen := twiml.NewResponse()
	screen.Gather(
		twiml.Gather{
			Action: "/complete",
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

type CallForward struct {
	Numbers []string
	ActiveCalls map[string]bool
	next_caller map[string]int
}

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

type Call struct {
	TwimlCall
	Active bool
	cf * CallForward

}

func (c * Call) CallCompleted () bool {
	return c.CallStatus == "completed"
}

func (c * Call) InProgress () bool {
	return c.Active
}

func (c * Call) End() {
	c.Active = false
	c.cf.next_caller[c.CallSid] = 0
	c.cf.ActiveCalls[c.CallSid] = false
}

func (c * Call) DialNext() bool {
	return c.CallCompleted() == false && c.DialCallDuration == 0
}

func (c * Call) NewCaller() bool {
	return c.CallCompleted() == false && c.InProgress() == false
}

func (c * Call) EndOfCall() bool {
	 return (c.CallCompleted() || c.DialCallDuration != 0) && c.InProgress() == true
}

func (c * Call) GetNextNumber() string {
	next := c.cf.Numbers[c.cf.next_caller[c.CallSid]]
	c.cf.next_caller[c.CallSid] += 1
	return next
}

func (cf * CallForward) SetActive(c * Call) {
	cf.ActiveCalls[c.CallSid] = true
	cf.next_caller[c.CallSid] = 0
	c.Active = true
}

func (cf * CallForward) DialNext(c * Call) bool {
	return c.DialNext() == true && cf.next_caller[c.CallSid] < len(cf.Numbers)
}

func (c * Call) EndOfList() bool {
	return c.cf.next_caller[c.CallSid] == len(c.cf.Numbers)
}

func (cf * CallForward) GetCall(r * http.Request) *Call {
	call := new(Call)
	r.ParseForm()
	if err := formam.Decode(r.Form, call); err != nil {
		fmt.Println(err)
	}

	call.cf = cf

	call.Active = cf.ActiveCalls[call.CallSid]


	return call
}


func (c * CallForward) HandleCall(w http.ResponseWriter, r *http.Request) {
	call := c.GetCall(r)

	if call.NewCaller() {
		log.Debug("Incoming call from %s", call.Caller)
		c.SetActive(call)
	}

	if call.DialNext() {
		next_dial := call.GetNextNumber()
		log.Info("Dialing %s", next_dial)
		dial := NewDial(next_dial)
		dial.Send(w)
	else if call.EndOfCall() {
		log.Info("Call ended")

		hangup := NewHangup()
		hangup.Send(w)

		call.End()
	} else if call.EndOfList() {
		log.Info("Out of Numbers")

		hangup := NewNoNumbers()
		hangup.Send(w)
	}
}

func (c CallForward) HandleScreen(w http.ResponseWriter, r *http.Request) {
	call := new(Call)
	r.ParseForm()
	if err := formam.Decode(r.Form, call); err != nil {
		fmt.Println(err)
	}

	log.Debug("Attempting to connect %s with %s", call.Caller, call.Called)

	screen := NewScreen()
	screen.Send(w)
}

func (c CallForward) HandleComplete(w http.ResponseWriter, r *http.Request) {
	call := new(Call)
	r.ParseForm()
	if err := formam.Decode(r.Form, call); err != nil {
		fmt.Println(err)
	}

	log.Debug("Call connected %s to %s", call.Caller, call.Called)

	complete := NewComplete()
	complete.Send(w)
}


func main () {

    backend := logging.NewLogBackend(os.Stderr, "", 0)

    backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	cf := new(CallForward)
	cf.Numbers = make([]string, 1, 10)
	cf.ActiveCalls = make(map[string]bool)
	cf.next_caller = make(map[string]int)
	cf.Numbers[0] = "5095912174"

	http.HandleFunc("/call", cf.HandleCall)
	http.HandleFunc("/screen", cf.HandleScreen)
	http.HandleFunc("/complete", cf.HandleComplete)

	http.ListenAndServe(":8090", nil)

}
