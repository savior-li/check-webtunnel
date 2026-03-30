package i18n

var zh = map[string]string{
	"app_name":        "Tor Bridge Collector",
	"app_description": "Tor 桥接节点采集工具",

	"init_success":   "初始化成功",
	"init_failed":    "初始化失败",
	"config_created": "配置文件已创建",
	"db_created":     "数据库已创建",

	"fetch_success": "采集成功",
	"fetch_failed":  "采集失败",
	"fetching":      "正在采集桥梁数据...",
	"bridges_found": "发现桥梁数: %d",

	"validate_success":  "验证完成",
	"validate_failed":   "验证失败",
	"validating":        "正在验证桥梁可用性...",
	"validation_result": "验证结果: 可用 %d, 不可用 %d",

	"export_success": "导出成功",
	"export_failed":  "导出失败",
	"exporting":      "正在导出数据...",
	"export_file":    "已导出到: %s",

	"stats_title":       "统计信息",
	"stats_total":       "总桥梁数",
	"stats_available":   "可用桥梁",
	"stats_unavailable": "不可用桥梁",
	"stats_unknown":     "未知状态",
	"stats_avg_time":    "平均响应时间",
	"stats_last_fetch":  "最后采集",

	"error_network":  "网络错误",
	"error_parse":    "解析错误",
	"error_database": "数据库错误",
	"error_timeout":  "连接超时",
	"error_proxy":    "代理错误",

	"bridge_available":   "可用",
	"bridge_unavailable": "不可用",
	"bridge_unknown":     "未知",

	"query_success": "查询完成",
	"query_failed":  "查询失败",
	"querying":      "正在查询桥梁数据...",

	"import_success":     "导入完成",
	"import_failed":      "导入失败",
	"import_file_failed": "导入文件失败",
	"importing":          "正在导入数据...",
	"import_total":       "文件总数",
	"import_imported":    "已导入",
	"import_skipped":     "已跳过(重复)",

	"cmd_init":     "初始化配置文件和数据库",
	"cmd_fetch":    "采集桥梁数据",
	"cmd_validate": "验证桥梁可用性",
	"cmd_export":   "导出桥梁数据",
	"cmd_stats":    "显示统计信息",
	"cmd_query":    "按条件查询桥梁",
	"cmd_import":   "从文件导入桥梁",

	"flag_config":  "配置文件路径",
	"flag_lang":    "语言 (en/zh)",
	"flag_proxy":   "代理服务器地址",
	"flag_timeout": "超时时间(秒)",
	"flag_workers": "并发数",
	"flag_format":  "输出格式 (torrc/json/all)",
	"flag_output":  "输出目录",
	"flag_period":  "统计周期 (day/week/month)",

	"help_init":     "运行 init 命令初始化配置文件和数据库",
	"help_fetch":    "运行 fetch 命令从 Tor 服务器采集桥接节点",
	"help_validate": "运行 validate 命令测试桥接节点连通性",
	"help_export":   "运行 export 命令导出桥接数据",
	"help_stats":    "运行 stats 命令查看统计信息",
}
