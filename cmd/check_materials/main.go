package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	db, err := sql.Open("mysql", "root:n208966737@tcp(43.143.133.62:3306)/hoj")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询班级资料
	rows, err := db.Query("SELECT id, file_name, file_path, status FROM classroom_material ORDER BY id DESC LIMIT 50")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== 班级资料记录 ===")
	fmt.Println("ID\t文件名\t\t\t文件路径\t\t\t状态")
	fmt.Println("-------------------------------------------------------------------")

	count := 0
	for rows.Next() {
		var id int
		var fileName, filePath string
		var status int
		if err := rows.Scan(&id, &fileName, &filePath, &status); err != nil {
			log.Fatal(err)
		}
		count++
		statusStr := "✓"
		if status != 1 {
			statusStr = "×"
		}
		fmt.Printf("%d\t%s\t\t%s\t\t%d [%s]\n", id, fileName, filePath, status, statusStr)
	}

	fmt.Printf("\n总计: %d 条记录\n", count)

	// 统计
	var total, active, inactive int
	db.QueryRow("SELECT COUNT(*) FROM classroom_material").Scan(&total)
	db.QueryRow("SELECT COUNT(*) FROM classroom_material WHERE status = 1").Scan(&active)
	db.QueryRow("SELECT COUNT(*) FROM classroom_material WHERE status = 0").Scan(&inactive)

	fmt.Printf("\n统计信息:\n")
	fmt.Printf("- 总记录数: %d\n", total)
	fmt.Printf("- 活跃记录: %d\n", active)
	fmt.Printf("- 已删除记录: %d\n", inactive)
}
