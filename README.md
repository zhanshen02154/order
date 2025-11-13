## 微服务架构演进实践之——订单服务

## 项目描述
以“订单支付成功回调扣减商品库存”链路进行微服务架构演进实践，订单服务为领域驱动架构，包含领域层、应用层、基础设施层、接口层。

## 项目特点
- 采用GRPC跨服务通信实现订单支付成功回调调用商品服务扣减库存
- 使用自编的事务管理器结合GORM的事务完成事务处理实现数据一致性
- 原版为配置文件不易维护故改为Consul的K/V存储
- 增加K8S专用的健康检查探针
- 原版使用的GORM 1.9.6升级为1.30.0
- 依托GitHub结合Jenkins流水线实现CI/CD
- 废除原版的common、config目录
- 废除原版Dockerfile改用自编文件以适应实际业务需要

## 项目文档
- [变更日志](./docs/CHANGELOG.md)
- [决策记录](./docs/DECISIONS.md)

## 目录结构
```treeofiles
├─.chglog git-chglog    配置文件及模板
├─cmd                   入口文件
├─docs                  文档
├─internal
│  ├─application        应用层
│  │  ├─dto             DTO
│  │  └─service         应用层服务
│  ├─config             配置
│  ├─domain             领域层
│  │  ├─model           模型层
│  │  ├─repository      仓储层
│  │  └─service         领域层服务
│  ├─infrastructure     基础设施层
│  │  ├─cache           缓存
│  │  ├─client          客户端
│  │  ├─config          配置
│  │  ├─persistence     持久化
│  │  │  ├─gorm         GORM
│  │  │  └─transaction  事务
│  │  └─registry        服务注册
│  └─interfaces         接口层
│      └─handler        
├─pkg                   组件包
├─proto                 Protobuf
│  ├─order              订单服务
│  └─product            商品服务
└─utils                 辅助类（即将废弃）
```

## 工作流程
1. 客户端请求API接口
2. Apisix接收请求，通过Consul发现服务
3. 调用Order服务
4. Order服务的PayNotify调用Product服务的DeductInvetory
5. 返回结果

## 技术选型
| 开发语言           | 开发框架           | 数据库          | 服务注册/发现      |
|----------------|----------------|--------------|--------------|
| Golang 1.20.10 | Go-micro 2.9.1 | MySQL 5.7.26 | Consul 1.7.3 |

## 服务器配置
| 厂商  | 配置               | 数量 | 操作系统       | Docker版本 | Kubernetes版本 |
|-----|------------------|----|------------|----------|--------------|
| 阿里云 | CPU x 4 + 8GB 内存 | 2  | CentOS 7.9 | 20.10.7  | 1.23.1       |

## 本地开发环境搭建

1. 安装Golang 1.20.10、Apisix 3.4.1。
2. 安装protoc-gen-go。
```bash
 go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.23.0
```
3. 安装Go-micro对应版本的protoc-gen-micro。
```bash
  go install github.com/micro/micro/v2/cmd/protoc-gen-micro@v2.9.1
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
  echo $(base64 -w0) > order.txt  # 上传到Apisix用的是这个文件里的内容
```