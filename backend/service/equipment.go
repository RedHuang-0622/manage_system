package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"manage_system/dao"
	"manage_system/models"
	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"manage_system/pkg/safego"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BorrowStatusChecker 设备模块对借阅模块的查询契约
type BorrowStatusChecker interface {
	CountByEquipmentAndStatusInTx(tx *gorm.DB, equipID uint, status string) (int64, error)
	CountByEquipmentAndStatus(equipID uint, status string) (int64, error)
}

type EquipmentService struct {
	db          *gorm.DB
	dao         dao.EquipmentDAO
	borrowDAO   BorrowStatusChecker
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewEquipmentService(db *gorm.DB, equipDAO dao.EquipmentDAO, borrowDAO BorrowStatusChecker, rdb *redis.Client, logger *zap.Logger) *EquipmentService {
	return &EquipmentService{
		db:          db,
		dao:         equipDAO,
		borrowDAO:   borrowDAO,
		redisClient: rdb,
		logger:      logger,
	}
}

type CreateEquipReq struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Model       string `json:"model" binding:"max=64"`
	Category    string `json:"category" binding:"max=32"`
	TotalStock  uint   `json:"total_stock" binding:"gte=0"`
	Location    string `json:"location" binding:"max=64"`
	Description string `json:"description" binding:"max=1024"`
}

type UpdateEquipReq struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=128"`
	Model       string `json:"model" binding:"omitempty,max=64"`
	Category    string `json:"category" binding:"omitempty,max=32"`
	TotalStock  *uint  `json:"total_stock"`
	Location    string `json:"location" binding:"omitempty,max=64"`
	Description string `json:"description" binding:"omitempty,max=1024"`
	Status      *int8  `json:"status"`
}

type ListEquipReq struct {
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
	Keyword       string `form:"keyword"`
	Category      string `form:"category"`
	Status        *int   `form:"status"` // 指针类型：nil 表示不过滤
	OnlyAvailable int    `form:"only_available"`
}

