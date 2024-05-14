package database

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Employee struct {
	UserName              string `json:"UserName"`
	UserErpId             string `json:"UserErpId"`
	UserRank              int    `json:"UserRank"`
	UserDesignation       string `json:"UserDesignation"`
	ManagerErpId          string `json:"ManagerErpId"`
	RegionErpId           string `json:"RegionErpId"`
	IsFieldUser           bool   `json:"IsFieldUser"`
	HQ                    string `json:"HQ"`
	IsOrderBookingAllowed bool   `json:"IsOrderBookingAllowed"`
	Phone                 string `json:"Phone"`
	Email                 string `json:"Email"`
	ImeiNo                string `json:"ImeiNo"`
	DateOfJoining         string `json:"DateOfJoining"`
	DateOfLeaving         string `json:"DateOfLeaving"`
	UserType              string `json:"UserType"`
	UserStatus            string `json:"UserStatus"`
	IsNewEntry            bool   `json:"IsNewEntry"`
	LastUpdatedAtAsEpochTime int `json:"LastUpdatedAtAsEpochTime"`
}

func SyncEmployeeData() {
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
		err = DB.Exec("TRUNCATE TABLE erprecords").Error
		if err != nil {
			log.Fatal(err)
		}

		for _, emp := range employees {
			if emp.UserErpId != "" && emp.UserStatus == "Active" {
				err = DB.Create(&emp).Error
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		log.Println("Employee data synced successfully. Total employees: ", len(employees))
}
