QuotePro 便携包说明
==================

目录内容
- Windows: quotepro.exe
- Linux:   quotepro（可执行文件）
- config.example.yaml  配置示例，复制为 config.yaml 后按说明填写

使用步骤
1. 将 config.example.yaml 复制为 config.yaml，修改 jwtSecret 与 AI 相关配置。
2. 在解压目录下运行可执行文件（与 config.yaml 同目录）。
3. 浏览器访问 http://127.0.0.1:8080（端口可在 config.yaml 中改 port）。

环境变量可覆盖部分配置，参见后端 config 包。设置 SKIP_OPEN_BROWSER=1 可禁止启动时自动打开浏览器。

数据文件 quotepro.db 与上传目录 uploads/ 会生成在运行时的当前工作目录，建议在命令行中 cd 到程序所在目录后启动。
