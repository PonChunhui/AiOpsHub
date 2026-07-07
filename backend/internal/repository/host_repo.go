package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type HostRepository struct {
	db *gorm.DB
}

func NewHostRepository() *HostRepository {
	return &HostRepository{db: database.DB}
}

func (r *HostRepository) GetGroupTree() ([]model.HostGroup, error) {
	var groups []model.HostGroup
	err := r.db.Where("parent_id IS NULL").Order("name ASC").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	for i := range groups {
		err := r.loadGroupChildren(&groups[i])
		if err != nil {
			return nil, err
		}
	}

	return groups, nil
}

func (r *HostRepository) loadGroupChildren(group *model.HostGroup) error {
	var children []model.HostGroup
	err := r.db.Where("parent_id = ?", group.ID).Order("name ASC").Find(&children).Error
	if err != nil {
		return err
	}

	for i := range children {
		err := r.loadGroupChildren(&children[i])
		if err != nil {
			return err
		}
	}

	group.Children = children
	return nil
}

func (r *HostRepository) GetGroupByID(id string) (*model.HostGroup, error) {
	var group model.HostGroup
	err := r.db.Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *HostRepository) CreateGroup(group *model.HostGroup) error {
	return r.db.Create(group).Error
}

func (r *HostRepository) UpdateGroup(group *model.HostGroup) error {
	return r.db.Save(group).Error
}

func (r *HostRepository) DeleteGroup(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.HostGroup{}).Error
}

func (r *HostRepository) HasChildrenOrHosts(groupID string) (bool, error) {
	var childCount int64
	err := r.db.Model(&model.HostGroup{}).Where("parent_id = ?", groupID).Count(&childCount).Error
	if err != nil {
		return false, err
	}

	if childCount > 0 {
		return true, nil
	}

	var hostCount int64
	err = r.db.Model(&model.Host{}).Where("group_id = ?", groupID).Count(&hostCount).Error
	if err != nil {
		return false, err
	}

	return hostCount > 0, nil
}

func (r *HostRepository) ListHosts(groupID string, page, pageSize int) ([]model.Host, int64, error) {
	var hosts []model.Host
	var total int64

	query := r.db.Model(&model.Host{})
	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&hosts).Error
	return hosts, total, err
}

func (r *HostRepository) GetHostByID(id string) (*model.Host, error) {
	var host model.Host
	err := r.db.Where("id = ?", id).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepository) CreateHost(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *HostRepository) UpdateHost(host *model.Host) error {
	return r.db.Save(host).Error
}

func (r *HostRepository) DeleteHost(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Host{}).Error
}

func (r *HostRepository) BatchCreateHosts(hosts []*model.Host) error {
	return r.db.Create(&hosts).Error
}

func (r *HostRepository) BatchDeleteHosts(ids []string) error {
	return r.db.Where("id IN ?", ids).Delete(&model.Host{}).Error
}

func (r *HostRepository) CreateSSHSessionLog(log *model.SSHSessionLog) error {
	return r.db.Create(log).Error
}

func (r *HostRepository) ListSSHSessionLogs(hostID string, limit int) ([]model.SSHSessionLog, error) {
	var logs []model.SSHSessionLog
	query := r.db.Model(&model.SSHSessionLog{})
	if hostID != "" {
		query = query.Where("host_id = ?", hostID)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *HostRepository) FindHostByIdentifier(identifier string) (*model.Host, error) {
	var host model.Host

	err := r.db.Where("ip = ?", identifier).First(&host).Error
	if err == nil {
		return &host, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	err = r.db.Where("name = ?", identifier).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &host, nil
}
