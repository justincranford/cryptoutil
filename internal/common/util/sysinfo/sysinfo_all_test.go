package sysinfo

import (
	"testing"
	"time"
)

const expectedSysInfos = 13

func TestSysInfoAll(t *testing.T) {
	all, err := GetAllInfoWithTimeout(mockSysInfoProvider, 5*time.Second)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(all) != expectedSysInfos {
		t.Errorf("Expected %d values, got: %d", expectedSysInfos, len(all))
	}

	for i, value := range all {
		if value == nil {
			t.Errorf("sysinfo[%d] is nil", i)
		} else if len(value) == 0 {
			t.Errorf("sysinfo[%d] is empty", i)
		} else {
			t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
		}
	}
}
