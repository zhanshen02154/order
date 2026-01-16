
<a name="v6.0.0"></a>
## [v6.0.0](https://github.com/zhanshen02154/order/compare/v5.0.0...v6.0.0) (2026-01-15)

### Bug Fixes

* 调整日志级别配置检测
* 调整配置信息检查逻辑
* 修复库存扣减事件
* 增加投递到死信队列链路追踪
* 投递到死信队列后标记为处理成功

### Code Refactoring

* 调整消费者最大处理时间
* 调整sarama组件配置参数
* **事件侦听器:** 删除client
* **事件侦听器:** 日志和链路追踪剥离事件侦听器

### Features

* 新增prometheus监控

### BREAKING CHANGE
- 删除client
- 事件侦听器的日志移到包装器并在入口函数初始化
- 事件侦听器的发布消息回调链路追踪移到包装器并在入口函数初始化
- 事件侦听器的死信队列移到包装器并在入口函数初始化
- 移除utils目录


<a name="v5.0.0"></a>
## [v5.0.0](https://github.com/zhanshen02154/order/compare/v4.0.0...v5.0.0) (2026-01-06)

### Bug Fixes

* 删除数据库查询的Debug语句
* callerSkip增加到2
* 删除SQL操作的Debug语句
* 修复事件侦听器
* 移除TraceId
* 修复type为core的日志缺失问题
* 调整callerSkip
* 修复发布日志的发布时间阈值判断
* 恢复丢失的代码
* 恢复丢失的代码
* 取消切片对象池
* 修复事件侦听器日志
* 修复GORM日志级别判断
* **GORM日志:** 增加logLevel赋值
* **GORM日志:** 找不到记录输出警告日志
* **日志:** 删除对象池
* **死信队列包装器:** 修复死信队列循环推入问题
* **链路追踪:** 调整包装器顺序

### Code Refactoring

* 取消GORM事务的独立会话
* 优化日志和字符串生成
* broker配置改为从config获取
* GRPC请求日志Logger的TraceID提取自Span
* **事件侦听器:** 从context获取TraceId
* **事件侦听器:** 事件侦听器改为异步发送
* **事件元数据:** 修改事件元数据时间戳为毫秒级时间戳
* **日志:** 调整发布事件时间阈值
* **日志:** 优化GRPC请求和发布事件及订阅事件日志

### Features

