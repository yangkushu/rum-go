package utils

import (
	"errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type CpuInfo struct {
	CoreSize   int     `json:"core_size"`
	VendorID   string  `json:"vendor_id"`
	ModelName  string  `json:"model_name"`
	Mhz        float64 `json:"mhz"`
	CacheSize  int32   `json:"cache_size"`
	Family     string  `json:"family"`
	Model      string  `json:"model"`
	Stepping   int32   `json:"stepping"`
	PhysicalID string  `json:"physical_id"`
}

// GetMemoryInfo 获取内存信息
func GetMemoryInfo() (MemoryInfo, error) {
	if memInfo, err := mem.VirtualMemory(); IsFailed(err) {
		return MemoryInfo{}, err
	} else {
		return MemoryInfo{
			Total:       memInfo.Total >> 20,
			Used:        memInfo.Used >> 20,
			UsedPercent: memInfo.UsedPercent,
		}, nil
	}
}

// GetCpuInfo 获取CPU信息
func GetCpuInfo() (CpuInfo, error) {
	if infos, err := cpu.Info(); nil != err {
		return CpuInfo{}, err
	} else {
		coreSize := len(infos)
		if coreSize <= 0 {
			return CpuInfo{}, errors.New("get cpu info fail: core size is 0")
		}

		return CpuInfo{
			CoreSize:   coreSize,
			VendorID:   infos[0].VendorID,
			ModelName:  infos[0].ModelName,
			Mhz:        infos[0].Mhz,
			CacheSize:  infos[0].CacheSize,
			Family:     infos[0].Family,
			Model:      infos[0].Model,
			Stepping:   infos[0].Stepping,
			PhysicalID: infos[0].PhysicalID,
		}, nil
	}
}
