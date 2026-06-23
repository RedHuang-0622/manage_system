package service

import (
	"context"
	"fmt"
	"time"

	"manage_system/dao"
	"manage_system/models"
	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"manage_system/pkg/safego"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BorrowService struct {
	db        *gorm.DB
	borrowDAO dao.BorrowDAO
	equipDAO  dao.EquipmentDAO
	cacheSvc  EquipmentCacheInvalidator
	logger    *zap.Logger
}

// EquipmentCacheInvalidator 借阅模块对设备缓存失效的契约
type EquipmentCacheInvalidator interface {
	InvalidateEquipmentCache(ctx context.Context, equipID uint)
}

func NewBorrowService(db *gorm.DB, borrowDAO dao.BorrowDAO, equipDAO dao.EquipmentDAO, cacheSvc EquipmentCacheInvalidator, logger *zap.Logger) *BorrowService {
	return &BorrowService{
		db:        db,
		borrowDAO: borrowDAO,
		equipDAO:  equipDAO,
		cacheSvc:  cacheSvc,
		logger:    logger,
	}
}

type ApplyBorrowReq struct {
	EquipmentID uint   `json:"equipment_id" binding:"required,gt=0"`
	Quantity    uint   `json:"quantity" binding:"required,gt=0"`
	ApplyNote   string `json:"apply_note" binding:"max=256"`
}

type ApproveBorrowReq struct {
	Approve     bool   `json:"approve"`
	ApproveNote string `json:"approve_note" binding:"max=256"`
}

type ListBorrowReq struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	UserID      uint   `form:"user_id"`
	EquipmentID uint   `form:"equipment_id"`
	Status      string `form:"status"`
}

type BorrowDTO struct {
	ID          uint                `json:"id"`
	UserID      uint                `json:"user_id"`
	EquipmentID uint                `json:"equipment_id"`
	Quantity    uint                `json:"quantity"`
	Status      models.BorrowStatus `json:"status"`
	ApplyNote   string              `json:"apply_note"`
	ApproveNote string              `json:"approve_note"`
	ApproverID  uint                `json:"approver_id"`
	ApplyAt     time.Time           `json:"apply_at"`
	ApproveAt   *time.Time          `json:"approve_at"`
	ReturnAt    *time.Time          `json:"return_at"`
	User        *UserBrief          `json:"user,omitempty"`
	Equipment   *EquipmentBrief     `json:"equipment,omitempty"`
	Approver    *UserBrief          `json:"approver,omitempty"`
}

type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
}

type EquipmentBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (s *BorrowService) Apply(ctx context.Context, userID uint, req *ApplyBorrowReq) (*models.BorrowRecord, error) {
	// 校验设备存在且上架
	equip, err := s.equipDAO.FindByID(req.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if equip == nil || equip.Status == 0 {
		return nil, fmt.Errorf("[%d] 设备不存在或已下架", errcode.ErrEquipmentNotFound)
	}

	// 初步库存检查（不做最终保障，最终由审批时的 FOR UPDATE 保证）
	if equip.AvailableStock < req.Quantity {
		return nil, fmt.Errorf("[%d] 设备库存不足", errcode.ErrStockInsufficient)
	}

	// 检查重复借阅
	count, err := s.borrowDAO.CountActiveByUserAndEquipment(userID, req.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if count > 0 {
		return nil, fmt.Errorf("[%d] 您已有该设备的未归还借阅", errcode.ErrDuplicateBorrow)
	}

	record := &models.BorrowRecord{
		UserID:      userID,
		EquipmentID: req.EquipmentID,
		Quantity:    req.Quantity,
		Status:      models.BorrowStatusApplied,
		ApplyNote:   req.ApplyNote,
		ApplyAt:     time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.borrowDAO.Insert(record); err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	return record, nil
}

func (s *BorrowService) Approve(ctx context.Context, borrowID uint, approve bool, approverID uint, note string) (*models.BorrowRecord, error) {
	var record *models.BorrowRecord

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// ① 锁定工单行
		var err error
		record, err = s.borrowDAO.GetForUpdate(tx, borrowID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if record == nil {
			return fmt.Errorf("[%d] 工单不存在", errcode.ErrNotFound)
		}
		if record.Status != models.BorrowStatusApplied {
			return fmt.Errorf("[%d] 工单状态异常，无法审批", errcode.ErrBorrowStatusInvalid)
		}

		// ② 锁定设备库存行（跨表行锁，防超卖）
		equip, err := s.equipDAO.GetForUpdate(tx, record.EquipmentID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if equip == nil {
			return fmt.Errorf("[%d] 设备不存在", errcode.ErrEquipmentNotFound)
		}

		now := time.Now()
		if approve {
			// 审批通过 → 扣减库存
			if equip.AvailableStock < record.Quantity {
				return fmt.Errorf("[%d] 库存不足，无法借出", errcode.ErrBorrowApproveFailed)
			}
			equip.AvailableStock -= record.Quantity
			record.Status = models.BorrowStatusBorrowed

			if err := s.equipDAO.UpdateWithTx(tx, equip); err != nil {
				return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
			}
		} else {
			// 审批拒绝
			record.Status = models.BorrowStatusRejected
		}

		// ③ 更新工单
		record.ApproverID = &approverID
		record.ApproveNote = note
		record.ApproveAt = &now
		record.UpdatedAt = now
		if err := s.borrowDAO.UpdateWithTx(tx, record); err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}

		return nil // COMMIT
	})

	// 事务 COMMIT 成功后再异步失效设备缓存（使用 safego 防止 panic 传播）
	if err == nil && s.cacheSvc != nil {
		safego.Go(s.logger, func() {
			s.cacheSvc.InvalidateEquipmentCache(context.Background(), record.EquipmentID)
		})
	}

	return record, err
}

func (s *BorrowService) Return(ctx context.Context, borrowID uint, operatorID uint, isAdmin bool) (*models.BorrowRecord, error) {
	var record *models.BorrowRecord

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		record, err = s.borrowDAO.GetForUpdate(tx, borrowID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if record == nil {
			return fmt.Errorf("[%d] 工单不存在", errcode.ErrNotFound)
		}
		if record.Status != models.BorrowStatusBorrowed {
			return fmt.Errorf("[%d] 工单状态异常，无法归还", errcode.ErrBorrowStatusInvalid)
		}

		// 本人校验：非管理员只能归还自己的借阅
		if !isAdmin && record.UserID != operatorID {
			return fmt.Errorf("[%d] 权限不足", errcode.ErrPermissionDenied)
		}

		// 锁定设备
		equip, err := s.equipDAO.GetForUpdate(tx, record.EquipmentID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if equip == nil {
			return fmt.Errorf("[%d] 设备不存在", errcode.ErrEquipmentNotFound)
		}

		// 恢复库存
		equip.AvailableStock += record.Quantity
		if err := s.equipDAO.UpdateWithTx(tx, equip); err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}

		// 更新状态
		now := time.Now()
		record.Status = models.BorrowStatusReturned
		record.ReturnAt = &now
		record.UpdatedAt = now
		if err := s.borrowDAO.UpdateWithTx(tx, record); err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}

		return nil
	})

	if err == nil && s.cacheSvc != nil {
		safego.Go(s.logger, func() {
			s.cacheSvc.InvalidateEquipmentCache(context.Background(), record.EquipmentID)
		})
	}

	return record, err
}

