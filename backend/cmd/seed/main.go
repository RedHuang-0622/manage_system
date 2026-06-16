package main

import (
	"fmt"
	"log"
	"time"

	"manage_system/models"
	"manage_system/pkg/config"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("conf/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	dsn := cfg.MySQL.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// AutoMigrate (确保表存在，首次也能跑)
	db.AutoMigrate(
		&models.SysUser{},
		&models.SysRole{},
		&models.LabEquipment{},
		&models.BorrowRecord{},
	)

	// Seed roles
	roles := []models.SysRole{
		{RoleName: "super_admin", Description: "超级管理员（指导老师）", IsSystem: 1},
		{RoleName: "lab_admin", Description: "实验室负责人", IsSystem: 1},
		{RoleName: "member", Description: "普通成员", IsSystem: 1},
	}
	for _, r := range roles {
		db.Where("role_name = ?", r.RoleName).FirstOrCreate(&r)
	}
	fmt.Println("角色: OK")

	// Seed demo users (password: 123456)
	type userSeed struct {
		Username string
		RealName string
		Email    string
		Phone    string
		RoleName string
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), 12)
	users := []userSeed{
		{"zhangwei", "张伟", "zhangwei@lab.edu.cn", "13800138001", "lab_admin"},
		{"liulei", "刘磊", "liulei@lab.edu.cn", "13800138006", "equipment_manager"},
		{"lina", "李娜", "lina@lab.edu.cn", "13800138002", "member"},
		{"wangqiang", "王强", "wangqiang@lab.edu.cn", "13800138003", "member"},
		{"zhaomin", "赵敏", "zhaomin@lab.edu.cn", "13800138004", "member"},
		{"chenjing", "陈静", "chenjing@lab.edu.cn", "13800138005", "member"},
		{"sunyue", "孙悦", "sunyue@lab.edu.cn", "13800138007", "viewer"},
	}

	for _, u := range users {
		var role models.SysRole
		db.Where("role_name = ?", u.RoleName).First(&role)

		user := models.SysUser{
			Username:     u.Username,
			PasswordHash: string(hash),
			RealName:     u.RealName,
			Email:        u.Email,
			Phone:        u.Phone,
			RoleID:       role.ID,
			Status:       1,
		}
		result := db.Where("username = ?", u.Username).FirstOrCreate(&user)
		if result.RowsAffected > 0 {
			fmt.Printf("用户: NEW %s (%s)\n", u.Username, u.RealName)
		} else {
			fmt.Printf("用户: SKIP %s (exists)\n", u.Username)
		}
	}

	// Seed demo equipment
	type equipSeed struct {
		Name        string
		Model       string
		Category    string
		TotalStock  uint
		Location    string
	}
	equips := []equipSeed{
		{"示波器 TDS2024C", "TDS2024C", "测量仪器", 3, "A301实验室"},
		{"GPU服务器 DGX-A100", "DGX-A100", "服务器", 4, "A302机房"},
		{"3D打印机 Ultimaker S5", "Ultimaker S5", "制造设备", 2, "B101创客空间"},
		{"频谱分析仪 N9020B", "Keysight N9020B", "测量仪器", 1, "A303实验室"},
		{"MacBook Pro M3", "MacBook Pro 16 M3 Pro", "笔记本电脑", 10, "B202设备室"},
		{"逻辑分析仪 16862A", "Keysight 16862A", "测量仪器", 2, "A304实验室"},
	}

	for _, e := range equips {
		equip := models.LabEquipment{
			Name:           e.Name,
			Model:          e.Model,
			Category:       e.Category,
			TotalStock:     e.TotalStock,
			AvailableStock: e.TotalStock,
			Location:       e.Location,
			Status:         1,
		}
		result := db.Where("name = ?", e.Name).FirstOrCreate(&equip)
		if result.RowsAffected > 0 {
			fmt.Printf("设备: NEW %s (库存:%d)\n", e.Name, e.TotalStock)
		} else {
			fmt.Printf("设备: SKIP %s (exists)\n", e.Name)
		}
	}

	// Seed borrow records (simulate a realistic workflow)
	fmt.Println()
	type borrowSeed struct {
		Username     string
		EquipmentName string
		Quantity     uint
		ApplyNote    string
		ApproveNote  string
		Approve      bool
		Return       bool
	}
	borrows := []borrowSeed{
		{"lina", "GPU服务器 DGX-A100", 1, "深度学习训练需要GPU资源", "批准使用，请于周五前归还", true, true},
		{"wangqiang", "示波器 TDS2024C", 1, "数字电路实验测量", "同意借用", true, false},
		{"zhaomin", "MacBook Pro M3", 1, "远程办公和文档处理", "批准", true, false},
		{"chenjing", "3D打印机 Ultimaker S5", 1, "毕业设计模型制作", "", false, false},
	}

	for _, b := range borrows {
		var user models.SysUser
		var equip models.LabEquipment
		if err := db.Where("username = ?", b.Username).First(&user).Error; err != nil {
			fmt.Printf("借阅: SKIP (用户%s不存在)\n", b.Username)
			continue
		}
		if err := db.Where("name = ?", b.EquipmentName).First(&equip).Error; err != nil {
			fmt.Printf("借阅: SKIP (设备%s不存在)\n", b.EquipmentName)
			continue
		}

		// Create borrow record
		now := time.Now()
		record := models.BorrowRecord{
			UserID:      user.ID,
			EquipmentID: equip.ID,
			Quantity:    b.Quantity,
			Status:      "申请中",
			ApplyNote:   b.ApplyNote,
			ApplyAt:     now,
		}

		// Check for duplicate
		var existing models.BorrowRecord
		if err := db.Where("user_id = ? AND equipment_id = ? AND status = ?",
			user.ID, equip.ID, "申请中").First(&existing).Error; err == nil {
			fmt.Printf("借阅: SKIP (%s已有%s的申请中记录)\n", b.Username, b.EquipmentName)
			continue
		}

		db.Create(&record)

		// If approved, process approval
		if b.Approve {
			approverID := uint(1)
			approveAt := time.Now()
			record.Status = "已借出"
			record.ApproveNote = b.ApproveNote
			record.ApproverID = &approverID
			record.ApproveAt = &approveAt
			db.Save(&record)
			// Deduct stock
			db.Model(&equip).Update("available_stock", gorm.Expr("available_stock - ?", b.Quantity))
		}

		// If returned, process return
		if b.Return {
			returnAt := time.Now()
			record.Status = "已归还"
			record.ReturnAt = &returnAt
			db.Save(&record)
			// Restore stock
			db.Model(&equip).Update("available_stock", gorm.Expr("available_stock + ?", b.Quantity))
		}

		status := record.Status
		if !b.Approve {
			status = "申请中(待审批)"
		}
		fmt.Printf("借阅: %s 借 %s x%d -> %s\n", b.Username, b.EquipmentName, b.Quantity, status)
	}

	fmt.Println("\n种子数据植入完成!")
}
