package database

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

/*
async function fetch_daily_punch() {
  // fetch todays all punch
  // store in our data base
  const date = getdate();

  const day = parseInt(date.day);
  const month = parseInt(date.month);
  const frt_max_logs = await sequelize.query(
    `select max(cast(frt_log_id as signed)) max_id from frt_logs where log_date > '${date.year}-${date.month}-01'`,
    { type: QueryTypes.SELECT },
  );
  console.log(frt_max_logs);
  var max_fetch_id = '0';
  if (frt_max_logs[0]?.max_id) {
    max_fetch_id = frt_max_logs[0]?.max_id;
  }
  if (day == 1) {
    const frt_max_logs_day1 = await sequelize.query(
      `select max(cast(frt_log_id as signed)) max_id from frt_logs where log_date > '${date.year}-${date.month}-${date.day}'`,
      { type: QueryTypes.SELECT },
    );
    console.log(frt_max_logs_day1);
    if (frt_max_logs_day1[0]?.max_id) {
      max_fetch_id = frt_max_logs_day1[0]?.max_id;
    } else {
      max_fetch_id = '0';
    }
  }
  const frt_data = await frtdb.sequelize.query(
    `select top 20000 DeviceLogId frt_log_id, DeviceId device_id, UserId user_id, LogDate log_date, C1 log_type, CreatedDate frt_created_date from DeviceLogs_${month}_${date.year} where DeviceLogId > ${max_fetch_id}  order by DeviceLogId`,
    { type: QueryTypes.SELECT },
  );
  console.log(frt_data);
  if (!frt_data.length) {
  }
  const maxRetries = 3;
  let currentRetry = 0;
  let backoffDelay = 100; // milliseconds
  frt_data.map(async (item) => {
    while (currentRetry < maxRetries) {
      try {
        const frt_logs_insert = await sequelize.query(
          `REPLACE INTO frt_logs (device_id, user_id, log_date, log_type, frt_created_date, frt_log_id) VALUES ('${item.device_id}','${
            item.user_id
          }','${item.log_date.toISOString().slice(0, 19).replace('T', ' ')}','${item.log_type}','${item.frt_created_date
            .toISOString()
            .slice(0, 19)
            .replace('T', ' ')}','${item.frt_log_id}');`,
          { type: QueryTypes.INSERT },
        );
        break; // If the query is successful, break the loop
      } catch (error) {
        if ((error.name === 'SequelizeDatabaseError' && error.parent && error.parent.errno === 1213) || error.parent.errno === 1205) {
          console.log('sandeep', error);
          console.log(`Error occurred (either deadlock or timeout). Retry attempt ${currentRetry + 1} after ${backoffDelay}ms.`);
          currentRetry++;
          await new Promise((resolve) => setTimeout(resolve, backoffDelay));
          backoffDelay *= 2; // Double the delay for the next retry if needed
        } else {
          throw error; // If the error is not a deadlock or timeout, throw it
        }
      }
    }
  });
}
*/

type Date struct {
	Day   string
	Month string
	Year  string
}

func FetchFRTMaxFetchId() int {
	date := getTodayDate()

	type MaxIdFRTResult struct {
		MaxID int `gorm:"column:max_id"`
	}

	var result MaxIdFRTResult

	queryDate := fmt.Sprintf("%s-%s-01", date.Year, date.Month)
	if date.Day == "01" {
		queryDate = fmt.Sprintf("%s-%s-%s", date.Year, date.Month, date.Day)
	}

	err := ProgressionDB.Raw("select max(cast(frt_log_id as signed)) max_id from frt_logs where log_date > ?", queryDate).Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch max id: %v", err)
	}

	return result.MaxID
}

func getTodayDate() Date {
	today := time.Now()

	return Date{
		Day:   fmt.Sprintf("%02d", today.Day()),
		Month: fmt.Sprintf("%02d", int(today.Month())),
		Year:  fmt.Sprintf("%04d", today.Year()),
	}
}

type FRTData struct {
	FRTLogID       int       `gorm:"column:frt_log_id"`
	DeviceID       int       `gorm:"column:device_id"`
	UserID         string    `gorm:"column:user_id"`
	LogDate        time.Time `gorm:"column:log_date"`
	LogType        string    `gorm:"column:log_type"`
	FRTCreatedDate time.Time `gorm:"column:frt_created_date"`
}

func FetchFRTData(maxFetchID int) []FRTData {
	date := getTodayDate()
	month := date.Month
	year := date.Year

	var frtData []FRTData
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		log.Fatalf("failed to convert month to integer: %v", err)
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Fatalf("failed to convert year to integer: %v", err)
	}

	tableName := fmt.Sprintf("DeviceLogs_%d_%d", monthInt, yearInt)
	query := fmt.Sprintf(`SELECT TOP 20000 DeviceLogId frt_log_id, DeviceId device_id, UserId user_id, LogDate log_date, C1 log_type, CreatedDate frt_created_date FROM %s WHERE DeviceLogId > ? ORDER BY DeviceLogId`, tableName)
	err = AwsDB.Raw(query, maxFetchID).Scan(&frtData).Error
	if err != nil {
		log.Fatalf("failed to fetch FRT data: %v", err)
	}

	return frtData
}

func InsertFRTLogs(frtData []FRTData) {
	var wg sync.WaitGroup
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	for _, data := range frtData {
		wg.Add(1)
		go func(data FRTData) {
			defer wg.Done()
			data.LogDate = data.LogDate.In(loc)
			data.FRTCreatedDate = data.FRTCreatedDate.In(loc)
			err := ProgressionDB.Exec(`REPLACE INTO frt_logs (device_id, user_id, log_date, log_type, frt_created_date, frt_log_id) VALUES (?, ?, ?, ?, ?, ?)`, data.DeviceID, data.UserID, data.LogDate, data.LogType, data.FRTCreatedDate, data.FRTLogID).Error
			if err != nil {
				log.Fatalf("failed to insert FRT logs: %v", err)
			}
		}(data)
	}

	wg.Wait()
}