func (s *BorrowService) Cancel(ctx context.Context, borrowID uint, operatorID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		record, err := s.borrowDAO.GetForUpdate(tx, borrowID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if record == nil {
			return fmt.Errorf("[%d] 工单不存在", errcode.ErrNotFound)
		}
		if record.Status != models.BorrowStatusApplied {
			return fmt.Errorf("[%d] 工单状态异常，无法取消", errcode.ErrBorrowStatusInvalid)
		}

		// 本人校验
		if record.UserID != operatorID {
			return fmt.Errorf("[%d] 权限不足", errcode.ErrPermissionDenied)
		}

		record.Status = models.BorrowStatusCanceled
		record.UpdatedAt = time.Now()
		if err := s.borrowDAO.UpdateWithTx(tx, record); err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		return nil
	})
}

func (s *BorrowService) ListMyRecords(ctx context.Context, userID uint, req *ListBorrowReq) (*PageResult, error) {
	p, ps, offset := response.NormalizePagination(req.Page, req.PageSize, 10, 100)
	req.Page, req.PageSize = p, ps

	records, total, err := s.borrowDAO.FindPage(offset, req.PageSize, userID, 0, req.Status)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	dtos := toBorrowDTOs(records)
	return &PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     dtos,
	}, nil
}

func (s *BorrowService) ListPending(ctx context.Context, req *ListBorrowReq) (*PageResult, error) {
	p, ps, offset := response.NormalizePagination(req.Page, req.PageSize, 10, 100)
	req.Page, req.PageSize = p, ps

	records, total, err := s.borrowDAO.FindPage(offset, req.PageSize, 0, 0, "申请中")
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	dtos := toBorrowDTOs(records)
	return &PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     dtos,
	}, nil
}

func (s *BorrowService) ListAll(ctx context.Context, req *ListBorrowReq) (*PageResult, error) {
	p, ps, offset := response.NormalizePagination(req.Page, req.PageSize, 10, 100)
	req.Page, req.PageSize = p, ps

	records, total, err := s.borrowDAO.FindPage(offset, req.PageSize, req.UserID, req.EquipmentID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	dtos := toBorrowDTOs(records)
	return &PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     dtos,
	}, nil
}

func toBorrowDTOs(records []models.BorrowRecord) []BorrowDTO {
	dtos := make([]BorrowDTO, len(records))
	for i, r := range records {
		dtos[i] = toBorrowDTO(&r)
	}
	return dtos
}

func toBorrowDTO(r *models.BorrowRecord) BorrowDTO {
	var approverID uint
	if r.ApproverID != nil {
		approverID = *r.ApproverID
	}
	dto := BorrowDTO{
		ID:          r.ID,
		UserID:      r.UserID,
		EquipmentID: r.EquipmentID,
		Quantity:    r.Quantity,
		Status:      r.Status,
		ApplyNote:   r.ApplyNote,
		ApproveNote: r.ApproveNote,
		ApproverID:  approverID,
		ApplyAt:     r.ApplyAt,
		ApproveAt:   r.ApproveAt,
		ReturnAt:    r.ReturnAt,
	}
	if r.User.ID != 0 {
		dto.User = &UserBrief{ID: r.User.ID, Username: r.User.Username, RealName: r.User.RealName}
	}
	if r.Equipment.ID != 0 {
		dto.Equipment = &EquipmentBrief{ID: r.Equipment.ID, Name: r.Equipment.Name}
	}
	if r.Approver.ID != 0 {
		dto.Approver = &UserBrief{ID: r.Approver.ID, Username: r.Approver.Username, RealName: r.Approver.RealName}
	}
	return dto
}
