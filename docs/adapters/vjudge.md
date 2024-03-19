---
outline: deep
---

# VJudge适配器

支持同步[VJudge](https://vjudge.net)评测状态。

和其他的评测方式不同，VJudge适配器不需要用户在AOI平台提交代码。相反，用户需要在VJudge平台上提交代码，并在AOI平台上提交对应VJudge提交的链接。之后，评测插件将自动获取VJudge上的评测状态并上传。

## 配置文件

```json
{
  "adapter": "uoj",
  "config": {
    "oj": "LibreOJ",
    "probNum": "1"
  }
}
```

## 说明

- `oj`: VJudge上的OJ名称。
- `probNum`: VJudge上的题目编号。

例如：对于 `https://vjudge.net/problem/CodeForces-1328B`，应该如下配置：

```json
{
  "adapter": "vjudge",
  "config": {
    "oj": "CodeForces",
    "probNum": "1328B"
  }
}
```
