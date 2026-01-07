# 构建 AI Agent 应用

[概览](#概览) • [博客文章](#博客文章) • [特性](#特性) • [安装](#安装) • [使用](#使用) • [目录结构](#目录结构) • [示例](#示例) • [测试](#测试)

---

## 概览

本仓库提供了一个**统一的、多框架的平台**，用于设计、实现和评估 AI 驱动的 Agent（智能体）。通过将**场景定义**与**框架特定代码**分离，我们实现了：

* 每个场景的**单一规范**（位于 `src/scenarios/` 下）。
* 在 LangGraph、LangChain、Autogen（以及更多）中的**并行实现**。
* 一个**共享的评估工具**，用于比较不同框架的输出。
* **内置的可观测性**（Loki 日志记录 & OpenTelemetry/Tempo）。
* 核心工具和遥测设置的**单元测试**。

无论您是在构建电子商务支持机器人、IT 支持台助手、语音代理，还是介于两者之间的任何东西，此代码库都能帮助您从原型扩展到生产覆盖——同时保持一致性和可重用性。

---

## 博客文章

* [AI 智能体从入门到进阶再到落地完整教程](https://alanhou.org/ai-agents-complete-course/)
* [第一章 智能体导论](https://alanhou.org/chapter-1-introduction-to-agents/)
* [第二章 设计智能体系统](https://alanhou.org/chapter-2-designing-agent-systems/)
* [第三章 智能体系统的用户体验设计](https://alanhou.org/chapter-3-user-experience-design-for-agentic-systems/)
* [第四章 工具使用](https://alanhou.org/chapter-4-tool-use/)
* [第五章 智能体编排](https://alanhou.org/chapter-5-orchestration/)
* [第六章 知识与记忆](https://alanhou.org/chapter-6-knowledge-and-memory/)
* [第七章 智能体系统中的学习](https://alanhou.org/chapter-7-learning-in-agentic-systems/)


---

## 特性

* **与框架无关的场景规范**
  `src/scenarios/<scenario_name>/` 下的每个场景包含：

  * `spec.md`：用户旅程和成功标准的通俗英语描述。
  * `data/`：用于快速测试或演示的示例输入/输出 JSON。
  * `evaluation/`：一个共享的 `run_eval.py` 工具加上一个“黄金”评估集（JSON 或 CSV）。

* **多框架实现**
  在以下目录下并行实现每个场景：

  * `src/frameworks/langgraph/`
  * `src/frameworks/autogen/`
    *（通过遵循相同的文件夹模式轻松添加更多框架。）*

* **内置可观测性**

  * **Loki 日志记录**：
    `src/common/observability/loki_logger.py` 将结构化日志发布到本地 Loki 端点。
  * **OpenTelemetry / Tempo**：
    `src/common/observability/instrument_tempo.py` 设置 OTLP 导出器并将跨度（父级和子级）检测到 Tempo。

* **核心工具和遥测的单元测试**

  * 评估工具的测试：
    `tests/evaluation/test_ai_judge.py` & `test_memory_evaluation.py`
  * 可观测性代码的测试（monkeypatching 导出器）：
    `tests/observability/test_loki_logger.py` & `test_instrument_tempo.py`

---

## 安装

1. **克隆仓库**

   ```bash
   git clone https://github.com/your-org/agents.git
   cd agents
   ```

2. **创建 Conda（或 Virtualenv）环境**

   ```bash
   # 使用 Conda
   conda env create -f environment.yml
   conda activate agents
   ```

3. **安装 Python 依赖项（以及可编辑的 “src” 包）**

   ```bash
   pip install -r requirements.txt
   pip install -e src
   ```

   * `pip install -e src` 确保 `src/` 下的模块（例如 `common.*`，`frameworks.*`）是可导入的。

---

## 使用

### 1. 运行场景评估

每个场景都包含一个共享的评估脚本：

```bash
# 从仓库根目录：
cd src/common/scenarios/ecommerce_customer_support/evaluation

python -m src.common.evaluation.batch_evaluation \
  --dataset src/common/evaluation/scenarios/ecommerce_customer_support_evaluation_set.json \
  --graph_py src/frameworks/langgraph_agents/ecommerce_customer_support/customer_support_agent.py
```

### 2. 启动单个框架 Agent

如果您想手动运行电子商务 Agent 的 LangGraph 版本：

```bash
python - << 'PYCODE'
from frameworks.langgraph.scenarios.ecommerce_customer_support.implementation import run_ecommerce_support

payload = {
  "order": {"order_id": "A12345", "status": "Delivered", "total": 19.99},
  "messages": [{"type": "human", "content": "My mug arrived broken. Refund?"}]
}

response = run_ecommerce_support(payload)
print(response)
PYCODE
```

根据其他场景或框架相应地替换 `run_ecommerce_support` 和 payload 形状。

### 3. 可观测性

* **Loki 日志记录**
  代码中对 `log_to_loki(label, message)` 的任何调用都会将 JSON payload 发送到：

  ```
  http://localhost:3100/loki/api/v1/push
  ```

  将 Grafana/Loki 指向该端点以实时查看日志。

* **OpenTelemetry / Tempo**

  ```python
  from common.observability.instrument_tempo import do_work
  do_work()  # 向 OTLP 端点 (localhost:3200) 发出一个父跨度和三个子跨度
  ```

  要检测您自己的函数，请导入 `tracer = common.observability.instrument_tempo.tracer` 并将代码包装在 `with tracer.start_as_current_span("span-name"):` 块中。

---

## 目录结构

以下是所有内容的组织方式概览：

```
agents/
├── README.md
├── scenarios/                 
│   ├── ecommerce_customer_support.jsonl
│   └── ...
├── python/                    
│   ├── .gitignore
│   ├── environment.yml
│   ├── requirements.txt
│   ├── conftest.py
│   ├── src/
│   │   ├── common/
│   │   └── frameworks/
│   └── tests/
└── go/                        
    ├── go.mod
    ├── cmd/
    └── internal/
```

---

## 示例

### 1. 运行 LangChain Agent（电子商务支持）

```bash
# 从仓库根目录：
cd src/frameworks/langchain/scenarios/ecommerce_customer_support

# 示例用法：
python - << 'PYCODE'
from frameworks.langchain.scenarios.ecommerce_customer_support.implementation import run_ecommerce_support

payload = {
  "order": {"order_id": "A12345", "status": "Delivered", "total": 19.99},
  "messages": [{"type": "human", "content": "My mug arrived broken. Refund?"}]
}

response = run_ecommerce_support(payload)
print(response)
PYCODE
```

### 2. 运行 LangGraph Agent（电子商务支持）

```bash
# 从仓库根目录：
cd src/frameworks/langgraph/scenarios/ecommerce_customer_support

# 示例用法：
python - << 'PYCODE'
from frameworks.langgraph.scenarios/ecommerce_customer_support.implementation import run_ecommerce_support

payload = {
  "order": {"order_id": "A12345", "status": "Delivered", "total": 19.99},
  "messages": [{"type": "human", "content": "My mug arrived broken. Refund?"}]
}

response = run_ecommerce_support(payload)
print(response)
PYCODE
```

---

## 测试

我们使用 **pytest** 进行所有单元测试：

* **评估工具测试**：

  * `tests/evaluation/test_ai_judge.py`
  * `tests/evaluation/test_memory_evaluation.py`

* **可观测性测试**：

  * `tests/observability/test_loki_logger.py`
  * `tests/observability/test_instrument_tempo.py`

要运行完整的测试套件：

```bash
cd /Users/your-user/dev/agents
pytest -q
```

所有测试都应该通过且没有错误。
