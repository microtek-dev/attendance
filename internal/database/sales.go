package database

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CustomTime to handle multiple time formats and null values
type CustomTime struct {
	time.Time
	Valid bool // indicates if the time value is valid
}

// UnmarshalJSON parses time string into CustomTime, handling multiple formats and null values
// UnmarshalJSON parses time string into CustomTime, handling multiple formats and null values
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	str := string(b)

	// Check for null or '0000-00-00 00:00:00'
	if str == "null" || strings.Contains(str, "0001-01-01T00:00:00") {
		ct.Valid = false
		return nil
	}

	// Define possible time formats
	timeFormats := []string{
		`"2006-01-02T15:04:05.0000000"`, // With nanoseconds
		`"2006-01-02T15:04:05.000"`,     // With milliseconds
		`"2006-01-02T15:04:05"`,         // Standard format
	}

	for _, format := range timeFormats {
		ct.Time, err = time.Parse(format, str)
		if err == nil {
			ct.Valid = true
			return nil
		}
	}
	return fmt.Errorf("cannot parse %q as time: %v", str, err)
}

// MarshalJSON handles the JSON marshaling for CustomTime
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if !ct.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ct.Time)
}

type Employee struct {
	UserName                 string `json:"UserName"`
	UserErpId                string `json:"UserErpId"`
	UserRank                 int    `json:"UserRank"`
	UserDesignation          string `json:"UserDesignation"`
	ManagerErpId             string `json:"ManagerErpId"`
	RegionErpId              string `json:"RegionErpId"`
	IsFieldUser              bool   `json:"IsFieldUser"`
	HQ                       string `json:"HQ"`
	IsOrderBookingAllowed    bool   `json:"IsOrderBookingAllowed"`
	Phone                    string `json:"Phone"`
	Email                    string `json:"Email"`
	ImeiNo                   string `json:"ImeiNo"`
	DateOfJoining            string `json:"DateOfJoining"`
	DateOfLeaving            string `json:"DateOfLeaving"`
	UserType                 string `json:"UserType"`
	UserStatus               string `json:"UserStatus"`
	IsNewEntry               bool   `json:"IsNewEntry"`
	LastUpdatedAtAsEpochTime int    `json:"LastUpdatedAtAsEpochTime"`
}

