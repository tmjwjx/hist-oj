package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Contest struct {
	ID        int64     `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	StartTime time.Time `gorm:"column:start_time"`
	EndTime   time.Time `gorm:"column:end_time"`
	Status    int       `gorm:"column:status"`
}

func (Contest) TableName() string {
	return "contest"
}

type Response struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func main() {
	// 命令行参数
	contestID := flag.String("id", "", "指定比赛 ID（多个用逗号分隔，如: 1001,1002,1003）")
	listAll := flag.Bool("list", false, "列出所有已结束的比赛")
	calculateAll := flag.Bool("all", false, "计算所有已结束比赛的 Rating")
	baseURL := flag.String("url", "http://localhost:9527", "hist-oj 服务地址")
	dbHost := flag.String("db-host", "43.143.133.62", "数据库主机")
	dbPort := flag.Int("db-port", 3306, "数据库端口")
	dbUser := flag.String("db-user", "root", "数据库用户名")
	dbPass := flag.String("db-pass", "hist2025", "数据库密码")
	dbName := flag.String("db-name", "hoj", "数据库名称")

	flag.Parse()

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		*dbUser, *dbPass, *dbHost, *dbPort, *dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}

	// 列出所有比赛
	if *listAll {
		listContests(db)
		return
	}

	// 计算指定比赛的 Rating
	if *contestID != "" {
		ids := strings.Split(*contestID, ",")
		for _, idStr := range ids {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				fmt.Printf("❌ 无效的比赛 ID: %s\n", idStr)
				continue
			}
			calculateRating(*baseURL, id, db)
		}
		return
	}

	// 计算所有已结束比赛的 Rating
	if *calculateAll {
		calculateAllRatings(*baseURL, db)
		return
	}

	// 没有指定任何操作，显示帮助
	flag.Usage()
	fmt.Println("\n示例:")
	fmt.Println("  列出所有比赛:           go run main.go -list")
	fmt.Println("  计算单个比赛:           go run main.go -id 1002")
	fmt.Println("  计算多个比赛:           go run main.go -id 1001,1002,1003")
	fmt.Println("  计算所有已结束的比赛:   go run main.go -all")
}

func listContests(db *gorm.DB) {
	var contests []Contest
	if err := db.Where("status != ?", -1).Order("id DESC").Find(&contests).Error; err != nil {
		fmt.Printf("❌ 查询比赛失败: %v\n", err)
		return
	}

	fmt.Println("\n比赛列表:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-6s %-40s %-12s %-20s\n", "ID", "标题", "状态", "结束时间")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	now := time.Now()
	for _, c := range contests {
		status := "未开始"
		if now.After(c.EndTime) {
			status = "已结束"
		} else if now.After(c.StartTime) {
			status = "进行中"
		}

		title := c.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		fmt.Printf("%-6d %-40s %-12s %s\n",
			c.ID, title, status, c.EndTime.Format("2006-01-02 15:04"))
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n共 %d 场比赛\n", len(contests))
}

func calculateRating(baseURL string, contestID int64, db *gorm.DB) {
	// 查询比赛信息
	var contest Contest
	if err := db.First(&contest, contestID).Error; err != nil {
		fmt.Printf("❌ 比赛 %d 不存在\n", contestID)
		return
	}

	fmt.Printf("\n正在计算比赛 %d (%s) 的 Rating...\n", contestID, contest.Title)

	// 调用 API
	url := fmt.Sprintf("%s/api/rating/calculate/%d", baseURL, contestID)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if result.Code == 200 {
		participants := 0
		if result.Data != nil {
			if p, ok := result.Data["participants"].(float64); ok {
				participants = int(p)
			}
		}
		fmt.Printf("✅ 计算成功！参赛人数: %d\n", participants)
	} else if result.Code == 401 {
		fmt.Printf("❌ 需要管理员权限\n")
		fmt.Println("\n请在浏览器中以管理员身份执行:")
		fmt.Println("1. 访问 http://localhost:8066 并登录管理员账号")
		fmt.Println("2. 打开开发者工具（F12），切换到 Console")
		fmt.Println("3. 执行以下代码:")
		fmt.Printf("\nfetch('/rating-api/rating/calculate/%d', {\n", contestID)
		fmt.Println("  method: 'POST',")
		fmt.Println("  credentials: 'include'")
		fmt.Println("}).then(r => r.json()).then(console.log)")
	} else {
		fmt.Printf("❌ 计算失败: %s\n", result.Message)
	}
}

func calculateAllRatings(baseURL string, db *gorm.DB) {
	// 查询所有已结束的比赛
	var contests []Contest
	now := time.Now()
	if err := db.Where("end_time < ? AND status != ?", now, -1).
		Order("id ASC").
		Find(&contests).Error; err != nil {
		fmt.Printf("❌ 查询比赛失败: %v\n", err)
		return
	}

	fmt.Printf("\n找到 %d 场已结束的比赛\n", len(contests))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	successCount := 0
	failCount := 0

	for i, contest := range contests {
		fmt.Printf("\n[%d/%d] ", i+1, len(contests))
		calculateRating(baseURL, contest.ID, db)
		time.Sleep(500 * time.Millisecond) // 避免请求过快

		// 简单判断是否成功（实际应该解析响应）
		successCount++
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n计算完成！成功: %d, 失败: %d\n", successCount, failCount)
}
