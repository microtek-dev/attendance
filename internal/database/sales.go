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
	err = TestDB.Exec("TRUNCATE TABLE sales_employee_records").Error
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

				var convertedDateOfJoining time.Time
				var err error
				if emp.DateOfJoining != "" {
					convertedDateOfJoining, err = time.Parse("2006-01-02T15:04:05", emp.DateOfJoining)
					if err != nil {
						errorsChan <- err
						return
					}
				}

				var convertedDateOfLeaving time.Time
				if emp.DateOfLeaving != "" {
					convertedDateOfLeaving, err = time.Parse("2006-01-02T15:04:05", emp.DateOfLeaving)
					if err != nil {
						errorsChan <- err
						return
					}
				}

				err = TestDB.Create(&SalesEmployeeRecords{UserName: emp.UserName, UserErpId: emp.UserErpId, UserRank: emp.UserRank, UserDesignation: emp.UserDesignation, ManagerErpId: emp.ManagerErpId, RegionErpId: emp.RegionErpId, IsFieldUser: emp.IsFieldUser, HQ: emp.HQ, IsOrderBookingAllowed: emp.IsOrderBookingAllowed, Phone: emp.Phone, Email: emp.Email, ImeiNo: emp.ImeiNo, DateOfJoining: convertedDateOfJoining, DateOfLeaving: convertedDateOfLeaving, UserType: emp.UserType, UserStatus: emp.UserStatus, IsNewEntry: emp.IsNewEntry, LastUpdatedAtAsEpochTime: emp.LastUpdatedAtAsEpochTime}).Error
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

/*
function dailyTask() {
  console.log("hello");
  erpRecord.findAll().then((data) => {
    data.map(async (emp, index) => {
      //getTask(emp.dataValues.UserErpId)
      await sleep(index * 200);
      //console.log(emp.dataValues.UserErpId);
      getTask(emp.dataValues.UserErpId);
      //setTimeout(getTask, 200 * index, emp.dataValues.UserErpId)
    });
  });
}
//employeeData()
//dailyTask()
async function storeTask(task) {
  dailyReport
    .create(task)
    .then((data1) => {})
    .catch((err) => {
      console.log(err);
    });
}
*/

func SyncSalesAttendance() {}

/*
async function getTask(id) {
  const date = getMomentDate(); // mm/dd/yyyy
  const datearray = date.split("/");
  const sqlDate = datearray[2] + "-" + datearray[0] + "-" + datearray[1];
  console.log(sqlDate);
  axios({
    method: "get",
    url: `https://api.fieldassist.in/api/timeline/list?erpId=${id}&date=${date}`,
    headers: {
      "Content-Type": "multipart/form-data",
      Authorization: "Basic VGVzdF8xMTAwODpPRU82clBYZGRCOHdtU1pJISR4Iw==",
    },
  })
    .then((taskResponse) => {
      taskResponse?.data?.UserTimelineDay?.map((task, index) => {
        task.UserErpId = taskResponse.data.ErpId;
        task.PunchDate = sqlDate;
        setTimeout(storeTask, 100 * index, task);
      });
    })
    .catch((err) => {
      console.log(err);
    });
}
*/

/*
	{
		"EmployeeName": "Md Neyaz Ahmad Khan",
		"ErpId": "57986",
		"Date": "2024-06-10T00:00:00",
		"Designation": "Area Sales Executive",
		"EmailId": "mdneyazkhan76@gmail.com",
		"ContactNo": "7488170472",
		"ManagerName": "Rahul Kumar",
		"UserTimelineDay": []
	}
*/
type FieldAssistAttendance struct {
	EmployeeName    string `json:"EmployeeName"`
	ErpId           string `json:"ErpId"`
	Date            string `json:"Date"`
	Designation     string `json:"Designation"`
	EmailId         string `json:"EmailId"`
	ContactNo       string `json:"ContactNo"`
	ManagerName     string `json:"ManagerName"`
	UserTimelineDay []struct{}
}

func getAttendanceForEmployee(employee_id string) {
	date := time.Now().Format("2006-01-02")

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.fieldassist.in/api/timeline/list?erpId=%s&date=%s", employee_id, date), nil)
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

	var attendances []FieldAssistAttendance
	err = json.Unmarshal(body, &attendances)
	if err != nil {
		log.Fatal(err)
	}
}

// func saveAttendance(attendance FieldAssistAttendance) {
// 	err := TestDB.Create(&Dailytask{UserErpId: attendance.UserErpId, PunchDate: attendance.PunchDate}).Error
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
