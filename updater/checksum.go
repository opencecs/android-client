package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"strings"
)

type ChecksumType string

const (
	ChecksumSHA256 ChecksumType = "sha256"
	ChecksumSHA512 ChecksumType = "sha512"
)

type Verifier struct {
	publicKeyPath string
}

func NewVerifier(publicKeyPath string) *Verifier {
	return &Verifier{
		publicKeyPath: publicKeyPath,
	}
}

func (v *Verifier) VerifyFile(filePath, expectedChecksum string) (bool, error) {
	log.Printf("[Verifier] 开始校验文件: %s", filePath)
	log.Printf("[Verifier] 预期校验和: %s", expectedChecksum)

	// 校验和为空，直接跳过
	if strings.TrimSpace(expectedChecksum) == "" {
		log.Printf("[Verifier] 预期校验和为空，跳过校验")
		return true, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	log.Printf("[Verifier] 文件大小: %d bytes", fileInfo.Size())

	checksumType, checksumValue := v.parseChecksum(expectedChecksum)
	preview := checksumValue
	if len(preview) > 16 {
		preview = preview[:16]
	}
	log.Printf("[Verifier] 校验和类型: %s, 校验值前16位: %s", checksumType, preview)

	var hasher hash.Hash
	switch checksumType {
	case ChecksumSHA256:
		hasher = sha256.New()
	case ChecksumSHA512:
		hasher = sha256.New()
	default:
		hasher = sha256.New()
	}

	writer := &HashWriter{
		Hash: hasher,
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return false, fmt.Errorf("计算校验和失败: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("[Verifier] 实际校验和: %s", actualChecksum)

	normalizedExpected := v.normalizeChecksum(expectedChecksum)
	normalizedActual := v.normalizeChecksum(actualChecksum)

	if normalizedExpected == normalizedActual {
		log.Printf("[Verifier] 校验成功")
		return true, nil
	}

	log.Printf("[Verifier] 校验失败")
	return false, nil
}

func (v *Verifier) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", fmt.Errorf("计算校验和失败: %w", err)
	}

	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func (v *Verifier) CompareChecksums(expected, actual string) bool {
	normExpected := v.normalizeChecksum(expected)
	normActual := v.normalizeChecksum(actual)
	return normExpected == normActual
}

func (v *Verifier) parseChecksum(checksum string) (ChecksumType, string) {
	checksum = strings.TrimSpace(checksum)

	if strings.HasPrefix(checksum, "sha512:") {
		return ChecksumSHA512, strings.TrimPrefix(checksum, "sha512:")
	}

	if strings.HasPrefix(checksum, "sha256:") {
		return ChecksumSHA256, strings.TrimPrefix(checksum, "sha256:")
	}

	if len(checksum) == 128 {
		return ChecksumSHA512, checksum
	}

	if len(checksum) == 64 {
		return ChecksumSHA256, checksum
	}

	return ChecksumSHA256, checksum
}

func (v *Verifier) normalizeChecksum(checksum string) string {
	checksum = strings.TrimSpace(checksum)
	checksum = strings.ToLower(checksum)
	checksum = strings.TrimPrefix(checksum, "sha256:")
	checksum = strings.TrimPrefix(checksum, "sha512:")
	return checksum
}

type HashWriter struct {
	Hash hash.Hash
}

func (hw *HashWriter) Write(p []byte) (n int, err error) {
	return hw.Hash.Write(p)
}

type SignatureVerifier struct {
	publicKeyPath string
}

func NewSignatureVerifier(publicKeyPath string) *SignatureVerifier {
	return &SignatureVerifier{
		publicKeyPath: publicKeyPath,
	}
}

func (v *SignatureVerifier) VerifySignature(filePath, signature string) (bool, error) {
	log.Printf("[SignatureVerifier] 签名验证功能需要额外的加密库支持")

	return false, fmt.Errorf("签名验证功能尚未实现")
}

func (v *SignatureVerifier) LoadPublicKey() (interface{}, error) {
	if v.publicKeyPath == "" {
		return nil, fmt.Errorf("未配置公钥路径")
	}

	data, err := os.ReadFile(v.publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %w", err)
	}

	return data, nil
}

type ChecksumCache struct {
	checksums map[string]string
}

func NewChecksumCache() *ChecksumCache {
	return &ChecksumCache{
		checksums: make(map[string]string),
	}
}

func (c *ChecksumCache) Get(filePath string) (string, bool) {
	val, ok := c.checksums[filePath]
	return val, ok
}

func (c *ChecksumCache) Set(filePath, checksum string) {
	c.checksums[filePath] = checksum
}

func (c *ChecksumCache) Clear() {
	c.checksums = make(map[string]string)
}
