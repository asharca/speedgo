package core

import "net/http"

const userAgent = "Mozilla/5.0 (compatible; SpeedGo/1.0)"

// SetUA sets the User-Agent header on a request.
func SetUA(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
}

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
// All URLs verified with Go HTTP client (curl + go run) as of 2026-03.
// Cloudflare __down removed: returns 403 due to WAF/TLS fingerprint detection.
var DownloadServers = []TestServer{
	// Global — accessible from outside China; slow from CN but reachable
	{Name: "OVH", URL: "https://proof.ovh.net/files/100Mb.dat", Region: RegionGlobal, Provider: "OVH"},
	// China — verified: no redirects, repeatable, Go HTTP client compatible
	{Name: "Huawei Cloud", URL: "https://mirrors.huaweicloud.com/debian/dists/stable/main/Contents-amd64.gz", Region: RegionCN, Provider: "Huawei"},
	{Name: "Aliyun", URL: "https://mirrors.aliyun.com/debian/dists/stable/main/Contents-amd64.gz", Region: RegionCN, Provider: "Aliyun"},
	{Name: "Tencent Cloud", URL: "https://mirrors.cloud.tencent.com/debian/dists/stable/main/Contents-amd64.gz", Region: RegionCN, Provider: "Tencent"},
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
