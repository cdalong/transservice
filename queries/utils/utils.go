package dbutils

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	//"strconv"
	//"strings"

	//"transaction_service/utils"
	"transaction_service/queries/models"
)

var db *sql.DB

func SetUtilsDB(database *sql.DB) {
	db = database
}

func GetQuoteServerURL() string {
    port := os.Getenv("QUOTE_SERVER_PORT")
    host := os.Getenv("QUOTE_SERVER_HOST")
    url := fmt.Sprintf("http://%s:%s", host, port)
    return string(url)
}

func QueryQuote(username string, stock string) (body []byte, err error) {
	URL := GetQuoteServerURL()
	log.Println(URL)
	res, err := http.Get(URL + "/api/getQuote/" + username + "/" + stock)

	if err != nil {
		return
	} else {
		body, err = ioutil.ReadAll(res.Body)
		log.Println(string(body))
	}

	return
}

func QueryUserAvailableBalance(username string) ( balance int, err error) {
	query := `SELECT (SELECT money FROM USERS WHERE username = $1) -
			 (SELECT COALESCE(SUM(amount), 0) FROM RESERVATIONS WHERE username = $1 and type = $2)
			 as available_balance;`
	err = db.QueryRow(query, username, models.BUY).Scan(&balance)
	return
}

func QueryUserAvailableShares(username string, symbol string) (shares int, err error) {
	query := `SELECT (SELECT COALESCE(SUM(shares), 0) FROM Stocks WHERE username = $1 and symbol = $2) -
			 (SELECT COALESCE(SUM(shares), 0) FROM RESERVATIONS WHERE username = $1 and type = $3);`
	err = db.QueryRow(query, username, symbol, models.SELL).Scan(&shares)
	return 
}

func QueryUser(username string) (user models.User, err error) {
	query := "SELECT uid, username, money FROM users WHERE username = $1"
	err = db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Money)
	return
}

func QueryUserStock(username string, symbol string) (stock models.Stock, err error) {
	query := "SELECT sid, username, symbol, shares FROM stocks WHERE username = $1 AND symbol = $2"
	err = db.QueryRow(query, username, symbol).Scan(&stock.ID, &stock.Username, &stock.Symbol, &stock.Shares)
	return 
}

func QueryStockTrigger(tid int64) (trig models.Trigger, err error) {
	query := "SELECT tid, username, symbol, type, amount, shares, trigger_price, executable, time FROM triggers WHERE tid = $1"
	err = db.QueryRow(query, tid).Scan(&trig.ID, &trig.Username, &trig.Symbol, 
						&trig.Order, &trig.Amount, &trig.Shares, &trig.TriggerPrice, &trig.Executable, &trig.Time)
	return 
}

func QueryUserTrigger(username string, symbol string, orderType models.OrderType) (trig models.Trigger, err error) {
	query := "SELECT tid, username, symbol, type, amount, shares, trigger_price, executable, time FROM triggers WHERE username = $1 AND symbol=$2 AND type=$3"
	err = db.QueryRow(query, username, symbol, orderType).Scan(&trig.ID, &trig.Username, &trig.Symbol, 
						&trig.Order, &trig.Amount, &trig.Shares, &trig.TriggerPrice, &trig.Executable, &trig.Time)
	return 
}

// func QueryAndExecuteCurrentTriggers() {
// 	query := `SELECT username, symbol, type, shares, amount, trigger_price 
// 				FROM triggers 
// 					WHERE trigger_price IS NOT NULL AND amount IS NOT NULL`

// 	rows, err := db.Query(query)

// 	if err != nil {
// 		return
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		var username string
// 		var symbol string
// 		var orderType string
// 		var shares sql.NullInt64
// 		var amount sql.NullFloat64
// 		var triggerValue sql.NullFloat64

// 		err := rows.Scan(&username, &symbol, &orderType, &shares, &amount, &triggerValue)
// 		if err != nil {
// 			utils.LogErr(err)
// 		}

// 		isSell := strings.Compare(orderType, "sell") == 0
// 		if (isSell && shares.Int64 > 0) || (!isSell && triggerValue.Float64 > 0) {
// 			log.Println("Executing trigger (username,stock):")
// 			log.Println(username)
// 			log.Println(symbol)
// 			quoteStr, err := QueryQuote(username, symbol)
// 			if err == nil {
// 				quote, _ := strconv.ParseFloat(strings.Split(string(quoteStr), ",")[0], 64)
// 				if quote <= triggerValue.Float64 {
// 					url := fmt.Sprintf("http://localhost:8888/api/executeTrigger/%s/%s/%d/%f/%f/%s", username, symbol, shares.Int64, amount.Float64, triggerValue.Float64, orderType)
// 					go http.Get(url)
// 				}
// 			} else {
// 				utils.LogErr(err)
// 			}
// 		}
// 	}

// 	return
// }

func QueryReservation(rid int64) (res models.Reservation, err error) {
	query := "SELECT rid, username, symbol, shares, amount, type, time FROM reservations WHERE rid=$1"
	err = db.QueryRow(query, rid).Scan(&res.ID, &res.Username, &res.Symbol, &res.Shares, &res.Amount, &res.Order, &res.Time)
	return
}

func QueryLastReservation(username string, resType models.OrderType) (res models.Reservation, err error) {
	query := "SELECT rid, username, symbol, shares, amount, type, time FROM reservations WHERE username=$1 and type=$2 ORDER BY (time) DESC, rid DESC LIMIT 1"
	err = db.QueryRow(query, username, resType).Scan(&res.ID, &res.Username, &res.Symbol, &res.Shares, &res.Amount, &res.Order, &res.Time)
	return
}
