package provider

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	idgen "github.com/wakeapp/go-id-generator"
	sg "github.com/wakeapp/go-sql-generator"
)

// SQLManager db manager provide ability to call SQL methods
type SQLManager struct {
	conn *sql.DB
}

type config struct {
	Host     string
	UserPass string
	Port     string
	DBName   string
}

// StatResponse response for `stats` route
type StatResponse struct {
	Useragent string `json:"useragent"`
	IP        string `json:"ip"`
	City      string `json:"city"`
	Date      string `json:"date"`
	Count     string `json:"count"`
}

var managers sync.Map

// RecordStats about entry
func (m *SQLManager) RecordStats(ip, useragent, city, token string) {
	const TableName = "stats"

	now := time.Now().Format("2006-01-02")

	dataInsert := &sg.InsertData{
		TableName: TableName,
		Fields: []string{
			"id",
			"useragent",
			"ip",
			"city",
			"token",
			"created_at",
		},
		IsIgnore: true,
	}

	dataInsert.Add([]string{
		idgen.Id(),
		useragent,
		ip,
		city,
		token,
		now,
	})

	m.insert(dataInsert)
}

// FindURLByTokenAndUserID export statistic about entries
func (m *SQLManager) FindURLByTokenAndUserID(userID, token string) ([]StatResponse, error) {
	query := `
		SELECT
			s.useragent,
			s.ip,
			s.city,
		    COUNT(*) as count
		FROM stats s
		JOIN redirects r on r.token = s.token 
		WHERE 1
			and r.token = ?
			and r.user_id = ?
		GROUP BY s.city
	`

	rows, err := m.conn.Query(query, token, userID)
	if err != nil {
		return []StatResponse{}, err
	}

	var resp []StatResponse
	for rows.Next() {
		var r StatResponse
		err = rows.Scan(&r.Useragent, &r.IP, &r.City, &r.Count)
		if err != nil {
			return []StatResponse{}, err
		}

		resp = append(resp, r)
	}

	return resp, nil
}

// FindURLByToken find redirect url by token
func (m *SQLManager) FindURLByToken(token string) (string, error) {
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
		return "", err
	}

	return r, nil
}

// InsertToken insert redirect url and token
func (m *SQLManager) InsertToken(userID, uri, token string) error {
	if isEmpty(userID) || isEmpty(uri) || isEmpty(token) {
		return fmt.Errorf("null param provided: userID %s, uri %s, token %s", userID, uri, token)
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
		token,
		uri,
		userID,
		now,
	})

	if len(data.ValuesList) != m.insert(data) {
		return errors.New("unavailable to insert token")
	}

	return nil
}

// TokenExist check token for exist
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
		return false
	}

	return r > 0
}

// InitManager create db manager
func InitManager(userPass string) *SQLManager {
	if m, ok := managers.Load(userPass); ok {
		return m.(*SQLManager)
	}

	m := &SQLManager{}

	m.open(&config{
		Host:     os.Getenv("MYSQL_HOST"),
		UserPass: userPass,
		Port:     "3306",
		DBName:   os.Getenv("MYSQL_DATABASE"),
	})

	managers.Store(userPass, m)

	go func() {
		timeout := time.NewTicker(time.Second * 120)

		for {
			select {
			case <-timeout.C:
				m.Ping()
			}
		}
	}()

	return m
}

// Close db manager
func (m *SQLManager) Close() {
	_ = m.conn.Close()
}

// Ping - ping and reset connection with DB
func (m *SQLManager) Ping() {
	err := m.conn.Ping()

	if err != nil {
		handleErr(err)
	}
}

func (m *SQLManager) insert(dataInsert *sg.InsertData) int {
	if len(dataInsert.ValuesList) == 0 {
		return 0
	}

	sqlGenerator := sg.MysqlSqlGenerator{}

	query, args, err := sqlGenerator.GetInsertSql(*dataInsert)
	if err != nil {
		handleErr(err)
		return 0
	}

	var stmt *sql.Stmt
	stmt, err = m.conn.Prepare(query)
	if err != nil {
		handleErr(err)
		return 0
	}

	var result sql.Result
	result, err = stmt.Exec(args...)
	if err != nil {
		handleErr(err)
		return 0
	}

	ra, err := result.RowsAffected()
	if err != nil {
		handleErr(err)
	}

	return int(ra)
}

func handleErr(err error) {
	fmt.Printf("error occurred: " + err.Error())
}

func (m *SQLManager) open(c *config) {
	var conn *sql.DB
	var err error
	dsn := fmt.Sprintf("%v@tcp(%v:%v)/%v?collation=utf8_unicode_ci", c.UserPass, c.Host, c.Port, c.DBName)
	if conn, err = sql.Open("mysql", dsn); err != nil {
		handleErr(err)
		os.Exit(1)
	}

	m.conn = conn
}
