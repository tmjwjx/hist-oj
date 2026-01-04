package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var useMySQL = true // 改为 true 使用 MySQL

// 数据库模型
type Competition struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Fields      string    `json:"fields"`    // JSON字符串
	LogoURL     string    `json:"logo_url"`  // Logo图片URL
	Description string    `json:"description"` // 比赛说明（HTML格式）
	Visible     bool      `json:"visible"`
	CreatedAt   time.Time `json:"created_at"`
}

type Registration struct {
	ID            int       `json:"id"`
	CompetitionID int       `json:"competition_id"`
	UserUUID      string    `json:"user_uuid"` // 关联用户
	Name          string    `json:"name"`
	Class         string    `json:"class"`
	College       string    `json:"college"`
	StudentID     string    `json:"student_id"`
	Gender        string    `json:"gender"`
	ShirtSize     string    `json:"shirt_size"`
	TeamName      string    `json:"team_name"`
	QQ            string    `json:"qq"`
	Status        string    `json:"status"` // "pending", "approved", "rejected"
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
}

type FieldConfig struct {
	Name      bool `json:"name"`
	Class     bool `json:"class"`
	College   bool `json:"college"`
	StudentID bool `json:"student_id"`
	Gender    bool `json:"gender"`
	ShirtSize bool `json:"shirt_size"`
	TeamName  bool `json:"team_name"`
	QQ        bool `json:"qq"`
}

// 用户信息
type UserInfo struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func initDB() {
	var err error

	// 尝试连接 MySQL
	if !useMySQL {
		// 先尝试 SQLite（本地测试）
		db, err = sql.Open("sqlite3", "./registration.db")
		if err != nil {
			log.Fatal(err)
		}

		// 测试 SQLite 连接
		err = db.Ping()
		if err == nil {
			fmt.Println("✅ 使用 SQLite 本地数据库")
			initSQLiteTables()
			return
		}
	}

	// 使用 MySQL
	useMySQL = true
	// 指定 loc=Asia/Shanghai 确保时间使用中国时区
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?parseTime=true&loc=Asia%2FShanghai"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatal("无法连接到 MySQL 数据库:", err)
	}

	fmt.Println("✅ 成功连接到 MySQL 数据库")
	initMySQLTables()
}

