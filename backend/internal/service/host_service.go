package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
)

type HostService struct {
	BaseService
	repo *repository.HostRepository
}

func NewHostService() *HostService {
	return &HostService{
		repo: repository.NewHostRepository(),
	}
}

func (s *HostService) GetGroupTree() ([]model.HostGroup, error) {
	return s.repo.GetGroupTree()
}

func (s *HostService) GetGroupByID(id string) (*model.HostGroup, error) {
	return s.repo.GetGroupByID(id)
}

func (s *HostService) CreateGroup(name, parentID, description, createdBy string) (*model.HostGroup, error) {
	group := model.NewHostGroup()
	group.Name = name
	group.Description = description
	group.CreatedBy = createdBy
	group.UpdatedBy = createdBy

	if parentID != "" {
		parent, err := s.repo.GetGroupByID(parentID)
		if err != nil {
			return nil, s.HandleError(err, "父分组不存在")
		}
		group.ParentID = &parentID
		group.Level = parent.Level + 1
		group.Path = parent.Path + "/" + name
	} else {
		group.ParentID = nil
		group.Level = 0
		group.Path = "/" + name
	}

	if err := s.repo.CreateGroup(group); err != nil {
		return nil, s.HandleError(err, "创建分组失败")
	}

	s.LogInfo("主机分组创建成功: %s (%s)", name, group.ID)
	return group, nil
}

func (s *HostService) UpdateGroup(id, name, description, updatedBy string) (*model.HostGroup, error) {
	group, err := s.repo.GetGroupByID(id)
	if err != nil {
		return nil, s.HandleError(err, "分组不存在")
	}

	if name != "" {
		group.Name = name
	}
	if description != "" {
		group.Description = description
	}
	group.UpdatedBy = updatedBy
	group.UpdatedAt = time.Now()

	if err := s.repo.UpdateGroup(group); err != nil {
		return nil, s.HandleError(err, "更新分组失败")
	}

	s.LogInfo("主机分组更新成功: %s (%s)", group.Name, group.ID)
	return group, nil
}

func (s *HostService) DeleteGroup(id string) error {
	hasChildrenOrHosts, err := s.repo.HasChildrenOrHosts(id)
	if err != nil {
		return s.HandleError(err, "检查分组级联失败")
	}

	if hasChildrenOrHosts {
		return fmt.Errorf("分组下存在子分组或主机，无法删除")
	}

	if err := s.repo.DeleteGroup(id); err != nil {
		return s.HandleError(err, "删除分组失败")
	}

	s.LogInfo("主机分组删除成功: %s", id)
	return nil
}

func (s *HostService) HasChildrenOrHosts(groupID string) (bool, error) {
	return s.repo.HasChildrenOrHosts(groupID)
}

func (s *HostService) ListHosts(groupID string, page, pageSize int) ([]model.Host, int64, error) {
	return s.repo.ListHosts(groupID, page, pageSize)
}

func (s *HostService) GetHostByID(id string) (*model.Host, error) {
	return s.repo.GetHostByID(id)
}

func (s *HostService) CreateHost(groupID, name, hostType, ip string, port int, username, authType, password, privateKey, publicKey, remark, createdBy string) (*model.Host, error) {
	host := model.NewHost()
	host.GroupID = groupID
	host.Name = name
	host.HostType = hostType
	host.IP = ip
	host.Port = port
	host.Username = username
	host.AuthType = authType
	host.Password = password
	host.PrivateKey = privateKey
	host.PublicKey = publicKey
	host.Remark = remark
	host.CreatedBy = createdBy
	host.UpdatedBy = createdBy

	if err := s.repo.CreateHost(host); err != nil {
		return nil, s.HandleError(err, "创建主机失败")
	}

	s.LogInfo("主机创建成功: %s (%s) - %s@%s:%d", name, host.ID, username, ip, port)
	return host, nil
}

