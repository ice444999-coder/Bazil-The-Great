/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"gorm.io/gorm"
)

// SystemHealthController handles system health monitoring
type SystemHealthController struct {
	DB *gorm.DB
}

// NewSystemHealthController creates a new controller
func NewSystemHealthController(db *gorm.DB) *SystemHealthController{
	return &SystemHealthController{DB: db}
}

// SystemHealthResponse represents the complete system health data
type SystemHealthResponse struct {
	Timestamp      time.Time       `json:"timestamp"`
	CPU            CPUMetrics      `json:"cpu"`
	Memory         MemoryMetrics   `json:"memory"`
	Disk           DiskMetrics     `json:"disk"`
	PostgreSQL     PostgreSQLMetrics `json:"postgresql"`
	Network        NetworkMetrics  `json:"network"`
	System         SystemInfo      `json:"system"`
}

type CPUMetrics struct {
	Usage       float64 `json:"usage_percent"`
	Temperature float64 `json:"temperature_celsius"`
	Cores       int     `json:"cores"`
	LogicalCPUs int     `json:"logical_cpus"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
	UsedGB      float64 `json:"used_gb"`
	TotalGB     float64 `json:"total_gb"`
}

type DiskMetrics struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
	UsedGB      float64 `json:"used_gb"`
	TotalGB     float64 `json:"total_gb"`
	ReadSpeed   uint64  `json:"read_bytes_per_sec"`
	WriteSpeed  uint64  `json:"write_bytes_per_sec"`
}

type PostgreSQLMetrics struct {
	DatabaseSize   int64   `json:"database_size_bytes"`
	DatabaseSizeMB float64 `json:"database_size_mb"`
	ConnectionCount int    `json:"connection_count"`
	ActiveQueries  int     `json:"active_queries"`
}

type NetworkMetrics struct {
	BytesSent     uint64  `json:"bytes_sent"`
	BytesReceived uint64  `json:"bytes_received"`
	SendRateMBps  float64 `json:"send_rate_mbps"`
	RecvRateMBps  float64 `json:"recv_rate_mbps"`
	Latency       float64 `json:"latency_ms"`
}

type SystemInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	Uptime          uint64 `json:"uptime_seconds"`
	GoVersion       string `json:"go_version"`
}

// GetHealth returns comprehensive system health metrics
// @Summary Get system health metrics
// @Description Get detailed system metrics including CPU, RAM, disk, PostgreSQL, and network stats
// @Tags System
// @Produce json
// @Success 200 {object} SystemHealthResponse
// @Router /api/v1/system/health [get]
func (ctrl *SystemHealthController) GetHealth(c *gin.Context) {
	response := SystemHealthResponse{
		Timestamp: time.Now(),
	}

	// CPU Metrics
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		response.CPU.Usage = cpuPercent[0]
	}
	response.CPU.Cores = runtime.NumCPU()
	response.CPU.LogicalCPUs = runtime.NumCPU()
	
	// CPU Temperature (Windows-specific, may need WMI or external tool)
	temps, err := host.SensorsTemperatures()
	if err == nil && len(temps) > 0 {
		response.CPU.Temperature = temps[0].Temperature
	} else {
		response.CPU.Temperature = 0 // Temperature monitoring not available on Windows without additional tools
	}

	// Memory Metrics
	vmStat, _ := mem.VirtualMemory()
	response.Memory.Total = vmStat.Total
	response.Memory.Used = vmStat.Used
	response.Memory.Free = vmStat.Free
	response.Memory.UsedPercent = vmStat.UsedPercent
	response.Memory.UsedGB = float64(vmStat.Used) / (1024 * 1024 * 1024)
	response.Memory.TotalGB = float64(vmStat.Total) / (1024 * 1024 * 1024)

	// Disk Metrics (C: drive on Windows)
	diskStat, _ := disk.Usage("C:\\")
	if diskStat != nil {
		response.Disk.Total = diskStat.Total
		response.Disk.Used = diskStat.Used
		response.Disk.Free = diskStat.Free
		response.Disk.UsedPercent = diskStat.UsedPercent
		response.Disk.UsedGB = float64(diskStat.Used) / (1024 * 1024 * 1024)
		response.Disk.TotalGB = float64(diskStat.Total) / (1024 * 1024 * 1024)
	}

	// Disk IO (approximation)
	ioCounters, _ := disk.IOCounters()
	for _, io := range ioCounters {
		response.Disk.ReadSpeed += io.ReadBytes
		response.Disk.WriteSpeed += io.WriteBytes
		break // Just take first disk
	}

	// PostgreSQL Metrics
	if ctrl.DB != nil {
		// Database size
		var dbSize int64
		ctrl.DB.Raw("SELECT pg_database_size('solace_db')").Scan(&dbSize)
		response.PostgreSQL.DatabaseSize = dbSize
		response.PostgreSQL.DatabaseSizeMB = float64(dbSize) / (1024 * 1024)

		// Connection count
		ctrl.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity").Scan(&response.PostgreSQL.ConnectionCount)

		// Active queries
		ctrl.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&response.PostgreSQL.ActiveQueries)
	}

	// Network Metrics
	netIO, _ := net.IOCounters(false)
	if len(netIO) > 0 {
		response.Network.BytesSent = netIO[0].BytesSent
		response.Network.BytesReceived = netIO[0].BytesRecv
		response.Network.SendRateMBps = float64(netIO[0].BytesSent) / (1024 * 1024)
		response.Network.RecvRateMBps = float64(netIO[0].BytesRecv) / (1024 * 1024)
	}
	response.Network.Latency = 0 // Placeholder - would need to ping a server

	// System Info
	hostInfo, _ := host.Info()
	hostname, _ := os.Hostname()
	response.System.Hostname = hostname
	response.System.OS = runtime.GOOS
	response.System.Platform = hostInfo.Platform
	response.System.PlatformVersion = hostInfo.PlatformVersion
	response.System.Uptime = hostInfo.Uptime
	response.System.GoVersion = runtime.Version()

	c.JSON(http.StatusOK, response)
}