func initSQLiteTables() {
	// 创建比赛表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS competitions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			fields TEXT NOT NULL,
			logo_url TEXT,
			visible BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 创建报名表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS registrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			competition_id INTEGER NOT NULL,
			name TEXT,
			class TEXT,
			college TEXT,
			student_id TEXT,
			gender TEXT,
			shirt_size TEXT,
			team_name TEXT,
			qq TEXT,
			status TEXT DEFAULT 'pending',
			remark TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (competition_id) REFERENCES competitions(id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ SQLite 数据库表检查完成")
}

func initMySQLTables() {
	fmt.Println("正在创建 MySQL 表...")

	// 创建比赛表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS histcontest_register_competitions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			fields TEXT NOT NULL,
			logo_url TEXT,
			visible BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("创建比赛表失败:", err)
	}
	fmt.Println("✓ 比赛表已创建或已存在")

	// 如果表已存在，添加 visible 字段（兼容旧数据）
	_, err = db.Exec(`
		ALTER TABLE histcontest_register_competitions
		ADD COLUMN IF NOT EXISTS visible BOOLEAN DEFAULT TRUE
	`)
	if err != nil {
		// 忽略字段已存在的错误
		fmt.Println("Note: visible字段可能已存在")
	}

	// 如果表已存在，添加 logo_url 字段（兼容旧数据）
	// 先检查字段是否存在
	columns, _ := db.Query("SHOW COLUMNS FROM histcontest_register_competitions LIKE 'logo_url'")
	hasLogoColumn := false
	if columns != nil {
		defer columns.Close()
		hasLogoColumn = columns.Next()
	}

	if !hasLogoColumn {
		_, err = db.Exec(`
			ALTER TABLE histcontest_register_competitions
			ADD COLUMN logo_url TEXT
		`)
		if err != nil {
			fmt.Println("添加 logo_url 字段:", err)
		} else {
			fmt.Println("✓ logo_url 字段已添加")
			// 为旧数据设置默认值
			_, err = db.Exec(`
				UPDATE histcontest_register_competitions SET logo_url = '' WHERE logo_url IS NULL
			`)
			if err != nil {
				fmt.Println("更新旧数据 logo_url 默认值:", err)
			} else {
				fmt.Println("✓ 已为旧数据设置 logo_url 默认值")
			}
		}
	} else {
		fmt.Println("✓ logo_url 字段已存在")
		// 为已存在但为NULL的旧数据设置默认值
		result, err := db.Exec(`
			UPDATE histcontest_register_competitions SET logo_url = '' WHERE logo_url IS NULL
		`)
		if err != nil {
			fmt.Println("更新旧数据 logo_url 默认值:", err)
		} else {
			rows, _ := result.RowsAffected()
			if rows > 0 {
				fmt.Printf("✓ 已为 %d 条旧数据设置 logo_url 默认值\n", rows)
			}
		}
	}

	// 如果表已存在，添加 description 字段（兼容旧数据）
	// 先检查字段是否存在
	columns, _ = db.Query("SHOW COLUMNS FROM histcontest_register_competitions LIKE 'description'")
	hasDescriptionColumn := false
	if columns != nil {
		defer columns.Close()
		hasDescriptionColumn = columns.Next()
	}

	if !hasDescriptionColumn {
		_, err = db.Exec(`
			ALTER TABLE histcontest_register_competitions
			ADD COLUMN description TEXT
		`)
		if err != nil {
			fmt.Println("添加 description 字段:", err)
		} else {
			fmt.Println("✓ description 字段已添加")
			// 为旧数据设置默认值
			_, err = db.Exec(`
				UPDATE histcontest_register_competitions SET description = '' WHERE description IS NULL
			`)
			if err != nil {
				fmt.Println("更新旧数据 description 默认值:", err)
			} else {
				fmt.Println("✓ 已为旧数据设置 description 默认值")
			}
		}
	} else {
		fmt.Println("✓ description 字段已存在")
		// 为已存在但为NULL的旧数据设置默认值
		result, err := db.Exec(`
			UPDATE histcontest_register_competitions SET description = '' WHERE description IS NULL
		`)
		if err != nil {
			fmt.Println("更新旧数据 description 默认值:", err)
		} else {
			rows, _ := result.RowsAffected()
			if rows > 0 {
				fmt.Printf("✓ 已为 %d 条旧数据设置 description 默认值\n", rows)
			}
		}
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS histcontest_register_registrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			competition_id INT NOT NULL,
			user_uuid VARCHAR(100),
			name VARCHAR(100),
			class VARCHAR(100),
			college VARCHAR(100),
			student_id VARCHAR(50),
			gender VARCHAR(10),
			shirt_size VARCHAR(10),
			team_name VARCHAR(100),
			qq VARCHAR(20),
			status VARCHAR(20) DEFAULT 'pending',
			remark TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (competition_id) REFERENCES histcontest_register_competitions(id)
		)
	`)
	if err != nil {
		log.Fatal("创建报名表失败:", err)
	}
	fmt.Println("✓ 报名表已创建或已存在")

	// 如果表已存在，添加 user_uuid 字段（兼容旧数据）
	// MySQL 不支持 IF NOT EXISTS，需要先检查
	rows, _ := db.Query("SHOW COLUMNS FROM histcontest_register_registrations LIKE 'user_uuid'")
	hasColumn := false
	if rows != nil {
		defer rows.Close()
		hasColumn = rows.Next()
	}

	if !hasColumn {
		_, err = db.Exec(`
			ALTER TABLE histcontest_register_registrations
			ADD COLUMN user_uuid VARCHAR(100)
		`)
		if err != nil {
			fmt.Println("添加 user_uuid 字段失败:", err)
		} else {
			fmt.Println("✓ user_uuid 字段已添加")
		}
	} else {
		fmt.Println("✓ user_uuid 字段已存在")
	}

	// 检查表是否真的存在
	tableRows, _ := db.Query("SHOW TABLES LIKE 'histcontest_register%'")
	defer tableRows.Close()
	fmt.Println("当前数据库中的表:")
	for tableRows.Next() {
		var tableName string
		tableRows.Scan(&tableName)
		fmt.Println("  -", tableName)
	}

	fmt.Println("✅ MySQL 数据库表检查完成")
}

// 获取表名
func getCompetitionTable() string {
	if useMySQL {
		return "histcontest_register_competitions"
	}
	return "competitions"
}

func getRegistrationTable() string {
	if useMySQL {
		return "histcontest_register_registrations"
	}
	return "registrations"
}

// CORS 中间件
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// 全局 CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 创建比赛
func createCompetition(w http.ResponseWriter, r *http.Request) {
	var comp Competition
	err := json.NewDecoder(r.Body).Decode(&comp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	compTable := getCompetitionTable()

	result, err := db.Exec(
		"INSERT INTO "+compTable+" (name, start_time, end_time, fields, logo_url, description, visible) VALUES (?, ?, ?, ?, ?, ?, ?)",
		comp.Name, comp.StartTime, comp.EndTime, comp.Fields, comp.LogoURL, comp.Description, comp.Visible,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	comp.ID = int(id)

	// 使用中国时区获取当前时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	comp.CreatedAt = time.Now().In(loc)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comp)
}

// 更新比赛
func updateCompetition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	competitionID := params["id"]

	var comp Competition
	err := json.NewDecoder(r.Body).Decode(&comp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	compTable := getCompetitionTable()

	_, err = db.Exec(
		"UPDATE "+compTable+" SET name = ?, start_time = ?, end_time = ?, fields = ?, logo_url = ?, description = ? WHERE id = ?",
		comp.Name, comp.StartTime, comp.EndTime, comp.Fields, comp.LogoURL, comp.Description, competitionID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 查询并返回更新后的比赛
	err = db.QueryRow(
		"SELECT id, name, start_time, end_time, fields, logo_url, description, visible, created_at FROM "+compTable+" WHERE id = ?",
		competitionID,
	).Scan(&comp.ID, &comp.Name, &comp.StartTime, &comp.EndTime, &comp.Fields, &comp.LogoURL, &comp.Description, &comp.Visible, &comp.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comp)
}

// 获取所有比赛
func getCompetitions(w http.ResponseWriter, r *http.Request) {
	showHidden := r.URL.Query().Get("show_hidden") == "true"
	compTable := getCompetitionTable()

	var rows *sql.Rows
	var err error

	if showHidden {
		rows, err = db.Query("SELECT id, name, start_time, end_time, fields, logo_url, description, visible, created_at FROM "+compTable+" ORDER BY created_at DESC")
	} else {
		if useMySQL {
			rows, err = db.Query("SELECT id, name, start_time, end_time, fields, logo_url, description, visible, created_at FROM "+compTable+" WHERE visible = true ORDER BY created_at DESC")
		} else {
			rows, err = db.Query("SELECT id, name, start_time, end_time, fields, logo_url, description, visible, created_at FROM "+compTable+" WHERE visible = 1 ORDER BY created_at DESC")
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var competitions []Competition
	for rows.Next() {
		var comp Competition
		err := rows.Scan(&comp.ID, &comp.Name, &comp.StartTime, &comp.EndTime, &comp.Fields, &comp.LogoURL, &comp.Description, &comp.Visible, &comp.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, comp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(competitions)
}

// 获取单个比赛
func getCompetition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	compTable := getCompetitionTable()

	var comp Competition
	err := db.QueryRow(
		"SELECT id, name, start_time, end_time, fields, logo_url, description, visible, created_at FROM "+compTable+" WHERE id = ?",
		id,
	).Scan(&comp.ID, &comp.Name, &comp.StartTime, &comp.EndTime, &comp.Fields, &comp.LogoURL, &comp.Description, &comp.Visible, &comp.CreatedAt)

	if err != nil {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comp)
}

// 创建报名
func createRegistration(w http.ResponseWriter, r *http.Request) {
	var reg Registration
	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证 user_uuid
	if reg.UserUUID == "" {
		http.Error(w, "缺少 user_uuid", http.StatusBadRequest)
		return
	}

	// 检查比赛时间是否允许报名
	compTable := getCompetitionTable()
	var comp Competition
	err = db.QueryRow(
		"SELECT id, name, start_time, end_time FROM "+compTable+" WHERE id = ?",
		reg.CompetitionID,
	).Scan(&comp.ID, &comp.Name, &comp.StartTime, &comp.EndTime)

	if err != nil {
		http.Error(w, "比赛不存在", http.StatusNotFound)
		return
	}

	now := time.Now()
	if now.Before(comp.StartTime) {
		http.Error(w, "比赛尚未开始报名", http.StatusBadRequest)
		return
	}
	if now.After(comp.EndTime) {
		http.Error(w, "报名时间已截止", http.StatusBadRequest)
		return
	}

	reg.Status = "pending"

	// 使用中国时区获取当前时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	reg.CreatedAt = time.Now().In(loc)

	regTable := getRegistrationTable()

	result, err := db.Exec(
		`INSERT INTO `+regTable+` (competition_id, user_uuid, name, class, college, student_id, gender, shirt_size, team_name, qq, status, remark)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reg.CompetitionID, reg.UserUUID, reg.Name, reg.Class, reg.College, reg.StudentID,
		reg.Gender, reg.ShirtSize, reg.TeamName, reg.QQ, reg.Status, reg.Remark,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	reg.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reg)
}

