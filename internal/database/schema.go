package database

import (
	"time"
)

type ErpRecord struct {
    UserName              string
    UserErpId             string
    UserRank              int
    UserDesignation       string
    ManagerErpId          string
    RegionErpId           string
    IsFieldUser           bool
    HQ                    string
    IsOrderBookingAllowed bool
    Phone                 string
    Email                 string
    ImeiNo                string
    DateOfJoining         time.Time
    DateOfLeaving         time.Time
    UserType              string
    UserStatus            string
    IsNewEntry            bool
    LastUpdatedAtAsEpochTime int
}

type ApprovalInbox struct {
		ApprovalId          int
		RequestType         string
		RequestId           int
		ApprovalLevel       string
		AppliedBy           string
		ApproverEmployeeId  string
		Status              string
		ApproverRemarks     string
		ApprovalDate        time.Time
		CreatedAt           time.Time
		CreatedBy           string
		UpdatedAt           time.Time
		UpdatedBy           string
}

type Attendancedata struct {
		PunchDate  string
		EngId      string
		Locusername string
		Dname       string
		Jobscount   float64
		InTime     string
		OutTime    string
}

type CrmDailyattendance struct {
		Id          int
		EmployeeId  string
		InTime      time.Time
		OutTime     time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
}

type CrmEmployeeMapping struct {
		Id                     int
		NewECode               int
		EmployeeId             int
		CrmId                  string
		NameOfTheEmployee      string
		EmploymentStatus       string
		Department             string
		FunctionalDesignation  string
		OfficeLocation         string
		State                  string
}

type Dailyattendance struct {
		Id           int
		UserErpId    string
		DayStartType int
		DayStart     time.Time
		DayEnd       time.Time
		Date         string
		CreatedAt    time.Time
		UpdatedAt    time.Time
}

type Dailytask struct {
		Id            int
		UserErpId     string
		PunchDate     string
		TransactionId string
		DayStartType  int
		InTime        time.Time
		Latitude      string
		ActivityType  string
		OutTime       time.Time
		Longitude     string
		CreatedAt     time.Time
		UpdatedAt     time.Time
}

type DatabaseMaster struct {
		DatabaseId          int
		DatabaseName        string
		DatabaseDescription string
}

type EmployeeRoster struct {}

type SalesDailyattendance struct {
		Id          int
		EmployeeId  string
		InTime      time.Time
		OutTime     time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
}

type SalesEmployeeMapping struct {
		Id                     int
		NewECode               int
		SalesId                int
		NameOfTheEmployee      string
		Department             string
		FunctionalDesignation  string
		OfficeLocation         string
		State                  string
}
