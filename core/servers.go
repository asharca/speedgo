package core

// Region represents a geographic region for server selection.
type Region string

const (
	RegionCN     Region = "cn"
	RegionGlobal Region = "global"
	RegionAuto   Region = "auto"
)

// TestServer holds information about a speed test server.
type TestServer struct {
	Name     string
	URL      string
	Region   Region
	Provider string
}

// DownloadServers lists available download test endpoints.
var DownloadServers = []TestServer{
	// Global
	{Name: "Cloudflare", URL: "https://speed.cloudflare.com/__down?bytes=100000000", Region: RegionGlobal, Provider: "Cloudflare"},
	{Name: "OVH", URL: "https://proof.ovh.net/files/100Mb.dat", Region: RegionGlobal, Provider: "OVH"},
	{Name: "Hetzner", URL: "https://speed.hetzner.de/100MB.bin", Region: RegionGlobal, Provider: "Hetzner"},
	// China — Cloudflare has CN PoPs, reliable domestic CDN mirrors with large files
	{Name: "Cloudflare CN", URL: "https://speed.cloudflare.com/__down?bytes=100000000", Region: RegionCN, Provider: "Cloudflare"},
	{Name: "Aliyun", URL: "https://mirrors.aliyun.com/debian-cd/current/amd64/iso-cd/debian-12.9.0-amd64-netinst.iso", Region: RegionCN, Provider: "Aliyun"},
	{Name: "Huawei Cloud", URL: "https://mirrors.huaweicloud.com/debian-cd/current/amd64/iso-cd/debian-12.9.0-amd64-netinst.iso", Region: RegionCN, Provider: "Huawei"},
	{Name: "Tencent Cloud", URL: "https://mirrors.cloud.tencent.com/debian-cd/current/amd64/iso-cd/debian-12.9.0-amd64-netinst.iso", Region: RegionCN, Provider: "Tencent"},
	{Name: "163 Mirror", URL: "https://mirrors.163.com/debian-cd/current/amd64/iso-cd/debian-12.9.0-amd64-netinst.iso", Region: RegionCN, Provider: "NetEase"},
}

// UploadServers lists available upload test endpoints.
var UploadServers = []TestServer{
	// Global
	{Name: "Cloudflare", URL: "https://speed.cloudflare.com/__up", Region: RegionGlobal, Provider: "Cloudflare"},
	// China — LibreSpeed compatible endpoints
	{Name: "Cloudflare (via CN)", URL: "https://speed.cloudflare.com/__up", Region: RegionCN, Provider: "Cloudflare"},
}

// PingTargets lists default ping targets per region.
var PingTargets = map[Region][]string{
	RegionGlobal: {
		"cloudflare.com",
		"google.com",
		"amazon.com",
	},
	RegionCN: {
		"baidu.com",
		"aliyun.com",
		"tencent.com",
		"qq.com",
		"bilibili.com",
	},
}

// GetDownloadServers returns servers filtered by region.
// RegionAuto returns all servers.
func GetDownloadServers(region Region) []TestServer {
	if region == RegionAuto {
		return DownloadServers
	}
	var result []TestServer
	for _, s := range DownloadServers {
		if s.Region == region {
			result = append(result, s)
		}
	}
	return result
}

// GetUploadServers returns upload servers filtered by region.
func GetUploadServers(region Region) []TestServer {
	if region == RegionAuto {
		return UploadServers
	}
	var result []TestServer
	for _, s := range UploadServers {
		if s.Region == region {
			result = append(result, s)
		}
	}
	return result
}

// GetPingTargets returns ping targets for the given region.
// RegionAuto returns both CN and global targets.
func GetPingTargets(region Region) []string {
	if region == RegionAuto {
		var all []string
		all = append(all, PingTargets[RegionCN]...)
		all = append(all, PingTargets[RegionGlobal]...)
		return all
	}
	return PingTargets[region]
}