func SyncEmployeeData() {
	fmt.Println("Syncing employee data...")
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.fieldassist.in/api/masterdata/employee/list?EpochTime=18", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Basic VGVzdF8xMTAwODpPRU82clBYZGRCOHdtU1pJISR4Iw==")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var employees []Employee
	err = json.Unmarshal(body, &employees)
	if err != nil {
		log.Fatal(err)
	}

	// look at the above axios request in the comments for the logic to store the employees in the database, first truncate the table and then store the active employees
	err = ProgressionDB.Exec("TRUNCATE TABLE microtek.erprecords").Error
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	errorsChan := make(chan error)

	for _, emp := range employees {
		if emp.UserErpId != "" && emp.UserStatus == "Active" {
			wg.Add(1)
			go func(emp Employee) {
				defer wg.Done()

				err = ProgressionDB.Exec(`INSERT INTO microtek.erprecords (UserName, UserErpId, UserRank, UserDesignation, ManagerErpId, RegionErpId, IsFieldUser, HQ, IsOrderBookingAllowed, Phone, Email, ImeiNo, DateOfJoining, DateOfLeaving, UserType, UserStatus, IsNewEntry, LastUpdatedAtAsEpochTime, createdAt, updatedAt) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, emp.UserName, emp.UserErpId, emp.UserRank, emp.UserDesignation, emp.ManagerErpId, emp.RegionErpId, emp.IsFieldUser, emp.HQ, emp.IsOrderBookingAllowed, emp.Phone, emp.Email, emp.ImeiNo, emp.DateOfJoining, emp.DateOfLeaving, emp.UserType, emp.UserStatus, emp.IsNewEntry, emp.LastUpdatedAtAsEpochTime, time.Now(), time.Now()).Error
				if err != nil {
					errorsChan <- err
					return
				}
			}(emp)
		}
	}

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Employee data synced successfully. Total employees: ", len(employees))
}

func SyncSalesAttendanceFromFieldAssist(date string) {
	fmt.Println("Syncing sales attendance...")
	var employees []Employee
	err := ProgressionDB.Raw("SELECT * FROM microtek.erprecords").Scan(&employees).Error
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	errorsChan := make(chan error)

	for i, emp := range employees {
		wg.Add(1)
		go func(i int, emp Employee) {
			defer wg.Done()

			// Introduce a delay of 500ms between each request to avoid rate limiting
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)

			getAttendanceForEmployee(emp.UserErpId, date)
		}(i, emp)
	}

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Sales attendance synced successfully.")
}

type FieldAssistAttendance struct {
	EmployeeName    string `json:"EmployeeName"`
	ErpId           string `json:"ErpId"`
	Date            string `json:"Date"`
	Designation     string `json:"Designation"`
	EmailId         string `json:"EmailId"`
	ContactNo       string `json:"ContactNo"`
	ManagerName     string `json:"ManagerName"`
	UserTimelineDay []struct {
		TransactionId int        `json:"TransactionId"`
		DayStartType  int        `json:"DayStartType"`
		InTime        CustomTime `json:"InTime"`
		Latitude      float32    `json:"Latitude"`
		Longitude     float32    `json:"Longitude"`
		ActivityType  string     `json:"ActivityType"`
		OutTime       CustomTime `json:"OutTime"`
	} `json:"UserTimelineDay"`
}

func getAttendanceForEmployee(employee_id string, date string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.fieldassist.in/api/timeline/list?erpId=%s&date=%s", employee_id, date), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Basic VGVzdF8xMTAwODpPRU82clBYZGRCOHdtU1pJISR4Iw==")
	req.Header.Set("Content-Type", "multipart/form-data")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var attendance FieldAssistAttendance
	err = json.Unmarshal(body, &attendance)
	if err != nil {
		log.Fatal("Error unmarshalling attendance data: ", err, string(body))
	}

	for _, task := range attendance.UserTimelineDay {
		saveSalesAttendance(employee_id, date, task)
	}
}

func saveSalesAttendance(userErpId string, punchDate string, task struct {
	TransactionId int        `json:"TransactionId"`
	DayStartType  int        `json:"DayStartType"`
	InTime        CustomTime `json:"InTime"`
	Latitude      float32    `json:"Latitude"`
	Longitude     float32    `json:"Longitude"`
	ActivityType  string     `json:"ActivityType"`
	OutTime       CustomTime `json:"OutTime"`
}) {
	var inTime, outTime interface{}
	if task.InTime.Valid {
		inTime = task.InTime.Time
	}
	if task.OutTime.Valid {
		outTime = task.OutTime.Time
	}

	err := ProgressionDB.Exec(`INSERT INTO microtek.dailytasks (UserErpId, PunchDate, TransactionId, DayStartType, InTime, Latitude, ActivityType, OutTime, Longitude, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, userErpId, punchDate, task.TransactionId, task.DayStartType, inTime, task.Latitude, task.ActivityType, outTime, task.Longitude, time.Now(), time.Now()).Error
	if err != nil {
		log.Fatal(err)
	}
}

type SalesAttendance struct {
	EmployeeId string    `gorm:"column:employee_id"`
	InTime     time.Time `gorm:"column:InTime"`
	OutTime    time.Time `gorm:"column:OutTime"`
}

func GetSalesAttendanceFromDailyTask(day string) []SalesAttendance {
	var salesAttendance []SalesAttendance
	var query string

	switch day {
	case "today":
		query = "select employee_id, InTime, OutTime from ( select daystarttype DayStartType,date_format(InTime,'%Y-%m-%d') Date,UserErpId,PunchDate, sales_mst.new_e_code employee_id, min(date_add(case when ActivityType='Day End (Normal)' then OutTime else case when InTime > PunchDate then InTime else case when OutTime > PunchDate then OutTime else NULL end end end,INTERVAL 330 minute)) InTime, max(date_add(case when ActivityType='Day Start' then InTime else case when OutTime > PunchDate then OutTime else case when InTime > PunchDate then InTime else NULL end end end,INTERVAL 330 minute)) OutTime from microtek.dailytasks as tasks, microtek.sales_employee_mapping as sales_mst where tasks.UserErpId = sales_mst.sales_id and PunchDate >= date_format(current_date(), '%Y-%m-%d') group by daystarttype,PunchDate,UserErpId order by daystarttype,usererpid,PunchDate) tt;"
	case "yesterday":
		query = "select employee_id, InTime, OutTime from ( select daystarttype DayStartType,date_format(InTime,'%Y-%m-%d') Date,UserErpId,PunchDate, sales_mst.new_e_code employee_id, min(date_add(case when ActivityType='Day End (Normal)' then OutTime else case when InTime > PunchDate then InTime else case when OutTime > PunchDate then OutTime else NULL end end end,INTERVAL 330 minute)) InTime, max(date_add(case when ActivityType='Day Start' then InTime else case when OutTime > PunchDate then OutTime else case when InTime > PunchDate then InTime else NULL end end end,INTERVAL 330 minute)) OutTime from microtek.dailytasks as tasks, microtek.sales_employee_mapping as sales_mst where tasks.UserErpId = sales_mst.sales_id and PunchDate >= date_format(current_date() - INTERVAL 1 DAY, '%Y-%m-%d') group by daystarttype,PunchDate,UserErpId order by daystarttype,usererpid,PunchDate) tt;"
	default:
		log.Fatal("Invalid day argument. Must be 'today' or 'yesterday'.")
	}

	err := ProgressionDB.Raw(query).Scan(&salesAttendance).Error
	if err != nil {
		log.Fatal("Error fetching sales attendance: ", err)
	}

	return salesAttendance
}

func GetSalesAttendanceFromDailyTaskUnmatched(day string) []SalesAttendance {
	var salesAttendance []SalesAttendance
	var query string

	switch day {
	case "today":
		query = "select UserErpId as employee_id, min(date_add(case when ActivityType='Day End (Normal)' then OutTime else  case when InTime > PunchDate then InTime else case when OutTime > PunchDate then OutTime else NULL end end end,INTERVAL 330 minute)) InTime, max(date_add(case when ActivityType='Day Start' then InTime else case when OutTime > PunchDate then OutTime else case when InTime > PunchDate then InTime else NULL end end end,INTERVAL 330 minute)) OutTime from microtek.dailytasks as tasks where PunchDate >= date_format(current_date(), '%Y-%m-%d') and UserErpId Not In (select distinct(sales_id) from microtek.sales_employee_mapping) group by daystarttype,PunchDate,UserErpId order by daystarttype,usererpid,PunchDate;"
	case "yesterday":
		query = "select UserErpId as employee_id, min(date_add(case when ActivityType='Day End (Normal)' then OutTime else  case when InTime > PunchDate then InTime else case when OutTime > PunchDate then OutTime else NULL end end end,INTERVAL 330 minute)) InTime, max(date_add(case when ActivityType='Day Start' then InTime else case when OutTime > PunchDate then OutTime else case when InTime > PunchDate then InTime else NULL end end end,INTERVAL 330 minute)) OutTime from microtek.dailytasks as tasks where PunchDate >= date_format(current_date() - INTERVAL 1 DAY, '%Y-%m-%d') and UserErpId Not In (select distinct(sales_id) from microtek.sales_employee_mapping) group by daystarttype,PunchDate,UserErpId order by daystarttype,usererpid,PunchDate;"
	default:
		log.Fatal("Invalid day argument. Must be 'today' or 'yesterday'.")
	}

	err := ProgressionDB.Raw(query).Scan(&salesAttendance).Error
	if err != nil {
		log.Fatal("Error fetching sales attendance: ", err)
	}

	return salesAttendance
}

func SaveSalesAttendanceLocallyBulk(salesAttendance []SalesAttendance) {
	var wg sync.WaitGroup
	errorsChan := make(chan error)

	for _, attendance := range salesAttendance {
		wg.Add(1)
		go func(attendance SalesAttendance) {
			defer wg.Done()

			var inTime, outTime time.Time
			if attendance.InTime.IsZero() {
				if !attendance.OutTime.IsZero() {
					inTime = attendance.OutTime
				} else {
					inTime = time.Now()
				}
			} else {
				inTime = attendance.InTime
			}

			if attendance.OutTime.IsZero() {
				if !attendance.InTime.IsZero() {
					outTime = attendance.InTime
				} else {
					outTime = time.Now()
				}
			} else {
				outTime = attendance.OutTime
			}

			err := ProgressionDB.Exec(`INSERT INTO microtek.sales_dailyattendances (employee_id, InTime, OutTime, createdAt, updatedAt) VALUES (?,?,?,?,?)`, attendance.EmployeeId, inTime, outTime, time.Now(), time.Now()).Error
			if err != nil {
				errorsChan <- err
				return
			}
		}(attendance)
	}

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		if err != nil {
			log.Fatal(err)
		}
	}
}

func InsertSalesToAwsFrtDataBulk(punchData []SalesAttendance) {
	var wg sync.WaitGroup
	chunkSize := 100

	for i := 0; i < len(punchData); i += chunkSize {
		end := i + chunkSize

		if end > len(punchData) {
			end = len(punchData)
		}

		wg.Add(1)
		go func(punchData []SalesAttendance) {
			defer wg.Done()

			for _, data := range punchData {
				var inTime, outTime time.Time
				if data.InTime.IsZero() {
					if !data.OutTime.IsZero() {
						inTime = data.OutTime
					} else {
						break
					}
				} else {
					inTime = data.InTime
				}

				if data.OutTime.IsZero() {
					if !data.InTime.IsZero() {
						outTime = data.InTime
					} else {
						break
					}
				} else {
					outTime = data.OutTime
				}

				InsertIntoAwsFrtData(data.EmployeeId, inTime)
				InsertIntoAwsFrtData(data.EmployeeId, outTime)
			}
		}(punchData[i:end])
	}

	wg.Wait()

	log.Println("Inserted data into AWS successfully.")
}
