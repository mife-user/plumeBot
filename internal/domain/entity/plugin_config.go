package entity

// PluginConfig 表示某个插件在某个群的配置，以 JSON blob 存储。
type PluginConfig struct {
	GroupID    string // 群 ID
	PluginName string // 插件名称
	Config     string // 配置 JSON
}