// 获取比赛的所有报名
func getRegistrations(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	competitionID := params["competitionId"]
	regTable := getRegistrationTable()

	rows, err := db.Query(
		`SELECT id, competition_id, user_uuid, name, class, college, student_id, gender, shirt_size, team_name, qq, status, remark, created_at
		FROM `+regTable+` WHERE competition_id = ? ORDER BY created_at DESC`,
		competitionID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var registrations []Registration
	for rows.Next() {
		var reg Registration
		err := rows.Scan(&reg.ID, &reg.CompetitionID, &reg.UserUUID, &reg.Name, &reg.Class, &reg.College,
			&reg.StudentID, &reg.Gender, &reg.ShirtSize, &reg.TeamName, &reg.QQ, &reg.Status, &reg.Remark, &reg.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		registrations = append(registrations, reg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registrations)
}

// 获取用户在某个比赛的报名
func getUserRegistration(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	competitionID := params["competitionId"]
	userUUID := r.URL.Query().Get("user_uuid")
	regTable := getRegistrationTable()

	if userUUID == "" {
		http.Error(w, "缺少 user_uuid 参数", http.StatusBadRequest)
		return
	}

	var reg Registration
	err := db.QueryRow(
		`SELECT id, competition_id, user_uuid, name, class, college, student_id, gender, shirt_size, team_name, qq, status, remark, created_at
		FROM `+regTable+` WHERE competition_id = ? AND user_uuid = ?`,
		competitionID, userUUID,
	).Scan(&reg.ID, &reg.CompetitionID, &reg.UserUUID, &reg.Name, &reg.Class, &reg.College,
		&reg.StudentID, &reg.Gender, &reg.ShirtSize, &reg.TeamName, &reg.QQ, &reg.Status, &reg.Remark, &reg.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(nil)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reg)
}

// 更新报名状态
func updateRegistration(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	regTable := getRegistrationTable()

	var update Registration
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取当前报名状态，检查是否是被退回的报名
	var currentReg Registration
	err = db.QueryRow(
		"SELECT id, status, competition_id FROM "+regTable+" WHERE id = ?",
		id,
	).Scan(&currentReg.ID, &currentReg.Status, &currentReg.CompetitionID)

	if err != nil {
		http.Error(w, "报名记录不存在", http.StatusNotFound)
		return
	}

	// 如果不是被退回的报名，需要检查比赛时间
	if currentReg.Status != "rejected" {
		compTable := getCompetitionTable()
		var comp Competition
		err = db.QueryRow(
			"SELECT id, name, start_time, end_time FROM "+compTable+" WHERE id = ?",
			currentReg.CompetitionID,
		).Scan(&comp.ID, &comp.Name, &comp.StartTime, &comp.EndTime)

		if err != nil {
			http.Error(w, "比赛不存在", http.StatusNotFound)
			return
		}

		now := time.Now()
		if now.Before(comp.StartTime) {
			http.Error(w, "比赛尚未开始报名", http.StatusBadRequest)
			return
		}
		if now.After(comp.EndTime) {
			http.Error(w, "报名时间已截止", http.StatusBadRequest)
			return
		}
	}

	// 构建动态更新SQL
	query := "UPDATE " + regTable + " SET "
	args := []interface{}{}
	updateFields := []string{}

	if update.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, update.Name)
	}
	if update.Class != "" {
		updateFields = append(updateFields, "class = ?")
		args = append(args, update.Class)
	}
	if update.College != "" {
		updateFields = append(updateFields, "college = ?")
		args = append(args, update.College)
	}
	if update.StudentID != "" {
		updateFields = append(updateFields, "student_id = ?")
		args = append(args, update.StudentID)
	}
	if update.Gender != "" {
		updateFields = append(updateFields, "gender = ?")
		args = append(args, update.Gender)
	}
	if update.ShirtSize != "" {
		updateFields = append(updateFields, "shirt_size = ?")
		args = append(args, update.ShirtSize)
	}
	if update.TeamName != "" {
		updateFields = append(updateFields, "team_name = ?")
		args = append(args, update.TeamName)
	}
	if update.QQ != "" {
		updateFields = append(updateFields, "qq = ?")
		args = append(args, update.QQ)
	}
	if update.Status != "" {
		updateFields = append(updateFields, "status = ?")
		args = append(args, update.Status)
	}
	if update.Remark != "" {
		updateFields = append(updateFields, "remark = ?")
		args = append(args, update.Remark)
	}

	if len(updateFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query += strings.Join(updateFields, ", ")
	query += " WHERE id = ?"
	args = append(args, id)

	_, err = db.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration updated successfully"})
}

// 用户登录
func userLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 从 user_info 表查询用户
	var user UserInfo
	err = db.QueryRow(
		"SELECT uuid, username, password FROM user_info WHERE username = ?",
		loginReq.Username,
	).Scan(&user.UUID, &user.Username, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 验证密码（MD5加密后比较）
	hash := md5.Sum([]byte(loginReq.Password))
	hashedPassword := hex.EncodeToString(hash[:])

	if user.Password != hashedPassword {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}

	// 登录成功，返回用户信息（不包含密码）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uuid":     user.UUID,
		"username": user.Username,
		"message":  "登录成功",
	})
}

