//go:build production

package main

// IsDevBuild 在 production 构建（即 wails3 build）时返回 false
func IsDevBuild() bool {
	return false
}
