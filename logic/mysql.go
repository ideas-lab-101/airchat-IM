package main

//mysql -h 121.42.237.244 -u root -p
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var (
	dbhostsip  = "gz-cdb-09xb7bmb.sql.tencentcdb.com:63109" //IP地址
	dbusername = "root"                                     //用户名
	dbpassword = "Sagacity@db2"                             //密码
	dbname     = "engram"                                   //表名
)

//******* 私信处理
func authIMUser(userName string, password string) (userId int64, result bool, err error) {
	mysqlInfo := dbusername + ":" + dbpassword + "@tcp(" + dbhostsip + ")/" + dbname + "?charset=utf8"
	// fmt.Println("mysqlInfo：", mysqlInfo)

	db, err := sql.Open("mysql", mysqlInfo)
	if err != nil {
		fmt.Println("mysql连接错误：", err)
		result = false
	}

	defer db.Close()

	err = db.QueryRow("SELECT id FROM user_login WHERE login_name = ? AND password = ?", userName, password).Scan(&userId)

	// fmt.Println("err = ", err, "userId = ", userId)

	if err == nil {
		result = true
	}

	return
}

func getUserChatId(userAccount string) (userChatID int64, err error) {
	mysqlInfo := dbusername + ":" + dbpassword + "@tcp(" + dbhostsip + ")/" + dbname + "?charset=utf8"
	// fmt.Println("mysqlInfo：", mysqlInfo)

	db, err := sql.Open("mysql", mysqlInfo)
	if err != nil {
		fmt.Println("mysql连接错误：", err)
	}

	defer db.Close()

	err = db.QueryRow("SELECT id FROM user_login WHERE login_name = ? ", userAccount).Scan(&userChatID)

	return
}

//********* 推送的处理
func getUserDeviceToken(userAccount string) (deviceToken, voiceSetting string, err error) {
	mysqlInfo := dbusername + ":" + dbpassword + "@tcp(" + dbhostsip + ")/" + dbname + "?charset=utf8"
	// fmt.Println("mysqlInfo：", mysqlInfo)

	db, err := sql.Open("mysql", mysqlInfo)
	if err != nil {
		fmt.Println("mysql连接错误：", err)
	}

	defer db.Close()

	// err = db.QueryRow("select ud.token from user_login ul left join user_device ud on ud.user_id=ul.id where ul.login_name = ? and ud.state = 1", userAccount).Scan(&deviceToken)
	err = db.QueryRow("select ud.token,ui.voice_settings from user_login ul left join user_device ud on ud.user_id=ul.id left join user_info ui on ul.id=ui.user_id where ul.login_name = ? and ud.state = 1", userAccount).Scan(&deviceToken, &voiceSetting)

	return
}