* 新增日志级别控制 fixed [#124](https://github.com/zhanshen02154/order/issues/124)
* 新增扣减库存死信队列操作
* 增加GRPC请求的链路追踪
* **链路追踪:** 新增订阅事件链路追踪
* **链路追踪:** 增加发布事件链路追踪
* **链路追踪:** 新增GORM数据库链路追踪

### Performance Improvements

* 日志字段和日志信息构造器用对象池

### Reverts

* Feat/jaeger

### BREAKING CHANGE


- 删除所有高频操作中的fmt.Sprinf
- 优化GRPC请求、发布/订阅日志的生成过程

- 修改事件元数据时间戳为毫秒级时间戳

- 移除同步生产者的logger，由事件侦听器里的logger代替。
- 新增异步生产者
- 移除同步生产者
- 新增异步生产者链路追踪
- 重构事件侦听器配置为Option，支持传入broker、client、logger和发布时间阈值

- 订阅日志新增订阅处理时间阈值，超过该值日志级别为警告
- 发布日志新增发布时间阈值，超过该值日志级别为警告
- GRPC请求新增请求时间阈值，超过该值日志级别为警告
- GRPC请求和订阅事件日志配置改为Option配置


<a name="v4.0.0"></a>
## [v4.0.0](https://github.com/zhanshen02154/order/compare/v3.0.0...v4.0.0) (2025-12-24)

### Bug Fixes

* 调整kafka拉取数据量
* 修复发布事件日志类型
* **GORM日志:** 关闭Info以下级别的日志
* **pprof:** 默认直接启动pprof
* **pprof:** 修复pprof协程控制问题
* **事件侦听器:** map全部清除后设置为nil
* **分布式锁:** map全部清除后设置为nil
* **日志组件:** 修复无法记录service和version的问题

### Code Refactoring

* 移除product的proto
* 删除订阅事件日志的error字段
* 删除error字段
* 调整日志记录器
* **ETCD分布式锁:** 采用共享Session
* **broker:** 扩大请求数量数量

### Features

* 添加订阅事件日志和请求日志（含GORM日志）
* **发布事件:** 新增日志记录
* **日志:** 优化日志类型
* **日志:** 新增日志类型

### Performance Improvements

* **事务管理:** 用独立会话启动事务


<a name="v3.0.0"></a>
## [v3.0.0](https://github.com/zhanshen02154/order/compare/v2.0.0...v3.0.0) (2025-12-09)

### Bug Fixes

* 删除示例事件
* 更改应用层事件侦听器为新名称
* 修复事件包装器元数据的时间戳转换
* 降低pprof采样频率
* **ETCD分布式锁:** 释放锁采用单独的超时上下文
* **ETCD分布式锁:** ETCD分布式锁由共享session改为每个锁独立维护session。
* **事件侦听器:** 消息的Key强制字符串类型

### Code Refactoring

* 接口层事件处理器结构体改为私有
* 事件侦听器结构体改为私有，仅开放接口
* Etcd分布式锁结构体改为私有
* 支付回调接口收到信息改为订单处理中
* 调整目录结构
* **pprof:** 导入pprof包
* **获取DB实例:** 没有事务实例则用WithContext

### Features

* 新增确认支付方法
* 新增响应代码判断是否移入死信队列
* 新增商品事件处理器
* 新增更新支付状态
* 事件侦听器支持传入消息的Key
* 应用死信队列包装器
* 新增死信队列
* 新增事件总线
* **broker:** 新增基于kafka的broker
* **接口层:** 新增订单事件处理器
* **配置结构体:** 新增Broker配置结构体

### Performance Improvements

* 优化依赖包
* **pprof:** 调整pprof采样频率

### BREAKING CHANGE


基础设施层移除的目录：
- broker
- config
- server
- registory

- 新增事件总线

- 新增订单事件处理器

- 新增基于kafka的broker
- 新增连接broker的事件侦听器
- 新增service客户端发布元数据处理包装器


<a name="v2.0.0"></a>
## [v2.0.0](https://github.com/zhanshen02154/order/compare/v1.0.1...v2.0.0) (2025-11-28)

### Bug Fixes

* 调整打印日志
* 恢复TTL到默认值
* **ETCD分布式锁:** 恢复 DialKeepAliveTime参数
* **ETCD分布式锁:** 加锁和解锁使用单独的context
* **ETCD分布式锁:** 先关闭会话再取消上下文
* **ETCD分布式锁:** 使用cancel上下文处理会话
* **ETCD分布式锁:** 初始化分布式锁时创建Mutex
* **ETCD分布式锁:** 删除AutoSyncInterval
* **ETCD分布式锁:** 创建会话失败无需关闭
* **打印错误日志:** 不需要格式化的使用log.Error
* **打印错误日志:** 改用log.Errorf和log.Error

### Code Refactoring

* 移除不使用的依赖
* **Consul配置源:** 获取Consul配置源逻辑调整到基础设施层
* **ETCD分布式锁:** 移动测试文件
* **ETCD分布式锁:** 删除ETCDLock内的客户端
* **pprof服务器:** 移动到基础设施层的server包
* **主函数和加锁逻辑:** 优化主函数和加锁逻辑
* **健康检查服务器:** 移动到基础设施层的server包
* **基础设施层:** 升级Go micro框架到4.11.0版本
* **服务上下文:** main函数初始化组件移入服务上下文
* **订单支付回调:** 调整订单支付回调功能

### Features

* **DTM分布式事务:** 新增分布式事务
* **ETCD分布式锁:** 新增tryLock
* **ETCD分布式锁:** 新增分布式锁的接口和ETCD实现类

### Performance Improvements

* **ETCD分布式锁:** 设置心跳和心跳超时
* **ETCD分布式锁:** 共享Session减少资源开销
* **商品服务proto文件:** 优化proto文件以支持go micro v4
* **商品服务客户端:** 移除商品服务客户端
* **订单服务Proto:** 修改订单服务的Proto文件

### BREAKING CHANGE


- 集成DTM分布式事务
- 新增事务管理器支持子事务屏障处理的方法

- 重写基础设施层里涉及go micro 2.9.1的方法


<a name="v1.0.1"></a>
## [v1.0.1](https://github.com/zhanshen02154/order/compare/v1.0.0...v1.0.1) (2025-11-23)

### Bug Fixes

* **打印错误日志:** 改用log.Errorf和log.Error


<a name="v1.0.0"></a>
## v1.0.0 (2025-11-17)

### Bug Fixes

* 设置订单服务客户端超时时间
* 已支付成功的订单不再操作
* 已支付成功的订单不再操作
* **client:** 修复商品服务客户端Consul.watch异常的问题
* **config:** 修复配置错误问题
* **main函数:** 优化服务beforeStop逻辑

### Code Refactoring

* **all:** 修改go mod名称
* **infrastructure:** 调整初始化数据库及健康检查探针

### Features

* **config:** 增加健康检查探针地址
* **config:** 增加健康检查探针地址
* **consul register:** 修复WaitTime超时设置失效问题

### Performance Improvements

* **ProductClient:** 更改商品服务客户端初始化逻辑
* **proto:** 细化返回结果

