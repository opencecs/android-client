//go:build !production

package main

// IsDevBuild 在非 production 构建（即 wails3 dev）时返回 true
func IsDevBuild() bool {
	return true
}
