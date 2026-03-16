package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// 查看 histcontest_register_competitions 表结构
	fmt.Println("=== histcontest_register_competitions 表结构 ===")
	rows, err := db.Query("DESCRIBE histcontest_register_competitions")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-20s %-20s %-20s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println("--------------------------------------------------------------------------------")

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err = rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Fatal(err)
		}
		defaultValStr := ""
		if defaultVal.Valid {
			defaultValStr = defaultVal.String
		}
		fmt.Printf("%-20s %-20s %-20s %-20s %-20s %-20s\n", field, fieldType, null.String, key, defaultValStr, extra)
	}

	fmt.Println("\n=== histcontest_register_registrations 表结构 ===")
	rows2, err := db.Query("DESCRIBE histcontest_register_registrations")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Printf("%-20s %-20s %-20s %-20s %-20s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println("--------------------------------------------------------------------------------")

	for rows2.Next() {
		var field, fieldType, null, key, extra sql.NullString
		var defaultVal sql.NullString
		err = rows2.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Fatal(err)
		}
		defaultValStr := ""
		if defaultVal.Valid {
			defaultValStr = defaultVal.String
		}
		fmt.Printf("%-20s %-20s %-20s %-20s %-20s %-20s\n", field.String, fieldType.String, null.String, key.String, defaultValStr, extra.String)
	}
}
