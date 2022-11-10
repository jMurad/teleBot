package Database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TGDB struct {
	PoolConn *pgxpool.Pool
	lCl      []logClient
}

type logClient struct {
	firstname string
	userid    string
	msg       string
	datetime  time.Time
}

func (tgdb *TGDB) DBinit() {
	var err error

	tgdb.PoolConn, err = pgxpool.Connect(context.Background(), os.Getenv("PG_URL"))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to Database: %v\n", err)
		os.Exit(1)
	}
}

func (tgdb *TGDB) AddToLog(userid, firstname, msg string) {
	now := time.Now().Local()
	conn, err := tgdb.PoolConn.Acquire(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error acquiring connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(),
		`INSERT INTO log_client (firstname, userid, message, datetime) 
		VALUES ($1, $2, $3, $4::timestamptz)`,
		firstname,
		userid,
		msg,
		now,
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
}

func (tgdb *TGDB) UsersOfPeriod(beginTime, endTime time.Time) []logClient {
	conn, err := tgdb.PoolConn.Acquire(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error acquiring connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(),
		`SELECT firstname, userid, message
        FROM log_client 
        WHERE datetime>=$1::timestamptz AND datetime<=$2::timestamptz`,
		beginTime,
		endTime,
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var clRows []logClient
	for rows.Next() {
		clRow := logClient{}
		err := rows.Scan(&clRow.firstname, &clRow.userid, &clRow.msg)
		if err != nil {
			log.Fatal(err)
		}
		clRows = append(clRows, clRow)
		//fmt.Println("ct::",clRow)
	}
	return clRows
}

func (tgdb *TGDB) MsgUserOfPeriod(user string, beginTime, endTime time.Time) []logClient {
	conn, err := tgdb.PoolConn.Acquire(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error acquiring connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(),
		`SELECT firstname, userid, message
        FROM log_client 
        WHERE (firstname=$1 OR userid=$1) AND (datetime>=$2::timestamptz AND datetime<=$3::timestamptz)`,
		user,
		beginTime,
		endTime,
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var clRows []logClient
	for rows.Next() {
		clRow := logClient{}
		err := rows.Scan(&clRow.firstname, &clRow.userid, &clRow.msg)
		if err != nil {
			log.Fatal(err)
		}
		clRows = append(clRows, clRow)
		fmt.Println("ct::", clRow)
	}
	return clRows
}

func (tgdb *TGDB) ParseAdminMsg(msg string) bool {
	uop := regexp.MustCompile(`^T:\d\d?\.\d\d?\.\d\d\d\d-\d\d?\.\d\d?\.\d\d\d\d`)
	muop := regexp.MustCompile(`^T:\w+:\d\d?\.\d\d?\.\d\d\d\d-\d\d?\.\d\d?\.\d\d\d\d`)

	if uop.MatchString(msg) {
		prd1 := msg[2:strings.Index(msg, "-")]
		prd2 := msg[strings.Index(msg, "-")+1:]

		day, _ := strconv.Atoi(strings.Split(prd1, ".")[0])
		month, _ := strconv.Atoi(strings.Split(prd1, ".")[1])
		year, _ := strconv.Atoi(strings.Split(prd1, ".")[2])
		beginPeriod := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

		day, _ = strconv.Atoi(strings.Split(prd2, ".")[0])
		month, _ = strconv.Atoi(strings.Split(prd2, ".")[1])
		year, _ = strconv.Atoi(strings.Split(prd2, ".")[2])
		endPeriod := time.Date(year, time.Month(month), day, 23, 59, 59, 0, time.Local)

		tgdb.lCl = tgdb.UsersOfPeriod(beginPeriod, endPeriod)

		return true
	} else if muop.MatchString(msg) {

		prd1 := msg[strings.LastIndex(msg, ":")+1 : strings.Index(msg, "-")]
		prd2 := msg[strings.Index(msg, "-")+1:]

		user := msg[strings.Index(msg, ":")+1 : strings.LastIndex(msg, ":")]

		day, _ := strconv.Atoi(strings.Split(prd1, ".")[0])
		month, _ := strconv.Atoi(strings.Split(prd1, ".")[1])
		year, _ := strconv.Atoi(strings.Split(prd1, ".")[2])
		beginPeriod := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

		day, _ = strconv.Atoi(strings.Split(prd2, ".")[0])
		month, _ = strconv.Atoi(strings.Split(prd2, ".")[1])
		year, _ = strconv.Atoi(strings.Split(prd2, ".")[2])
		endPeriod := time.Date(year, time.Month(month), day, 23, 59, 59, 0, time.Local)

		tgdb.lCl = tgdb.MsgUserOfPeriod(user, beginPeriod, endPeriod)
		return true
	}
	return false
}

func (tgdb *TGDB) PrettyLog() string {
	var out string
	for _, l := range tgdb.lCl {
		out += fmt.Sprintf("*%s*|%s|: %s\n\n", l.firstname, l.userid, l.msg)
	}
	return out
}
