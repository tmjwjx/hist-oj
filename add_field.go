package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	db, err := sql.Open("mysql", "root:hist2025@tcp(43.143.133.62:3306)/hoj")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// 检查字段是否存在
	var columnExists int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'hoj'
		AND TABLE_NAME = 'homework_submit'
		AND COLUMN_NAME = 'judge_result'
	`).Scan(&columnExists)

	if err != nil {
		fmt.Printf("查询字段失败: %v\n", err)
		return
	}

	if columnExists > 0 {
		fmt.Println("字段 judge_result 已存在，跳过添加")
	} else {
		// 添加字段
		_, err = db.Exec(`
			ALTER TABLE homework_submit
			ADD COLUMN judge_result VARCHAR(50) DEFAULT NULL
			COMMENT '评测结果（编程题）: AC/WA/CE/TLE/MLE/RE等'
			AFTER is_officially_submitted
		`)
		if err != nil {
			fmt.Printf("添加字段失败: %v\n", err)
			return
		}
		fmt.Println("✅ 字段 judge_result 添加成功")
	}

	// 检查索引是否存在
	var indexExists int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = 'hoj'
		AND TABLE_NAME = 'homework_submit'
		AND INDEX_NAME = 'idx_judge_result'
	`).Scan(&indexExists)

	if err != nil {
		fmt.Printf("查询索引失败: %v\n", err)
		return
	}

	if indexExists > 0 {
		fmt.Println("索引 idx_judge_result 已存在，跳过添加")
	} else {
		// 添加索引
		_, err = db.Exec(`
			CREATE INDEX idx_judge_result ON homework_submit (judge_result)
		`)
		if err != nil {
			fmt.Printf("添加索引失败: %v\n", err)
			return
		}
		fmt.Println("✅ 索引 idx_judge_result 添加成功")
	}

	// 验证
	var columnType, isNullable, columnDefault, columnComment string
	err = db.QueryRow(`
		SELECT COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'hoj'
		AND TABLE_NAME = 'homework_submit'
		AND COLUMN_NAME = 'judge_result'
	`).Scan(&columnType, &isNullable, &columnDefault, &columnComment)

	if err != nil {
		fmt.Printf("验证字段失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 字段信息 ===")
	fmt.Printf("字段名: judge_result\n")
	fmt.Printf("类型: %s\n", columnType)
	fmt.Printf("可空: %s\n", isNullable)
	fmt.Printf("默认值: %s\n", columnDefault)
	fmt.Printf("注释: %s\n", columnComment)
	fmt.Println("\n✅ 所有操作完成！")
}
