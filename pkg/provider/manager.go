package provider

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	idgen "github.com/wakeapp/go-id-generator"
	sg "github.com/wakeapp/go-sql-generator"
	"os"
	"time"
)

type SQLManager struct {
	conn *sql.DB
}

type config struct {
	Host     string
	Username string
	Pass     string
	Port     string
	DBName   string
}

func (m *SQLManager) RecordStats(ip, useragent, city string) {
	now := time.Now().Format("2006-01-02")

	dataInsert := sg.InsertData{
		TableName: "Stats",
		Fields: []string{
			"id",
			"useragent",
			"ip",
			"city",
			"createAt",
		},
		IsIgnore: true,
	}

	dataInsert.Add([]string{
		idgen.Id(),
		useragent,
		ip,
		city,
		now,
	})

	m.insert(&dataInsert)
}

func (m *SQLManager) FindUrl(token string) (string, error) {
	query := `
		SELECT r.url
		FROM Redirects r
		WHERE 1
			and r.token = ?
		LIMIT 1
	`

	row := m.conn.QueryRow(query, token)

	var r string
	err := row.Scan(&r)
	if err != nil {
		if err == sql.ErrNoRows {
			return "https://ya.ru", nil
		}

		return "", err
	}

	return r, nil
}

func (m *SQLManager) TokenExist(token string) bool {
	query := `
	   	SELECT COUNT(*)
		FROM Redirects r
		WHERE 1
			AND token = ?
	`

	row := m.conn.QueryRow(query, token)

	var r int
	err := row.Scan(&r)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}

		handleErr(err)
	}

	return r > 0
}

func InitManager() *SQLManager {
	m := &SQLManager{}

	m.open(&config{
		Host:     os.Getenv("MYSQL_HOST"),
		Username: os.Getenv("MYSQL_USER"),
		Pass:     os.Getenv("MYSQL_PASSWORD"),
		Port:     "3306",
		DBName:   os.Getenv("MYSQL_DATABASE"),
	})

	return m
}

func (m *SQLManager) insert(dataInsert *sg.InsertData) int {
	if len(dataInsert.ValuesList) == 0 {
		return 0
	}

	sqlGenerator := sg.MysqlSqlGenerator{}

	query, args, err := sqlGenerator.GetInsertSql(*dataInsert)
	handleErr(err)

	var stmt *sql.Stmt
	stmt, err = m.conn.Prepare(query)
	handleErr(err)

	var result sql.Result
	result, err = stmt.Exec(args...)
	handleErr(err)

	ra, _ := result.RowsAffected()

	return int(ra)
}

func (m *SQLManager) Close() {
	_ = m.conn.Close()
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (m *SQLManager) open(c *config) {
	var conn *sql.DB
	var err error
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?collation=utf8_unicode_ci", c.Username, c.Pass, c.Host, c.Port, c.DBName)
	if conn, err = sql.Open("mysql", dsn); err != nil {
		handleErr(err)
	}

	m.conn = conn
}
