
<a name="v2.0.0"></a>
## v2.0.0 (2025-11-27)

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
- 升级go micro框架到4.11.0
- 重写基础设施层里涉及go micro 2.9.1的方法


<a name="v1.0.1"></a>
## v1.0.1 (2025-11-23)

### Bug Fixes

* **打印错误日志:** 改用log.Errorf和log.Error


<a name="v1.0.0"></a>
## v1.0.0 (2025-11-17)

### Bug Fixes

* 设置订单服务客户端超时时间
* 已支付成功的订单不再操作
* **client:** 修复商品服务客户端Consul.watch异常的问题
* **config:** 修复配置错误问题
* **main函数:** 优化服务beforeStop逻辑

### Code Refactoring

* **all:** 修改go mod名称
* **infrastructure:** 调整初始化数据库及健康检查探针

### Features

* **config:** 增加健康检查探针地址
* **consul register:** 修复WaitTime超时设置失效问题

### Performance Improvements

* **ProductClient:** 更改商品服务客户端初始化逻辑
* **proto:** 细化返回结果

