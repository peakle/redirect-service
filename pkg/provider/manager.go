package provider

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	idgen "github.com/wakeapp/go-id-generator"
	sg "github.com/wakeapp/go-sql-generator"
	"os"
	"runtime"
	"time"
)

type SQLManager struct {
	conn *sql.DB
}

type config struct {
	Host     string
	UserPass string
	Port     string
	DBName   string
}

type StatResponse struct {
	Useragent string `json:"useragent"`
	Ip        string `json:"ip"`
	City      string `json:"city"`
	CreatedAt string `json:"created_at"`
}

func (m *SQLManager) RecordStats(ip, useragent, city string) {
	const TableName = "stats"

	now := time.Now().Format("2006-01-02")

	dataInsert := sg.InsertData{
		TableName: TableName,
		Fields: []string{
			"id",
			"useragent",
			"ip",
			"city",
			"created_at",
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

func (m *SQLManager) FindUrlByTokenAndUserId(userId, token string) ([]StatResponse, error) {
	query := `
		SELECT
			s.useragent,
			s.ip,
			s.city,
			s.created_at
		FROM stats s
		JOIN redirects r on r.token = s.token 
		WHERE 1
			and r.token = ?
			and r.user_id = ?
	`

	rows, err := m.conn.Query(query, token, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return []StatResponse{}, nil
		}

		return []StatResponse{}, err
	}

	var resp []StatResponse
	for rows.Next() {
		var r StatResponse
		err = rows.Scan(&r.Useragent, &r.Ip, &r.City, &r.CreatedAt)
		if err != nil {
			return []StatResponse{}, err
		}

		resp = append(resp, r)
	}

	return resp, nil
}

func (m *SQLManager) FindUrl(token string) (string, error) {
	query := `
		SELECT r.url
		FROM redirects r
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

func (m *SQLManager) InsertToken(userId, uri, token string) error {
	if isEmpty(userId) || isEmpty(uri) || isEmpty(token) {
		return errors.New(fmt.Sprintf("null param provided: userId %s, uri %s, token %s", userId, uri, token))
	}

	const TableName = "redirects"

	data := &sg.InsertData{
		TableName: TableName,
		Fields: []string{
			"token",
			"url",
			"user_id",
			"created_at",
		},
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	data.Add([]string{
		idgen.Id(),
		token,
		uri,
		now,
	})
	m.insert(data)

	return nil
}

func (m *SQLManager) TokenExist(token string) bool {
	query := `
	   	SELECT COUNT(*)
		FROM redirects r
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

func InitManager(userPass string) *SQLManager {
	m := &SQLManager{}

	m.open(&config{
		Host:     os.Getenv("MYSQL_HOST"),
		UserPass: userPass,
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
		runtime.Goexit()
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