// HOJ用户自动登录
func hojAutoLogin(w http.ResponseWriter, r *http.Request) {
	// 从URL参数或Header中获取token
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
	}

	// 从URL参数获取username
	username := r.URL.Query().Get("username")

	if token == "" || username == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 400,
			"message": "缺少token或username参数",
		})
		return
	}

	// 验证用户是否存在,如果不存在则自动创建
	var user UserInfo
	err := db.QueryRow(
		"SELECT uuid, username FROM user_info WHERE username = ?",
		username,
	).Scan(&user.UUID, &user.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在,自动创建用户
			uuid := generateUUID()
			_, err = db.Exec(
				"INSERT INTO user_info (uuid, username, password) VALUES (?, ?, ?)",
				uuid, username, "",
			)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": 500,
					"message": "创建用户失败",
				})
				return
			}
			user.UUID = uuid
			user.Username = username
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": 500,
				"message": "数据库查询错误",
			})
			return
		}
	}

	// 登录成功,返回用户信息
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": 200,
		"data": map[string]string{
			"uuid":     user.UUID,
			"username": user.Username,
		},
		"message": "HOJ自动登录成功",
	})
}

// 生成UUID
func generateUUID() string {
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash.Sum(nil))[:32]
}

// 删除比赛
func deleteCompetition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	regTable := getRegistrationTable()
	compTable := getCompetitionTable()

	// 先删除该比赛的所有报名
	_, err := db.Exec("DELETE FROM "+regTable+" WHERE competition_id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 删除比赛
	_, err = db.Exec("DELETE FROM "+compTable+" WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Competition deleted successfully"})
}

