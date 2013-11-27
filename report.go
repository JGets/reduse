package main

import(
	"strings"
)

type ReportType int


const(
	UNKNOWN ReportType = iota
	SPAM ReportType = iota
	ILLEGAL ReportType = iota
	MORALLY_OBJECTIONABLE ReportType = iota
	TAKEDOWN_REQUEST ReportType = iota
)


func ReportTypeForString(str string) ReportType{
	switch strings.ToUpper(str) {
		case "SPAM":
			return SPAM
		case "ILLEGAL":
			return ILLEGAL
		case "MORALLY_OBJECTIONABLE":
		case "MORALLY OBJECTIONABLE":
			return MORALLY_OBJECTIONABLE
		case "TAKEDOWN_REQUEST":
		case "TAKEDOWN REQUEST":
			return TAKEDOWN_REQUEST
		default:
			return UNKNOWN
	}
}

func (rt ReportType) String() string{
	switch rt{
		case SPAM:
			return "Spam"
		case ILLEGAL:
			return "Illegal"
		case MORALLY_OBJECTIONABLE:
			return "Morally Objectionable"
		case TAKEDOWN_REQUEST:
			return "Takedown Request"
		case UNKNOWN:
		default:
			return "Unknown"
	}
}
