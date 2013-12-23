package main

import(
	"strings"
	"time"
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
	logger.Printf("ReportTypeForString: '%v'", str)
	switch strings.ToUpper(str) {
		case "SPAM":
			return SPAM
		case "ILLEGAL":
			return ILLEGAL
		case "MORALLY_OBJECTIONABLE":
			return MORALLY_OBJECTIONABLE
		case "MORALLY OBJECTIONABLE":
			return MORALLY_OBJECTIONABLE
		case "TAKEDOWN_REQUEST":
			return TAKEDOWN_REQUEST
		case "TAKEDOWN REQUEST":
			return TAKEDOWN_REQUEST
		default:
			return UNKNOWN
	}
	
	return UNKNOWN
}

func (rt ReportType) String() string{
	switch rt{
		case SPAM:
			return "SPAM"
		case ILLEGAL:
			return "ILLEGAL"
		case MORALLY_OBJECTIONABLE:
			return "MORALLY_OBJECTIONABLE"
		case TAKEDOWN_REQUEST:
			return "TAKEDOWN_REQUEST"
		case UNKNOWN:
		default:
			return "UNKNOWN"
	}
	
	return "UNKNOWN"
}



type Report struct{
	Hash string
	OriginIP string
	Type ReportType
	Comment string
	Date time.Time
}


func NewEmptyReport() *Report{
	t := time.Now().UTC()
	return &Report{"", "", UNKNOWN, "", t}
}

func NewReport(hash string, ip string, rtype ReportType, comment string) *Report{
	t := time.Now().UTC()
	return &Report{hash, ip, rtype, comment, t}
}