// 更新比赛可见性
func updateCompetitionVisibility(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	compTable := getCompetitionTable()

	var update struct {
		Visible bool `json:"visible"`
	}

	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		"UPDATE "+compTable+" SET visible = ? WHERE id = ?",
		update.Visible, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Competition visibility updated successfully"})
}

// 上传图片
func uploadImage(w http.ResponseWriter, r *http.Request) {
	// 限制上传文件大小为 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "文件太大,最大支持10MB", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	fileType := handler.Header.Get("Content-Type")
	if !allowedTypes[fileType] {
		http.Error(w, "不支持的文件类型,仅支持 jpg, png, gif, webp", http.StatusBadRequest)
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "创建上传目录失败", http.StatusInternalServerError)
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(handler.Filename)
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), handler.Filename)))
	filename := hex.EncodeToString(hash.Sum(nil)) + ext

	// 保存文件
	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		http.Error(w, "保存文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	// 返回文件URL
	fileURL := fmt.Sprintf("/uploads/logos/%s", filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      fileURL,
		"filename": filename,
	})
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// 全局 CORS 中间件
	r.Use(corsMiddleware)

	// 静态文件服务 - 用于访问上传的图片
	fs := http.FileServer(http.Dir("./uploads"))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", fs))

	// OPTIONS 处理器（用于 CORS 预检）
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
	}).Methods("OPTIONS")

	// 比赛路由
	r.HandleFunc("/api/competitions", createCompetition).Methods("POST")
	r.HandleFunc("/api/competitions", getCompetitions).Methods("GET")
	r.HandleFunc("/api/competitions/{id}", getCompetition).Methods("GET")
	r.HandleFunc("/api/competitions/{id}", updateCompetition).Methods("PUT")
	r.HandleFunc("/api/competitions/{id}", deleteCompetition).Methods("DELETE")
	r.HandleFunc("/api/competitions/{id}/visibility", updateCompetitionVisibility).Methods("PUT")

	// 报名路由
	r.HandleFunc("/api/registrations", createRegistration).Methods("POST")
	r.HandleFunc("/api/competitions/{competitionId}/registrations", getRegistrations).Methods("GET")
	r.HandleFunc("/api/competitions/{competitionId}/user-registration", getUserRegistration).Methods("GET")
	r.HandleFunc("/api/registrations/{id}", updateRegistration).Methods("PUT")

	// 用户路由
	r.HandleFunc("/api/user/login", userLogin).Methods("POST")
	r.HandleFunc("/api/user/hoj-auto-login", hojAutoLogin).Methods("GET", "POST")

	// 图片上传路由
	r.HandleFunc("/api/upload/image", uploadImage).Methods("POST")

	fmt.Println("=================================================")
	fmt.Println("       报名系统后端服务器")
	fmt.Println("=================================================")
	if useMySQL {
		fmt.Println("数据库: MySQL (远程)")
		fmt.Println("地址: 43.143.133.62:3306")
	} else {
		fmt.Println("数据库: SQLite (本地)")
		fmt.Println("文件: ./registration.db")
	}
	fmt.Println("=================================================")
	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("=================================================")
	log.Fatal(http.ListenAndServe(":8080", r))
}
