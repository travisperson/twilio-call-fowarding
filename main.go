package main

import (
	"bitbucket.org/ckvist/twilio/twiml"
	"fmt"
	"time"
	"os"
//	"io"
	"net/http"
	"github.com/op/go-logging"
	//"github.com/ajg/form"
	"github.com/monoculum/formam"
)

var log = logging.MustGetLogger("cf")
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc:20.20s} %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

type Number struct {
	CallNumber string
	StartTime time.Time
	EndTime time.Time
}

type CallForwardService struct {
	Numbers []Number
	ActiveCalls map[string] *Call
}

func (cfs * CallForwardService) GetShiftNumbers () []Number {
	nums := make([]Number, 0, 100)
	
	log.Debug("Shift Numbers:")

	t := time.Now()
	for _, n := range cfs.Numbers {
		if t.After(n.StartTime) && t.Before(n.EndTime) {
			log.Debug(" %s", n.CallNumber)
			nums = append(nums, n)
		}
	}

	return nums
}

func (cfs * CallForwardService) GetCall (r * http.Request) *Call {

	sid := r.Header.Get("X-Twilio-CallSid")

	call, ok := cfs.ActiveCalls[sid]

	if !ok {
		call = cfs.NewCall(r)
	} else {
		var tcall TwimlCall

		r.ParseForm()
		if err := formam.Decode(r.Form, &tcall); err != nil {
			log.Error("%s", err)
		}
		call.CurrentInfo = tcall
	}

	return call
}

func (cfs * CallForwardService) NewCall (r * http.Request) *Call {

	var tcall TwimlCall

	r.ParseForm()
	if err := formam.Decode(r.Form, &tcall); err != nil {
		log.Error("%s", err)
	}

	c := new(Call)

	c.Numbers = cfs.GetShiftNumbers()
	c.NumberIndex = 0
	c.CallSid = tcall.CallSid
	c.Caller = tcall.Caller
	c.OriginalCall = tcall
	c.CurrentInfo = tcall

	cfs.ActiveCalls[c.CallSid] = c

	return c
}

func (cfs * CallForwardService) HandleCall(w http.ResponseWriter, r *http.Request) {
	call := cfs.GetCall(r)

	if call.NewCall() {
		log.Info("[%s] Incoming call from %s", call.CallSid, call.Caller)
	}
	
	// The connected call has ended
	if call.CallForwardEnded() == true {
		log.Info("Call Forward Ended")
		return;
	}

	// Caller has hungup
	if call.EndOfCall() {
		log.Info("[%s] Call ended", call.CallSid)

		hangup := NewHangup()
		hangup.Send(w)

		//call.End()
	} else {
		if !call.EndOfNumbers() {
			next_dial := call.GetNextNumber()
			log.Info("[%s] Dialing %s", call.CallSid, next_dial)

			dial := NewDial(next_dial)
			dial.Send(w)
		} else {
			log.Info("[%s] Out of Numbers", call.CallSid)

			hangup := NewNoNumbers()
			hangup.Send(w)
		}
	}
}

func (cfs CallForwardService) HandleScreen(w http.ResponseWriter, r *http.Request) {
	call := cfs.GetCall(r)

	log.Info("[%s] Attempting to connect %s with %s", call.CallSid, call.CurrentInfo.Caller, call.CurrentInfo.Called)

	screen := NewScreen()
	screen.Send(w)
}

func (cfs CallForwardService) HandleComplete(w http.ResponseWriter, r *http.Request) {
	call := cfs.GetCall(r)

	log.Info("[%s] Call connected %s to %s", call.CallSid, call.CurrentInfo.Caller, call.CurrentInfo.Called)

	complete := NewComplete()
	complete.Send(w)
}

func main () {

    backend := logging.NewLogBackend(os.Stderr, "", 0)

    backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	cfs := new(CallForwardService)
	cfs.Numbers = make([]Number, 0, 10)
	cfs.ActiveCalls = make(map[string]*Call)

	t := time.Now()

	nb := Number{"5095912174", t, t.Add(time.Hour)}
	cfs.Numbers = append(cfs.Numbers, nb)

	/*
	nb = Number{"5090000000", t.Add((-1) * time.Hour), t}
	cfs.Numbers = append(cfs.Numbers, nb)
	*/

	http.HandleFunc("/call", cfs.HandleCall)
	http.HandleFunc("/screen", cfs.HandleScreen)
	http.HandleFunc("/complete", cfs.HandleComplete)

	http.ListenAndServe(":8090", nil)

}
