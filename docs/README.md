# 文档目录 / Documentation

本目录包含 go-cache 项目的详细文档。

This directory contains detailed documentation for the go-cache project.

## 📚 文档列表 / Document List

### 核心文档 / Core Documentation

- **[SERIALIZER_GUIDE.md](SERIALIZER_GUIDE.md)** - 序列化器使用指南 / Serializer usage guide
  - Gob 和 JSON 序列化器的详细使用说明
  - 性能对比和使用建议
  - 完整的代码示例

- **[NIL_VALUES.md](NIL_VALUES.md)** - Nil 值支持文档 / Nil value support documentation
  - Nil 值的完整支持说明
  - 使用示例和最佳实践
  - 常见问题解答

### 迁移和改进 / Migration and Improvements

- **[GOB_MIGRATION.md](GOB_MIGRATION.md)** - Gob 迁移文档 / Gob migration guide
  - 从 msgpack 迁移到 gob 的完整记录
  - 技术实现细节
  - 性能对比和测试结果

- **[IMPROVEMENTS.md](IMPROVEMENTS.md)** - 项目改进记录 / Project improvements log
  - 修复的重大 bug
  - 性能优化记录
  - 功能增强说明

### 技术分析 / Technical Analysis

- **[SERIALIZATION_OPTIONS.md](SERIALIZATION_OPTIONS.md)** - 序列化方案分析 / Serialization options analysis
  - 不同序列化方案的对比
  - 技术选型建议
  - 实施路线图

## 📖 阅读顺序建议 / Recommended Reading Order

### 初次使用 / First Time Users

1. 从主 [README](../README.md) 开始了解项目概况
2. 阅读 [SERIALIZER_GUIDE.md](SERIALIZER_GUIDE.md) 了解序列化器使用
3. 如需处理 nil 值，参考 [NIL_VALUES.md](NIL_VALUES.md)

### 迁移用户 / Migration Users

1. 阅读 [GOB_MIGRATION.md](GOB_MIGRATION.md) 了解迁移背景
2. 参考 [IMPROVEMENTS.md](IMPROVEMENTS.md) 了解修复的问题
3. 查看 [SERIALIZATION_OPTIONS.md](SERIALIZATION_OPTIONS.md) 理解技术选型

### 贡献者 / Contributors

1. 查看 [IMPROVEMENTS.md](IMPROVEMENTS.md) 了解代码质量改进
2. 参考所有技术文档深入理解实现细节
3. 阅读测试文档了解测试规范

## 🔗 相关链接 / Related Links

- [主 README](../README.md) - 项目主文档
- [测试文档](../test/README.md) - 测试说明文档
- [源代码](../) - 项目源代码

---

**维护**: 本目录下的文档应与代码保持同步更新  
**Maintenance**: Documentation in this directory should be kept in sync with the code

