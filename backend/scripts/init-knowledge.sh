#!/bin/bash

API_URL="http://localhost:8080/api/v1"
TOKEN=""

get_token() {
    response=$(curl -s -X POST "${API_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    TOKEN=$(echo $response | jq -r '.data.token')
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        echo "Failed to get token"
        exit 1
    fi
    echo "Token obtained: $TOKEN"
}

add_document() {
    id=$1
    title=$2
    content=$3
    category=$4
    tags=$5
    
    echo "Adding document: $title"
    
    curl -s -X POST "${API_URL}/rag/documents" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "$(jq -n \
            --arg id "$id" \
            --arg title "$title" \
            --arg content "$content" \
            --arg category "$category" \
            --argjson tags "$tags" \
            '{id: $id, title: $title, content: $content, category: $category, tags: $tags}')" > /dev/null
    
    sleep 0.5
}

main() {
    get_token
    
    add_document "sop-001" "CPU使用率异常高排查SOP" "# CPU使用率异常高排查SOP

## 问题现象
- CPU使用率持续超过80%
- 系统响应缓慢
- 进程卡顿或超时

## 排查步骤

### 1. 初步定位
查看整体CPU使用情况: top -bn1 | head -20
查看各进程CPU占用: ps aux --sort=-%cpu | head -10

### 2. 确认异常进程
查看进程详细信息: top -H -p PID
查看进程线程: ps -Lp PID

### 3. 分析原因
- 死循环: 查看代码逻辑，检查循环条件
- 频繁GC: 监控Java GC情况
- 算法问题: 分析代码复杂度
- 锁竞争: 查看线程等待状态

### 4. 解决方案
- 优化算法，减少CPU计算
- 调整GC参数（Java应用）
- 减少锁粒度或使用无锁结构
- 限流或降级处理" "troubleshooting" '["CPU","性能","排查","SOP"]'
    
    add_document "sop-002" "内存泄漏排查SOP" "# 内存泄漏排查SOP

## 问题现象
- 内存持续增长不释放
- 出现OOM错误
- 服务频繁重启

## 排查步骤

### 1. 监控内存趋势
实时监控内存: top -p PID
查看内存使用历史: vmstat 1 10

### 2. 分析内存分布

Java应用:
查看堆内存统计: jmap -histo PID
生成堆转储文件: jmap -dump:format=b,file=heap.hprof PID

Go应用:
pprof分析: go tool pprof http://localhost:6060/debug/pprof/heap

### 3. 定位泄漏源
- 查看大对象分布
- 分析对象引用链
- 检查未关闭的资源（连接、文件、流）

### 4. 解决方案
- 及时关闭资源
- 使用对象池
- 设置合理的缓存过期策略" "troubleshooting" '["内存","泄漏","排查","SOP"]'
    
    add_document "sop-003" "数据库慢查询排查SOP" "# 数据库慢查询排查SOP

## 问题现象
- SQL执行时间超过阈值
- 数据库CPU/IO高
- 应用响应超时

## 排查步骤

### 1. 查看慢查询日志
MySQL查看慢查询配置: SHOW VARIABLES LIKE 'slow_query%'
查看最近慢查询: SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10

### 2. 分析执行计划
查看SQL执行计划: EXPLAIN SELECT * FROM users WHERE name = 'test'
查看详细执行计划: EXPLAIN ANALYZE SELECT

### 3. 定位问题
- 全表扫描: type=ALL，需添加索引
- 索引失效: 使用函数、类型转换、模糊查询
- JOIN效率低: 关联表过多或关联字段无索引

### 4. 解决方案
添加索引: CREATE INDEX idx_name ON users(name)
复合索引: CREATE INDEX idx_name_status ON users(name, status)" "troubleshooting" '["数据库","慢查询","排查","SOP"]'
    
    add_document "sop-004" "网络连接超时排查SOP" "# 网络连接超时排查SOP

## 问题现象
- 请求超时或连接失败
- TCP连接建立慢
- 网络抖动频繁

## 排查步骤

### 1. 检查网络连通性
ping测试: ping -c 10 target_ip
查看延迟和丢包率: mtr target_ip
端口连通性测试: nc -zv target_ip port

### 2. 分析连接状态
查看TCP连接状态: netstat -an | grep ESTABLISHED
查看连接详情: ss -s

### 3. 检查防火墙和路由
查看防火墙规则: iptables -L -n
查看路由表: ip route show

### 4. 解决方案
- 增加超时时间配置
- 启用TCP Keep-Alive
- 优化连接池配置" "troubleshooting" '["网络","超时","排查","SOP"]'
    
    add_document "sop-005" "磁盘IO瓶颈排查SOP" "# 磁盘IO瓶颈排查SOP

## 问题现象
- 磁盘读写延迟高
- IOPS达到上限
- 应用卡顿

## 排查步骤

### 1. 监控磁盘IO
实时IO监控: iostat -x 1 10
查看进程IO: iotop -o

### 2. 分析磁盘使用
查看磁盘空间: df -h
查看inode使用: df -i
查看目录大小: du -sh /var/* | sort -h

### 3. 定位高IO进程
查看进程读写统计: lsof /var/log
查看文件打开数: lsof | wc -l

### 4. 解决方案
- 使用SSD替代HDD
- 分散IO负载（多磁盘）
- 启用异步IO" "troubleshooting" '["磁盘","IO","排查","SOP"]'
    
    add_document "sop-006" "Kubernetes Pod启动失败排查SOP" "# Kubernetes Pod启动失败排查SOP

## 问题现象
- Pod状态为Pending/ImagePullBackOff/CrashLoopBackOff
- 服务无法正常运行

## 排查步骤

### 1. 查看Pod状态
查看Pod详情: kubectl describe pod pod-name -n namespace
查看Pod事件: kubectl get events

### 2. 分析失败原因
ImagePullBackOff: 检查镜像是否存在, docker pull image-name
CrashLoopBackOff: 查看Pod日志, kubectl logs pod-name --previous
Pending: 检查资源是否充足, kubectl describe nodes

### 3. 解决方案
- 检查镜像名称和tag
- 验证镜像仓库认证
- 增加资源配置
- 修复应用启动错误" "troubleshooting" '["K8s","Pod","排查","SOP"]'
    
    add_document "sop-007" "服务响应慢排查SOP" "# 服务响应慢排查SOP

## 问题现象
- HTTP请求响应时间超过阈值
- 用户投诉慢
- 监控告警

## 排查步骤

### 1. 定位慢接口
查看接口响应统计: curl localhost:8080/actuator/metrics
分析日志中的慢请求: grep took app.log

### 2. 分析调用链路
使用分布式追踪: Jaeger/SkyWalking查看trace
查看依赖服务响应时间

### 3. 检查下游依赖
- 数据库查询时间
- 外部API调用时间
- 缓存命中率

### 4. 解决方案
- 添加缓存层
- 异步处理耗时操作
- 优化SQL查询
- 增加并发处理能力" "troubleshooting" '["响应慢","性能","排查","SOP"]'
    
    add_document "sop-008" "JVM频繁Full GC排查SOP" "# JVM频繁Full GC排查SOP

## 问题现象
- Full GC频率高（大于1次每分钟）
- GC暂停时间长
- 应用性能下降

## 排查步骤

### 1. 监控GC情况
查看GC统计: jstat -gc PID 1000 10
查看GC原因: jstat -gcutil PID 1000 10

### 2. 分析GC日志
启用GC日志JDK8: -XX:+PrintGCDetails -Xloggc:gc.log
使用GCViewer或GCEasy分析

### 3. 分析内存分配
查看堆内存分布: jmap -histo PID
查看对象年龄分布: jstat -gcnew PID

### 4. 解决方案
调整堆大小: -Xms4g -Xmx4g
使用G1收集器: -XX:+UseG1GC" "troubleshooting" '["JVM","GC","排查","SOP"]'
    
    add_document "sop-009" "Redis连接异常排查SOP" "# Redis连接异常排查SOP

## 问题现象
- Redis连接超时
- 连接数达到上限
- 响应慢或拒绝服务

## 排查步骤

### 1. 检查Redis状态
查看Redis信息: redis-cli info
查看连接数: redis-cli info clients
查看慢日志: redis-cli slowlog get 10

### 2. 分析连接来源
查看客户端连接: redis-cli client list
统计连接来源: redis-cli client list | awk统计

### 3. 检查网络延迟
测试Redis响应时间: redis-cli --latency
测试延迟历史: redis-cli --latency-history

### 4. 解决方案
增加最大连接数: CONFIG SET maxclients 10000
设置连接超时: CONFIG SET timeout 300" "troubleshooting" '["Redis","连接","排查","SOP"]'
    
    add_document "sop-010" "MySQL死锁排查SOP" "# MySQL死锁排查SOP

## 问题现象
- 事务执行失败
- 出现Deadlock found错误
- 业务流程中断

## 排查步骤

### 1. 查看死锁日志
查看最近死锁: SHOW ENGINE INNODB STATUS
查看锁等待: SELECT * FROM INNODB_LOCK_WAITS

### 2. 分析死锁原因
查看当前事务: SELECT * FROM INNODB_TRX
查看锁信息: SELECT * FROM INNODB_LOCKS

### 3. 定位问题事务
- 检查事务隔离级别
- 分析SQL执行顺序
- 查看索引使用情况

### 4. 解决方案
使用悲观锁: SELECT * FROM table WHERE id=1 FOR UPDATE
按固定顺序访问表" "troubleshooting" '["MySQL","死锁","排查","SOP"]'
    
    add_document "sop-011" "应用OOM崩溃排查SOP" "# 应用OOM崩溃排查SOP

## 问题现象
- 应用进程被Kill
- 日志出现OOM错误
- 服务自动重启

## 排查步骤

### 1. 查看系统日志
查看OOM Killer日志: dmesg | grep oom
查看进程被Kill记录: journalctl -k | grep oom

### 2. 分析内存限制
查看进程内存限制: cat /proc/PID/limits
查看容器内存限制K8s: kubectl describe pod

### 3. 检查内存使用
查看进程内存: ps aux --sort=-%mem
查看容器内存使用: kubectl top pods

### 4. 解决方案
- 增加内存限制配置
- 优化内存使用
- 启用JVM堆外内存限制" "troubleshooting" '["OOM","崩溃","排查","SOP"]'
    
    add_document "sop-012" "DNS解析失败排查SOP" "# DNS解析失败排查SOP

## 问题现象
- 域名无法解析
- 连接超时
- 偶发性解析失败

## 排查步骤

### 1. 测试DNS解析
使用dig测试: dig example.com
使用nslookup测试: nslookup example.com

### 2. 检查DNS配置
查看DNS配置: cat /etc/resolv.conf
测试DNS服务器: dig @dns-server-ip example.com

### 3. 分析DNS延迟
查看DNS缓存: systemd-resolve --statistics

### 4. 解决方案
- 配置多个DNS服务器
- 启用DNS缓存
- 使用IP直连替代域名" "troubleshooting" '["DNS","解析","排查","SOP"]'
    
    add_document "sop-013" "Kubernetes Service无法访问排查SOP" "# Kubernetes Service无法访问排查SOP

## 问题现象
- Service IP无法访问
- 域名解析失败
- ClusterIP/NodePort不通

## 排查步骤

### 1. 检查Service状态
查看Service详情: kubectl describe svc service-name
查看Endpoints: kubectl get endpoints service-name

### 2. 检查Pod状态
查看后端Pod: kubectl get pods -l app=app-name

### 3. 测试Service连通性
在集群内测试: kubectl run test wget service-ip:port
测试DNS解析: nslookup service-name

### 4. 解决方案
- 检查Service selector与Pod label匹配
- 确认Pod正常运行且Ready
- 检查kube-proxy运行状态" "troubleshooting" '["K8s","Service","排查","SOP"]'
    
    add_document "sop-014" "日志文件过大排查SOP" "# 日志文件过大排查SOP

## 问题现象
- 日志文件占用大量磁盘空间
- 磁盘空间告警
- 影响应用性能

## 排查步骤

### 1. 查看日志文件大小
查看日志目录大小: du -sh /var/log
查看大文件: find /var/log -type f -size +100M

### 2. 分析日志来源
统计日志文件数量: ls -lh /var/log
查看日志写入速率: lsof /var/log/app.log

### 3. 检查日志配置
查看logrotate配置: cat /etc/logrotate.d/app
查看应用日志配置

### 4. 解决方案
配置日志轮转: daily rotate 7 compress size 100M
限制日志级别: ERROR替代DEBUG" "troubleshooting" '["日志","磁盘","排查","SOP"]'
    
    add_document "sop-015" "Kubernetes节点异常排查SOP" "# Kubernetes节点异常排查SOP

## 问题现象
- 节点状态NotReady
- Pod无法调度到节点
- 节点资源耗尽

## 排查步骤

### 1. 查看节点状态
查看节点信息: kubectl describe node node-name
查看节点条件: kubectl get node

### 2. 检查节点资源
查看资源使用: kubectl top node
查看节点容量: kubectl describe node

### 3. 检查kubelet状态
查看kubelet日志: journalctl -u kubelet
查看kubelet状态: systemctl status kubelet

### 4. 解决方案
- 重启kubelet服务
- 清理节点资源
- 检查网络连接" "troubleshooting" '["K8s","节点","排查","SOP"]'
    
    add_document "sop-016" "API接口返回500错误排查SOP" "# API接口返回500错误排查SOP

## 问题现象
- HTTP 500 Internal Server Error
- 接口无法正常响应
- 用户无法使用功能

## 排查步骤

### 1. 查看错误日志
查看应用日志: tail -100 /var/log/app/error.log
搜索500错误: grep 500 access.log

### 2. 分析错误堆栈
查看Java堆栈: jstack PID
查看Go panic日志: grep panic app.log

### 3. 检查依赖服务
测试数据库连接: mysql -h db-host
测试Redis连接: redis-cli ping

### 4. 解决方案
- 捕获并记录完整异常堆栈
- 检查空指针类型转换问题
- 验证输入参数合法性" "troubleshooting" '["API","500","排查","SOP"]'
    
    add_document "sop-017" "消息队列堆积排查SOP" "# 消息队列堆积排查SOP

## 问题现象
- 消息队列积压严重
- 消息处理延迟
- 消费者无法跟上生产速率

## 排查步骤

### 1. 查看队列状态
RabbitMQ查看队列: rabbitmqctl list_queues
Kafka查看堆积: kafka-consumer-groups --describe

### 2. 分析堆积原因
- 消费者处理速度慢
- 消费者数量不足
- 消费者故障
- 消息体过大

### 3. 检查消费者状态
查看消费者进程: ps aux | grep consumer
查看消费者日志: tail -f consumer.log

### 4. 解决方案
增加消费者实例数: kubectl scale deployment --replicas=10
批量处理消息" "troubleshooting" '["MQ","堆积","排查","SOP"]'
    
    add_document "sop-018" "配置文件加载失败排查SOP" "# 配置文件加载失败排查SOP

## 问题现象
- 应用启动失败
- 配置未生效
- 找不到配置项

## 排查步骤

### 1. 查看配置文件路径
检查配置文件是否存在: ls -la /etc/app/config.yaml
检查应用配置路径: ps aux | grep config

### 2. 验证配置文件格式
YAML格式验证: python yaml.safe_load
JSON格式验证: python json.tool

### 3. 检查配置权限
查看文件权限: ls -la config.yaml
检查文件所有权: stat config.yaml

### 4. 解决方案
- 检查配置文件路径配置
- 使用绝对路径替代相对路径
- 验证配置文件格式正确" "troubleshooting" '["配置","加载","排查","SOP"]'
    
    add_document "sop-019" "证书过期问题排查SOP" "# 证书过期问题排查SOP

## 问题现象
- HTTPS连接失败
- SSL/TLS握手错误
- 服务间调用失败

## 排查步骤

### 1. 查看证书有效期
查看证书信息: openssl x509 -in cert.pem -text
查看证书过期时间: openssl x509 -enddate

### 2. 验证证书链
验证证书链: openssl verify -CAfile ca.pem cert.pem
检查证书信任链: openssl s_client -showcerts

### 3. 检查证书配置
查看应用证书配置: grep cert /etc/app
查看Nginx证书配置: nginx -T | grep ssl

### 4. 解决方案
生成新证书: openssl req -x509 -days 365
更新K8s TLS Secret: kubectl create secret tls" "troubleshooting" '["证书","SSL","排查","SOP"]'
    
    add_document "sop-020" "容器镜像拉取失败排查SOP" "# 容器镜像拉取失败排查SOP

## 问题现象
- ImagePullBackOff错误
- ErrImagePull错误
- Pod无法启动

## 排查步骤

### 1. 查看错误详情
查看Pod事件: kubectl describe pod
查看具体错误: kubectl get pod -o jsonpath

### 2. 检查镜像是否存在
手动拉取镜像: docker pull image-name
查看镜像仓库: curl registry

### 3. 检查认证配置
查看Docker认证: cat .docker/config.json
查看K8s Secret认证: kubectl get secret docker-registry

### 4. 解决方案
创建Docker Registry Secret: kubectl create secret docker-registry
配置imagePullSecrets
使用国内镜像仓库加速" "troubleshooting" '["容器","镜像","排查","SOP"]'
    
    echo "Completed: 20 SOP documents added"
}

main