func (s *HostService) UpdateHost(id string, updates map[string]interface{}) (*model.Host, error) {
	host, err := s.repo.GetHostByID(id)
	if err != nil {
		return nil, s.HandleError(err, "主机不存在")
	}

	if groupID, ok := updates["group_id"].(string); ok && groupID != "" {
		host.GroupID = groupID
	}
	if name, ok := updates["name"].(string); ok && name != "" {
		host.Name = name
	}
	if hostType, ok := updates["host_type"].(string); ok && hostType != "" {
		host.HostType = hostType
	}
	if ip, ok := updates["ip"].(string); ok && ip != "" {
		host.IP = ip
	}
	if port, ok := updates["port"].(int); ok {
		host.Port = port
	}
	if username, ok := updates["username"].(string); ok && username != "" {
		host.Username = username
	}
	if authType, ok := updates["auth_type"].(string); ok && authType != "" {
		host.AuthType = authType
	}
	if password, ok := updates["password"].(string); ok {
		host.Password = password
	}
	if privateKey, ok := updates["private_key"].(string); ok {
		host.PrivateKey = privateKey
	}
	if publicKey, ok := updates["public_key"].(string); ok {
		host.PublicKey = publicKey
	}
	if remark, ok := updates["remark"].(string); ok {
		host.Remark = remark
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		host.Status = status
	}
	if updatedBy, ok := updates["updated_by"].(string); ok && updatedBy != "" {
		host.UpdatedBy = updatedBy
	}

	host.UpdatedAt = time.Now()

	if err := s.repo.UpdateHost(host); err != nil {
		return nil, s.HandleError(err, "更新主机失败")
	}

	s.LogInfo("主机更新成功: %s (%s)", host.Name, host.ID)
	return host, nil
}

func (s *HostService) DeleteHost(id string) error {
	if err := s.repo.DeleteHost(id); err != nil {
		return s.HandleError(err, "删除主机失败")
	}

	s.LogInfo("主机删除成功: %s", id)
	return nil
}

func (s *HostService) BatchImportHosts(groupID string, file io.Reader, createdBy string) ([]*model.Host, []string, error) {
	reader := csv.NewReader(file)
	var hosts []*model.Host
	var errors []string

	lineNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("读取CSV文件失败: %v", err))
			break
		}

		lineNum++
		if lineNum == 1 {
			continue
		}

		if len(record) < 7 {
			errors = append(errors, fmt.Sprintf("第%d行数据不完整", lineNum))
			continue
		}

		port := 22
		if len(record) > 3 && record[3] != "" {
			p, err := strconv.Atoi(record[3])
			if err != nil {
				errors = append(errors, fmt.Sprintf("第%d行端口格式错误: %s", lineNum, record[3]))
				continue
			}
			port = p
		}

		host := model.NewHost()
		host.GroupID = groupID
		host.Name = strings.TrimSpace(record[0])
		host.HostType = strings.TrimSpace(record[1])
		host.IP = strings.TrimSpace(record[2])
		host.Port = port
		host.Username = strings.TrimSpace(record[4])
		host.AuthType = strings.TrimSpace(record[5])

		if host.AuthType == "password" {
			host.Password = strings.TrimSpace(record[6])
		} else if host.AuthType == "key" {
			host.PrivateKey = strings.TrimSpace(record[6])
		}

		if len(record) > 7 {
			host.Remark = strings.TrimSpace(record[7])
		}

		host.CreatedBy = createdBy
		host.UpdatedBy = createdBy

		hosts = append(hosts, host)
	}

	if len(hosts) > 0 {
		if err := s.repo.BatchCreateHosts(hosts); err != nil {
			return nil, nil, s.HandleError(err, "批量创建主机失败")
		}
		s.LogInfo("批量导入主机成功: %d 个", len(hosts))
	}

	return hosts, errors, nil
}

func (s *HostService) BatchDeleteHosts(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	if err := s.repo.BatchDeleteHosts(ids); err != nil {
		return s.HandleError(err, "批量删除主机失败")
	}

	s.LogInfo("批量删除主机成功: %d 个", len(ids))
	return nil
}

func (s *HostService) TestConnection(hostID string) error {
	host, err := s.repo.GetHostByID(hostID)
	if err != nil {
		return s.HandleError(err, "主机不存在")
	}

	s.LogInfo("测试主机连接: %s (%s@%s:%d)", host.Name, host.Username, host.IP, host.Port)

	return nil
}

func (s *HostService) CreateSSHSessionLog(hostID, userID, action, sessionID, ipAddress string) (*model.SSHSessionLog, error) {
	log := model.NewSSHSessionLog()
	log.HostID = hostID
	log.UserID = userID
	log.Action = action
	log.SessionID = sessionID
	log.IPAddress = ipAddress
	log.StartTime = time.Now()

	if err := s.repo.CreateSSHSessionLog(log); err != nil {
		return nil, s.HandleError(err, "创建SSH会话日志失败")
	}

	return log, nil
}

func (s *HostService) UpdateSSHSessionLog(sessionID string, endTime time.Time, result string) error {
	return nil
}

func (s *HostService) ListSSHSessionLogs(hostID string, limit int) ([]model.SSHSessionLog, error) {
	return s.repo.ListSSHSessionLogs(hostID, limit)
}
