package main

type Call struct {
	Numbers []Number
	NumberIndex int
	CallSid string
	Caller string
	OriginalCall TwimlCall
	CurrentInfo TwimlCall
}

// We assume if we haven't asked for a number we are a new call
func (c * Call) NewCall() bool {
	return c.NumberIndex == 0
}

func (c * Call) EndOfNumbers() bool {
	return !(c.NumberIndex < len(c.Numbers))
}

func (c * Call) GetNextNumber() string {
	num := c.Numbers[c.NumberIndex]
	c.NumberIndex++
	return num.CallNumber
}

func (c * Call) CallForwardEnded() bool {
	// in-progress means the staff ended, compelted means the caller ended
	return c.CurrentInfo.DialCallDuration != 0 && (c.CurrentInfo.CallStatus == "in-progress" ||
	c.CurrentInfo.CallStatus == "completed")
}

func (c * Call) EndOfCall() bool {
	return c.CurrentInfo.CallDuration != 0 && c.CurrentInfo.CallStatus == "completed"
}
