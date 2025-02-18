package database

import (
	"log"
	"sync"
	"time"
)

type CrmAttendanceLog struct {
	EmployeeId string `gorm:"column:employee_id"`
	InTime     string `gorm:"column:InTime"`
	OutTime    string `gorm:"column:OutTime"`
}

func InsertIntoCrmAttendanceLog(employeeId string, inTime time.Time, outTime time.Time) {
	err := TestDB.Exec("Insert into crm_dailyattendances (employee_id, InTime, OutTime, createdAt, updatedAt) values (?, ?, ?, ?, ?)", employeeId, inTime, outTime, time.Now(), time.Now()).Error
	if err != nil {
		log.Fatalf("failed to insert into crm_attendance_log: %v", err)
	}
}

func InsertIntoCRMAttendanceLogBulk(punchData []CrmAttendanceLog) {
	var wg sync.WaitGroup
	chunkSize := 100
	skippedRecords := 0

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	for i := 0; i < len(punchData); i += chunkSize {
		end := i + chunkSize

		if end > len(punchData) {
			end = len(punchData)
		}

		wg.Add(1)
		go func(punchData []CrmAttendanceLog) {
			defer wg.Done()

			for _, data := range punchData {
				// Parse time in IST timezone
				convertedInTime, err := time.ParseInLocation("2006-01-02 15:04:05", data.InTime, loc)
				if err != nil {
					log.Printf("Skipping record for employee %s: invalid InTime format: %v", data.EmployeeId, err)
					skippedRecords++
					continue
				}
				convertedOutTime, err := time.ParseInLocation("2006-01-02 15:04:05", data.OutTime, loc)
				if err != nil {
					log.Printf("Skipping record for employee %s: invalid OutTime format: %v", data.EmployeeId, err)
					skippedRecords++
					continue
				}

				// Validate times are not zero
				if convertedInTime.IsZero() || convertedOutTime.IsZero() {
					log.Printf("Skipping record for employee %s: zero timestamp detected", data.EmployeeId)
					skippedRecords++
					continue
				}

				// Adjust timezone offset
				_, offset := convertedInTime.In(loc).Zone()
				convertedInTime = convertedInTime.Add(-time.Duration(offset) * time.Second)
				_, offset = convertedOutTime.In(loc).Zone()
				convertedOutTime = convertedOutTime.Add(-time.Duration(offset) * time.Second)

				InsertIntoCrmAttendanceLog(data.EmployeeId, convertedInTime, convertedOutTime)
			}
		}(punchData[i:end])
	}

	wg.Wait()
	log.Printf("Inserted data into CRM successfully. Total records: %d, Skipped records: %d", len(punchData), skippedRecords)
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
	skippedRecords := 0

	// Load the "Asia/Kolkata" timezone
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	for i := 0; i < len(punchData); i += chunkSize {
		end := i + chunkSize

		if end > len(punchData) {
			end = len(punchData)
		}

		wg.Add(1)
		go func(punchData []CrmAttendanceLog) {
			defer wg.Done()

			for _, data := range punchData {
				// Parse time in IST timezone
				convertedInTime, err := time.ParseInLocation("2006-01-02 15:04:05", data.InTime, loc)
				if err != nil {
					log.Printf("Skipping record for employee %s: invalid InTime format: %v", data.EmployeeId, err)
					skippedRecords++
					continue
				}
				convertedOutTime, err := time.ParseInLocation("2006-01-02 15:04:05", data.OutTime, loc)
				if err != nil {
					log.Printf("Skipping record for employee %s: invalid OutTime format: %v", data.EmployeeId, err)
					skippedRecords++
					continue
				}

				// Validate times are not zero
				if convertedInTime.IsZero() || convertedOutTime.IsZero() {
					log.Printf("Skipping record for employee %s: zero timestamp detected", data.EmployeeId)
					skippedRecords++
					continue
				}

				// Adjust timezone offset
				_, offset := convertedInTime.In(loc).Zone()
				convertedInTime = convertedInTime.Add(-time.Duration(offset) * time.Second)
				_, offset = convertedOutTime.In(loc).Zone()
				convertedOutTime = convertedOutTime.Add(-time.Duration(offset) * time.Second)

				InsertIntoAwsFrtData(data.EmployeeId, convertedInTime)
				InsertIntoAwsFrtData(data.EmployeeId, convertedOutTime)
			}
		}(punchData[i:end])
	}

	wg.Wait()

	log.Printf("Inserted data into AWS successfully. Total records: %d, Skipped records: %d", len(punchData), skippedRecords)
}
