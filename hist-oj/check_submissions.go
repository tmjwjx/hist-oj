package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func main() {
	// 连接数据库
	db, err := sql.Open("mysql", "root:hist2025@tcp(43.143.133.62:3306)/hoj")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// 查询最近的提交记录
	query := `
		SELECT id, homework_id, question_id, problem_id, uid, is_officially_submitted, create_time
		FROM homework_submit
		ORDER BY create_time DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer rows.Close()

	fmt.Println("=== 最近的提交记录 ===")
	fmt.Println("ID\t\t作业ID\t\t题目ID\t\t编程题ID\t\t用户ID\t已提交\t\t创建时间")
	fmt.Println("------------------------------------------------------------------------------------------------------------------")

	for rows.Next() {
		var id, homeworkID uint64
		var questionID sql.NullInt64
		var problemID sql.NullString
		var uid string
		var isOfficiallySubmitted int
		var createTime string

		err := rows.Scan(&id, &homeworkID, &questionID, &problemID, &uid, &isOfficiallySubmitted, &createTime)
		if err != nil {
			log.Fatal("扫描失败:", err)
		}

		questionIDStr := "NULL"
		if questionID.Valid {
			questionIDStr = fmt.Sprintf("%d", questionID.Int64)
		}

		problemIDStr := "NULL"
		if problemID.Valid {
			problemIDStr = problemID.String
		}

		fmt.Printf("%d\t%d\t%s\t%s\t%s\t%d\t\t%s\n",
			id, homeworkID, questionIDStr, problemIDStr, uid, isOfficiallySubmitted, createTime)
	}

	// 查询 homework_question 表，看看 problem_id 字段
	fmt.Println("\n=== 作业题目信息（homework_question） ===")
	query2 := `
		SELECT id, homework_id, question_id, problem_id
		FROM homework_question
		WHERE problem_id IS NOT NULL AND problem_id != ''
		LIMIT 5
	`

	rows2, err := db.Query(query2)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer rows2.Close()

	fmt.Println("ID\t\t作业ID\t\t题目ID\t\t编程题ID")
	fmt.Println("------------------------------------------------------------")

	for rows2.Next() {
		var id, homeworkID uint64
		var questionID sql.NullInt64
		var problemID sql.NullString

		err := rows2.Scan(&id, &homeworkID, &questionID, &problemID)
		if err != nil {
			log.Fatal("扫描失败:", err)
		}

		questionIDStr := "NULL"
		if questionID.Valid {
			questionIDStr = fmt.Sprintf("%d", questionID.Int64)
		}

		fmt.Printf("%d\t%d\t%s\t%s\n", id, homeworkID, questionIDStr, problemID.String)
	}
}
