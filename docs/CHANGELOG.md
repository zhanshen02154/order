
<a name="v4.0.0"></a>
## [v4.0.0](https://github.com/zhanshen02154/order/compare/v3.0.0...v4.0.0) (2025-12-23)

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
* **broker:** 扩大请求数量

### Features

* 添加订阅事件日志和请求日志（含GORM日志）
* **发布事件:** 新增日志记录
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