type EquipmentDTO struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Model          string    `json:"model"`
	Category       string    `json:"category"`
	TotalStock     uint      `json:"total_stock"`
	AvailableStock uint      `json:"available_stock"`
	Location       string    `json:"location"`
	Description    string    `json:"description"`
	Status         int8      `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *EquipmentService) Create(ctx context.Context, req *CreateEquipReq) (*models.LabEquipment, error) {
	equip := &models.LabEquipment{
		Name:           req.Name,
		Model:          req.Model,
		Category:       req.Category,
		TotalStock:     req.TotalStock,
		AvailableStock: req.TotalStock, // 入库时可用=总量
		Location:       req.Location,
		Description:    req.Description,
		Status:         1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.dao.Insert(equip); err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	// 异步失效列表缓存（使用 safego 防止 panic 传播）
	safego.Go(s.logger, func() { s.invalidateListCache(context.Background()) })

	return equip, nil
}

func (s *EquipmentService) GetByID(ctx context.Context, id uint) (*EquipmentDTO, error) {
	// 查缓存
	cacheKey := fmt.Sprintf("equip:detail:%d", id)
	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var dto EquipmentDTO
			if json.Unmarshal([]byte(cached), &dto) == nil {
				return &dto, nil
			}
		}
	}

	equip, err := s.dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if equip == nil {
		return nil, fmt.Errorf("[%d] 设备不存在", errcode.ErrEquipmentNotFound)
	}

	dto := toEquipmentDTO(equip)

	// 写缓存
	if s.redisClient != nil {
		data, _ := json.Marshal(dto)
		s.redisClient.Set(ctx, cacheKey, data, 60*time.Second)
	}

	return dto, nil
}

func (s *EquipmentService) ListPage(ctx context.Context, req *ListEquipReq) (*PageResult, error) {
	p, ps, offset := response.NormalizePagination(req.Page, req.PageSize, 12, 100)
	req.Page, req.PageSize = p, ps

	// 查缓存
	cacheKey := buildListCacheKey(req)
	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var result PageResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	onlyAvailable := req.OnlyAvailable == 1
	equips, total, err := s.dao.FindPage(offset, req.PageSize, req.Keyword, req.Category, req.Status, onlyAvailable)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	dtos := make([]EquipmentDTO, len(equips))
	for i, e := range equips {
		dtos[i] = *toEquipmentDTO(&e)
	}

	result := &PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     dtos,
	}

	// 写缓存 (TTL 30s)
	if s.redisClient != nil {
		data, _ := json.Marshal(result)
		s.redisClient.Set(ctx, cacheKey, data, 30*time.Second)
	}

	return result, nil
}

func (s *EquipmentService) Update(ctx context.Context, id uint, req *UpdateEquipReq) error {
	equip, err := s.dao.FindByID(id)
	if err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if equip == nil {
		return fmt.Errorf("[%d] 设备不存在", errcode.ErrEquipmentNotFound)
	}

	if req.Name != "" {
		equip.Name = req.Name
	}
	if req.Model != "" {
		equip.Model = req.Model
	}
	if req.Category != "" {
		equip.Category = req.Category
	}
	if req.Location != "" {
		equip.Location = req.Location
	}
	if req.Description != "" {
		equip.Description = req.Description
	}
	if req.Status != nil {
		equip.Status = *req.Status
	}

	// 库存联动：delta = newTotal - oldTotal，可用库存同步增减
	if req.TotalStock != nil {
		delta := int(*req.TotalStock) - int(equip.TotalStock)
		newAvailable := int(equip.AvailableStock) + delta
		if newAvailable < 0 {
			return fmt.Errorf("[%d] 库存不足，当前可用库存为 %d，无法将总量下调至 %d",
				errcode.ErrStockInsufficient, equip.AvailableStock, *req.TotalStock)
		}
		equip.TotalStock = *req.TotalStock
		equip.AvailableStock = uint(newAvailable)
	}

	equip.UpdatedAt = time.Now()
	if err := s.dao.Update(equip); err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	// 事务成功后异步失效缓存（使用 safego 防止 panic 传播）
	safego.Go(s.logger, func() {
		ctx := context.Background()
		s.invalidateDetailCache(ctx, id)
		s.invalidateListCache(ctx)
	})

	return nil
}

func (s *EquipmentService) Disable(ctx context.Context, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// ① 锁定设备行
		equip, err := s.dao.GetForUpdate(tx, id)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if equip == nil {
			return fmt.Errorf("[%d] 设备不存在", errcode.ErrEquipmentNotFound)
		}
		if equip.Status == 0 {
			return nil // 幂等：已下架
		}

		// ② 在同一事务内检查未归还借阅（消除 TOCTOU 窗口）
		count, err := s.borrowDAO.CountByEquipmentAndStatusInTx(tx, id, "已借出")
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if count > 0 {
			return fmt.Errorf("[%d] 设备有未归还借阅，无法下架", errcode.ErrConflict)
		}

		// ③ 更新状态
		equip.Status = 0
		equip.UpdatedAt = time.Now()
		if err := s.dao.UpdateWithTx(tx, equip); err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}

		return nil
	})
	// 事务 COMMIT 成功后异步失效缓存
	// 注：缓存失效在事务成功返回后调用，消除 COMMIT 前 DEL 的竞态窗口
	// 实际实现中 Disable 返回 error 后由 controller 决定是否失效缓存，见 controller 层
}

// ──────────────────────────── 缓存工具 ────────────────────────────

func (s *EquipmentService) InvalidateEquipmentCache(ctx context.Context, equipID uint) {
	s.invalidateDetailCache(ctx, equipID)
	s.invalidateListCache(ctx)
}

func (s *EquipmentService) invalidateDetailCache(ctx context.Context, id uint) {
	if s.redisClient == nil {
		return
	}
	s.redisClient.Del(ctx, fmt.Sprintf("equip:detail:%d", id))
}

func (s *EquipmentService) invalidateListCache(ctx context.Context) {
	if s.redisClient == nil {
		return
	}
	iter := s.redisClient.Scan(ctx, 0, "equip:list:*", 100).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
}

func buildListCacheKey(req *ListEquipReq) string {
	status := 0
	if req.Status != nil {
		status = *req.Status
	}
	return fmt.Sprintf("equip:list:%d:%d:%s:%s:%d:%d",
		req.Page, req.PageSize, req.Keyword, req.Category, status, req.OnlyAvailable)
}

func toEquipmentDTO(e *models.LabEquipment) *EquipmentDTO {
	return &EquipmentDTO{
		ID:             e.ID,
		Name:           e.Name,
		Model:          e.Model,
		Category:       e.Category,
		TotalStock:     e.TotalStock,
		AvailableStock: e.AvailableStock,
		Location:       e.Location,
		Description:    e.Description,
		Status:         e.Status,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
