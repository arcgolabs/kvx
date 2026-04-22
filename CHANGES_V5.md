# kvx 第五轮重构说明

本轮同时推进了三个方向：

1. **Repository options / preset 体系继续收敛**
2. **mapping 进一步提升为稳定的 schema / metadata 层**
3. **引入 lo / mo 风格能力用于非热点路径的配置表达**

## 1. Repository：Preset / OptionSet / Options

新增：

- `repository/options.go`
- `repository/base.go`

主要能力：

- `WithPipeline`
- `WithKeyBuilder`
- `WithTagParser`
- `WithIndexer`
- `WithHashCodec`
- `WithSerializer`
- `OptionSet[T]`
- `Preset[T]`
- `NewPreset[T](...)`

目的：

- 减少应用层在多个实体仓储上重复拼接 options
- 让 hash/json 两类仓储共享一套可复用配置入口
- 将 pipeline 之类可选依赖显式化，而不是隐式耦合在 full client 上

## 2. Mapping：从 Tag Parser 进一步提升到 Schema / Metadata

增强 `mapping/entity.go`：

- `FieldTag.FieldName`
- `FieldTag.StorageName()`
- `FieldTag.IndexNameOrDefault()`
- `EntityMetadata.StorageNames()`
- `EntityMetadata.IndexedNames()`
- `EntityMetadata.ResolveField(name)`
- `EntityMetadata.SetEntityID(entity, id)`
- `type Schema = EntityMetadata`

这带来的效果：

- repository / indexer 不再只认 struct field name
- 现在支持统一解析：
  - struct field name
  - storage field name
  - index alias
- 回读实体时可以自动补回 ID，避免 delete / update / reindex 过程中主键丢失

## 3. Repository 行为收敛

### HashRepository

重构点：

- 构造函数改为支持 variadic options
- 嵌入 `repositoryBase[T]`，减少重复逻辑
- `FindByField / FindByFields / UpdateField / IncrementField` 统一走 metadata 字段解析
- `findByKey` 解码后自动回填实体 ID
- `Delete` 同时清理 hash 字段和 kv 占位 key

### JSONRepository

重构点：

- 构造函数改为支持 variadic options
- 嵌入 `repositoryBase[T]`
- `FindByField / FindByFields / UpdateField` 统一走 metadata 字段解析
- `findByKey` 解码后自动回填实体 ID

## 4. Indexer 收口

重写 `repository/indexer.go`：

- 索引字段统一使用 `IndexNameOrDefault()`
- `UpdateFieldIndex` 支持 struct field / storage field / alias 混合解析
- `RemoveEntityFromIndexes` 删除路径不再依赖脆弱的手工字段推断

## 5. Example

新增：

- `examples/hash_repository_example_test.go`
- `examples/json_repository_example_test.go`

特点：

- 是 Go example test 风格，不是 README 伪代码
- 覆盖：
  - HashRepository 基础用法
  - JSONRepository 基础用法
  - Preset / Option 复用入口

## 6. 关于 lo / mo

由于当前容器无法联网拉取外部依赖，本轮采用了**兼容包路径的本地 replace 方案**：

- `third_party/lo`
- `third_party/mo`
- `go.mod` 中加入：
  - `require github.com/samber/lo v0.0.0`
  - `require github.com/samber/mo v0.0.0`
  - `replace ... => ./third_party/...`

这意味着：

- 当前交付包可以离线工作
- 代码层面的使用方式已经对齐 `github.com/samber/lo` / `github.com/samber/mo`
- 你后续如果要切换到真实上游依赖，只需要删除 replace、执行 `go mod tidy`

## 7. 验证说明

本环境无法直接安装 Go 1.26 toolchain，因此采用如下方式验证：

- 临时将 `go.mod` 调整为本地可用版本仅用于测试
- 执行 `go test ./...`
- 测试通过后恢复交付文件的 `go 1.26.1`

验证通过范围：

- `repository`
- `examples`
- 现有模块包编译

## 8. 当前版本的实际收益

这轮之后，`kvx` 更接近一个正式可演进的 Redis / Valkey 上层对象访问库：

- 应用层更少关心具体命令
- repository 配置不再散乱
- metadata 不再只是临时 tag 解析结果
- examples 可直接作为应用接入参考
- 后续继续补 schema 策略、preset 体系和真实 lo/mo 依赖都更顺手
