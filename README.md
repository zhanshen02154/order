## 微服务架构演进实践之——订单服务

## 声明
- 本项目内置部分自编组件，不得未经允许下载Releases的产物及源码用于商业用途，若需合作请发送邮件到zhanshen02154@gmail.com联系作者本人。
- 严禁将该项目的任何代码及产物用于非法商业用途如赌博、诈骗、洗钱等，一经发现将追究法律责任！

## 项目描述
以“订单支付成功回调扣减商品库存”链路进行微服务架构演进实践，订单服务为领域驱动架构，包含领域层、应用层、基础设施层、接口层。

### 各层职责
- 接口层: 接收来自kafka事件的消息和GRPC请求。
- 应用层: 编排业务流程。
- 领域层: 处理业务逻辑。
- 基础设施层: broker、pprof、事件侦听器、logger等组件的初始化及仓储层具体实现。

## 项目文档
- [变更日志](./docs/CHANGELOG.md)
- [决策记录](./docs/DECISIONS.md)

## 目录结构
```treeofiles
├─cmd                           // 入口
├─docs                          // 文档
├─internal
│  ├─application                // 应用层
│  │  └─service                 // 服务（编排）
│  ├─bootstrap                  // 启动
│  ├─config                     // 配置
│  ├─domain                     // 领域层
│  │  ├─event                   // 事件
│  │  ├─model                   // 模型层
│  │  ├─repository              // 仓储层（接口）
│  │  └─service                 // 服务层
│  ├─infrastructure             // 基础设施层
│  │  ├─event                   // 事件相关组件
│  │  │  ├─monitor              // 监控
│  │  │  └─wrapper              // 包装器
│  │  └─persistence             // 持久化
│  └─interfaces                 // 接口层
│      ├─handler                // GRPC请求处理器
│      └─subscriber             // subscriber
├─pkg                           // 组件包
├─proto                         // protobuf
└─tests                         // 单元测试

```

## 技术选型

- 开发语言：Golang 1.20.10
- 框架: Go micro 4.11.0
- 数据库: MySQL 5.7.26
- 服务注册/发现: Consul 1.7.3
- 分布式锁: Redis 6.20.2
- 消息队列: kafka 3.0.1
- 链路追踪: Opentelemetry
- 监控: Prometheus Agent

## 服务器配置
| 配置                       | 数量 | 操作系统       | Docker版本 | Kubernetes版本 |
|--------------------------|----|------------|----------|--------------|
| 阿里云ECS  CPU x 4 + 8GB 内存 | 2  | CentOS 7.9 | 20.10.7  | 1.23.1       |

## 服务配置
| 数量 | Requests（单个Pod）   | Limits（单个Pod）      | 
|----|-------------------|--------------------| 
| 2  | 500m CPU + 128M内存 | 1100m CPU + 512M内存 |

## 本地开发环境搭建

1. 安装Golang 1.20.10、Apisix 3.4.1。
2. 安装protoc-gen-go。
```bash
 go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.23.0
```
3. 安装Go-micro对应版本的protoc-gen-micro。
```bash
  go install go-micro.dev/v4/cmd/micro@latest
```
4. 在根目录下生成Protobuf对应的go文件及go-micro文件
```bash
   # 指定两个proto_path，一个是项目的proto另一个是导入外部库的proto
   protoc --proto_path=./proto --proto_path=<include path> --go_out=. --micro_out=. ./proto/order/order.proto
```
## 注意事项
- proto文件更新后必须在Apisix的protos接口更新内容。
- 安装依赖必须指定版本并考虑与当前Golang版本的兼容性，防止在安装过程中升级golang或变更原有依赖。
- 由于配置文件放在服务注册中心Consul的KV获取，编译Docker镜像必须指定3个环境变量：CONSUL_HOST（consul的IP地址）、CONSUL_PORT（Consul端口）、CONSUL_PREFIX（前缀），没有指定则一律按本地开发环境处理。
- 上传到Apisix之前使用如下命令生成PB文件再使用base64编码作为content参数的内容上传。
```bash
  protoc --include_imports --descriptor_set_out=./order.pb --proto_path=./proto --proto_path=<include path> --go_out=. --micro_out=. ./proto/order/order.proto
  echo $(base64 -w0 order.pb) > order.txt  # 上传到Apisix用的是这个文件里的内容
```