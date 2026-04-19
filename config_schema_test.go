package kstypes

import (
	"encoding/hex"
	"strings"
	"testing"
)

// hexNoSpace 去掉十六进制字符串中的空格，便于将 testvectors.json 的
// expected_bytes_hex 直接嵌入测试。
func hexNoSpace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// aadTestCase 对应 testvectors.json 的 aad_canonical 条目。
type aadTestCase struct {
	name            string
	mcpServerID     string
	configVersion   uint64
	fingerprint     string
	expectedBytesHex string // 无空格
}

// 12 条样本直接来自 conformance/config-schema/testvectors.json（feature/spec-a-m0i SHA 9c7e0a4）
var aadTestCases = []aadTestCase{
	{
		name:          "basic_ascii_id",
		mcpServerID:   "ks-mcp-image-gen",
		configVersion: 2,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 10 6b 73 2d 6d 63 70 2d 69 6d 61 67 65 2d 67 65 6e 00 00 00 00 00 00 00 02 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "short_id",
		mcpServerID:   "a",
		configVersion: 0,
		fingerprint:   "6668:7aad:f862:bd77:6c8f:c18b:8e9f:8e20",
		expectedBytesHex: hexNoSpace("00 01 61 00 00 00 00 00 00 00 00 00 27 36 36 36 38 3a 37 61 61 64 3a 66 38 36 32 3a 62 64 37 37 3a 36 63 38 66 3a 63 31 38 62 3a 38 65 39 66 3a 38 65 32 30"),
	},
	{
		name:          "utf8_chinese_id",
		mcpServerID:   "ks-mcp-测试",
		configVersion: 1,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0d 6b 73 2d 6d 63 70 2d e6 b5 8b e8 af 95 00 00 00 00 00 00 00 01 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "max_length_id",
		mcpServerID:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		configVersion: 1,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 ff 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 00 00 00 00 00 00 00 01 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "version_zero",
		mcpServerID:   "ks-mcp-test",
		configVersion: 0,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 00 00 00 00 00 00 00 00 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "version_max_int63",
		mcpServerID:   "ks-mcp-test",
		configVersion: 9223372036854775807,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 7f ff ff ff ff ff ff ff 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "version_large",
		mcpServerID:   "ks-mcp-test",
		configVersion: 1000000000,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 00 00 00 00 3b 9a ca 00 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "fingerprint_all_zeros",
		mcpServerID:   "ks-mcp-test",
		configVersion: 1,
		fingerprint:   "6668:7aad:f862:bd77:6c8f:c18b:8e9f:8e20",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 00 00 00 00 00 00 00 01 00 27 36 36 36 38 3a 37 61 61 64 3a 66 38 36 32 3a 62 64 37 37 3a 36 63 38 66 3a 63 31 38 62 3a 38 65 39 66 3a 38 65 32 30"),
	},
	{
		name:          "fingerprint_all_f",
		mcpServerID:   "ks-mcp-test",
		configVersion: 1,
		fingerprint:   "af96:1376:0f72:635f:bdb4:4a5a:0a63:c39f",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 00 00 00 00 00 00 00 01 00 27 61 66 39 36 3a 31 33 37 36 3a 30 66 37 32 3a 36 33 35 66 3a 62 64 62 34 3a 34 61 35 61 3a 30 61 36 33 3a 63 33 39 66"),
	},
	{
		name:          "fingerprint_mixed",
		mcpServerID:   "ks-mcp-test",
		configVersion: 1,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 2d 6d 63 70 2d 74 65 73 74 00 00 00 00 00 00 00 01 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "id_with_dash",
		mcpServerID:   "ks-mcp-image-gen",
		configVersion: 5,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 10 6b 73 2d 6d 63 70 2d 69 6d 61 67 65 2d 67 65 6e 00 00 00 00 00 00 00 05 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
	{
		name:          "id_with_underscore",
		mcpServerID:   "ks_mcp_test",
		configVersion: 3,
		fingerprint:   "ab12:cd34:ef56:7890:1234:5678:9abc:def0",
		expectedBytesHex: hexNoSpace("00 0b 6b 73 5f 6d 63 70 5f 74 65 73 74 00 00 00 00 00 00 00 03 00 27 61 62 31 32 3a 63 64 33 34 3a 65 66 35 36 3a 37 38 39 30 3a 31 32 33 34 3a 35 36 37 38 3a 39 61 62 63 3a 64 65 66 30"),
	},
}

func TestAADCanonicalBytes(t *testing.T) {
	t.Parallel()
	for _, tc := range aadTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			want, err := hex.DecodeString(tc.expectedBytesHex)
			if err != nil {
				t.Fatalf("测试数据 expectedBytesHex 解码失败: %v", err)
			}
			got := AADCanonicalBytes(tc.mcpServerID, tc.configVersion, tc.fingerprint)
			if string(got) != string(want) {
				t.Errorf("AADCanonicalBytes(%q, %d, %q)\n  got  hex: %x\n  want hex: %x",
					tc.mcpServerID, tc.configVersion, tc.fingerprint, got, want)
			}
		})
	}
}

func TestFingerprint_AllZero(t *testing.T) {
	t.Parallel()
	pubkey := make([]byte, 32)
	got := Fingerprint(pubkey)
	// sha256(32×0x00) 前 16 字节 = 66687aadf862bd776c8fc18b8e9f8e20
	want := "6668:7aad:f862:bd77:6c8f:c18b:8e9f:8e20"
	if got != want {
		t.Errorf("Fingerprint(zero32) = %q, want %q", got, want)
	}
}

func TestFingerprint_AllFF(t *testing.T) {
	t.Parallel()
	pubkey := make([]byte, 32)
	for i := range pubkey {
		pubkey[i] = 0xff
	}
	got := Fingerprint(pubkey)
	// sha256(32×0xff) 前 16 字节 = af9613760f72635fbdb44a5a0a63c39f
	want := "af96:1376:0f72:635f:bdb4:4a5a:0a63:c39f"
	if got != want {
		t.Errorf("Fingerprint(ff32) = %q, want %q", got, want)
	}
}

func TestFingerprint_Panic_NonStandardLength(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("期望 31 字节输入触发 panic，但未 panic")
		}
	}()
	Fingerprint(make([]byte, 31))
}
