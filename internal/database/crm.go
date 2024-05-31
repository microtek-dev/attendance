package database

import (
	"log"
	"sync"
	"time"
)

type CrmAttendanceLog struct {
	EmployeeId string    `gorm:"column:employee_id"`
	InTime     time.Time `gorm:"column:InTime"`
	OutTime    time.Time `gorm:"column:OutTime"`
}

func InsertIntoCrmAttendanceLog(employeeId string, inTime time.Time, outTime time.Time) {
	err := TestDB.Exec("Insert into crm_dailyattendances (employee_id, InTime, OutTime, createdAt, updatedAt) values (?, ?, ?, ?, ?)", employeeId, inTime, outTime, time.Now(), time.Now()).Error
	if err != nil {
		log.Fatalf("failed to insert into crm_attendance_log: %v", err)
	}
}

func GetPreviousDayCRMAttendanceData() []CrmAttendanceLog {
	var result []CrmAttendanceLog
	err := TestDB.Raw("select new_e_code employee_id,in_time InTime, out_time OutTime from ( select * from attendancedata t1,crm_employee_mapping t2 where  t1.punch_date >=date_format(current_date() -  INTERVAL 1 DAY,'%Y-%m-%d') and t1.punch_date < date_format(current_date(),'%Y-%m-%d') and t1.eng_id=t2.employee_id ) tt;").Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch crm_dailyattendances: %v", err)
	}
	return result
}

func GetCurrentDayCRMAttendanceData() []CrmAttendanceLog {
	var result []CrmAttendanceLog
	err := TestDB.Raw("select new_e_code employee_id,in_time InTime, out_time OutTime from ( select * from attendancedata t1,crm_employee_mapping t2 where t1.punch_date >= date_format(current_date(), '%Y-%m-%d') and t1.eng_id = t2.employee_id ) tt;").Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch crm_dailyattendances: %v", err)
	}
	return result
}

func GetCurrentDayUnmatchedCRMAttendanceData() []CrmAttendanceLog {
	var result []CrmAttendanceLog
	err := TestDB.Raw("select eng_id employee_id,in_time InTime, out_time OutTime from ( select * from attendancedata t1 where  t1.punch_date >=date_format(current_date(),'%Y-%m-%d') and t1.eng_id Not In (select distinct(employee_id) from crm_employee_mapping) ) tt;").Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch crm_dailyattendances: %v", err)
	}
	return result
}

func GetPreviousDayUnmatchedCRMAttendanceData() []CrmAttendanceLog {
	var result []CrmAttendanceLog
	err := TestDB.Raw("select eng_id employee_id,in_time InTime, out_time OutTime from ( select * from attendancedata t1 where  t1.punch_date >=date_format(current_date() -  INTERVAL 1 DAY,'%Y-%m-%d') and  t1.punch_date < date_format(current_date(),'%Y-%m-%d') and  t1.eng_id Not In (select distinct(employee_id) from crm_employee_mapping) ) tt;").Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch crm_dailyattendances: %v", err)
	}
	return result
}

func InsertCrmToAwsFrtDataBulk(punchData []CrmAttendanceLog) {
	var wg sync.WaitGroup
	chunkSize := 100

	for i := 0; i < len(punchData); i += chunkSize {
		end := i + chunkSize

		if end > len(punchData) {
			end = len(punchData)
		}

		wg.Add(1)
		go func(punchData []CrmAttendanceLog) {
			defer wg.Done()

			for _, data := range punchData {
				InsertIntoAwsFrtData(data.EmployeeId, data.InTime)
				InsertIntoAwsFrtData(data.EmployeeId, data.OutTime)
			}
		}(punchData[i:end])
	}

	wg.Wait()

	log.Println("Inserted data into AWS successfully. Total records: ", len(punchData))
}
