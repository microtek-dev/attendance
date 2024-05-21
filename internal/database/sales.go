package database

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

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
	err = TestDB.Exec("TRUNCATE TABLE erprecords").Error
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

				err = TestDB.Exec(`INSERT INTO erprecords (UserName, UserErpId, UserRank, UserDesignation, ManagerErpId, RegionErpId, IsFieldUser, HQ, IsOrderBookingAllowed, Phone, Email, ImeiNo, DateOfJoining, DateOfLeaving, UserType, UserStatus, IsNewEntry, LastUpdatedAtAsEpochTime, createdAt, updatedAt) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, emp.UserName, emp.UserErpId, emp.UserRank, emp.UserDesignation, emp.ManagerErpId, emp.RegionErpId, emp.IsFieldUser, emp.HQ, emp.IsOrderBookingAllowed, emp.Phone, emp.Email, emp.ImeiNo, emp.DateOfJoining, emp.DateOfLeaving, emp.UserType, emp.UserStatus, emp.IsNewEntry, emp.LastUpdatedAtAsEpochTime, time.Now(), time.Now()).Error
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

func SyncSalesAttendance() {
	fmt.Println("Syncing sales attendance...")
	var employees []Employee
	err := TestDB.Raw("SELECT * FROM erprecords").Scan(&employees).Error
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	errorsChan := make(chan error)

	for i, emp := range employees {
		wg.Add(1)
		go func(i int, emp Employee) {
			defer wg.Done()

			// Introduce a delay between each request
			time.Sleep(time.Duration(i) * 50 * time.Millisecond)

			getAttendanceForEmployee(emp.UserErpId)
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
		UserErpId     string    `json:"UserErpId"`
		PunchDate     string    `json:"PunchDate"`
		TransactionId string    `json:"TransactionId"`
		DayStartType  int       `json:"DayStartType"`
		InTime        time.Time `json:"InTime"`
		Latitude      string    `json:"Latitude"`
		ActivityType  string    `json:"ActivityType"`
		OutTime       time.Time `json:"OutTime"`
		Longitude     string    `json:"Longitude"`
	} `json:"UserTimelineDay"`
}

func getAttendanceForEmployee(employee_id string) {
	date := time.Now().Format("2006-01-02")

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
	fmt.Println(string(body))
	err = json.Unmarshal(body, &attendance)
	if err != nil {
		log.Fatal(err)
	}

	for _, task := range attendance.UserTimelineDay {
		saveSalesAttendance(task)
	}
}

func saveSalesAttendance(task struct {
	UserErpId     string    `json:"UserErpId"`
	PunchDate     string    `json:"PunchDate"`
	TransactionId string    `json:"TransactionId"`
	DayStartType  int       `json:"DayStartType"`
	InTime        time.Time `json:"InTime"`
	Latitude      string    `json:"Latitude"`
	ActivityType  string    `json:"ActivityType"`
	OutTime       time.Time `json:"OutTime"`
	Longitude     string    `json:"Longitude"`
}) {
	err := TestDB.Raw(`INSERT INTO dailytasks (UserErpId, PunchDate, TransactionId, DayStartType, InTime, Latitude, ActivityType, OutTime, Longitude) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, task.UserErpId, task.PunchDate, task.TransactionId, task.DayStartType, task.InTime, task.Latitude, task.ActivityType, task.OutTime, task.Longitude).Error
	if err != nil {
		log.Fatal(err)
	}
}

type SalesAttendance struct {
	EmployeeId string    `gorm:"column:employee_id"`
	InTime     time.Time `gorm:"column:InTime"`
	OutTime    time.Time `gorm:"column:OutTime"`
}

func GetSalesAttendanceDailyTask() []SalesAttendance {
	var salesAttendance []SalesAttendance
	err := TestDB.Raw("select employee_id, InTime, OutTime from ( select daystarttype DayStartType,date_format(InTime,'%Y-%m-%d') Date,UserErpId,PunchDate, sales_mst.new_e_code employee_id, min(date_add(case when ActivityType='Day End (Normal)' then OutTime else case when InTime > PunchDate then InTime else case when OutTime > PunchDate then OutTime else NULL end end end,INTERVAL 330 minute)) InTime, max(date_add(case when ActivityType='Day Start' then InTime else case when OutTime > PunchDate then OutTime else case when InTime > PunchDate then InTime else NULL end end end,INTERVAL 330 minute)) OutTime from dailytasks as tasks, sales_employee_mapping as sales_mst where tasks.UserErpId = sales_mst.sales_id and PunchDate >= date_format(current_date() - INTERVAL 1 DAY, '%Y-%m-%d') group by daystarttype,PunchDate,UserErpId order by daystarttype,usererpid,PunchDate) tt;").Scan(&salesAttendance).Error
	if err != nil {
		log.Fatal("Error fetching sales attendance: ", err)
	}

	return salesAttendance
}

func SaveSalesAttendanceBulk(salesAttendance []SalesAttendance) {
	var wg sync.WaitGroup
	errorsChan := make(chan error)

	for _, attendance := range salesAttendance {
		wg.Add(1)
		go func(attendance SalesAttendance) {
			defer wg.Done()

			err := TestDB.Exec(`INSERT INTO sales_dailyattendances (employee_id, InTime, OutTime, createdAt, updatedAt) VALUES (?,?,?,?)`, attendance.EmployeeId, attendance.InTime, attendance.OutTime, time.Now(), time.Now()).Error
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
				InsertIntoAwsFrtData(data.EmployeeId, data.InTime)
				InsertIntoAwsFrtData(data.EmployeeId, data.OutTime)
			}
		}(punchData[i:end])
	}

	wg.Wait()

	log.Println("Inserted data into AWS successfully.")
}
