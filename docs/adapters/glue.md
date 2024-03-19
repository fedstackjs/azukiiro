---
outline: deep
---

# Glue适配器

顾名思义，这是一个胶水评测插件，用于复用现有的评测工具并接入AOI系统中。

## 配置文件

```yml
adapter: glue
config:
  run: |
    unzip -d problem $GLUE_PROBLEM_DATA
    unzip -d solution $GLUE_SOLUTION_DATA
    bash ./problem/judge.sh
  timeout: 600
```

## 说明

Glue插件的允许执行一些用户指定的命令，这些命令将对解答进行评测，并通过写入特殊的文件上报评测结果。

::: warning
Glue 评测插件会直接在评测机上执行 `run` 字段中的命令，因此请确保题目配置仅由可信任的人员进行。

同时，必须指定 `unsafe` 编译标签，该评测插件才会被编译并启用。
:::

命令执行的环境中包含一些特殊的环境变量：

- `GLUE_PROBLEM_DATA`: 题目数据（zip）的路径
- `GLUE_SOLUTION_DATA`: 解答数据（zip）的路径
- `GLUE_REPORT`: 写入这个文件以上报评测状态
- `GLUE_DETAILS`: 评测详情文件（json）的路径

其中，`GLUE_REPORT`的使用方式类似于Github Action中的`$GITHUB_OUTPUT`，写入这个文件的内容将被解析并上报。支持的字段如下：

```bash
# 设置分数
echo score=$total_score >"$GLUE_REPORT"
# 设置评测状态
echo status=Running >"$GLUE_REPORT"
# 设置评测消息
echo message=Running >"$GLUE_REPORT"
# 设置性能信息
echo metrics='{"cpu":1,"mem":1024}' >$GLUE_REPORT
# 开始上报
echo commit=1 >"$GLUE_REPORT"
```
