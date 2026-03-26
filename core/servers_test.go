package core

import "testing"

func TestGetDownloadServers_Global(t *testing.T) {
	servers := GetDownloadServers(RegionGlobal)
	if len(servers) == 0 {
		t.Fatal("expected global download servers, got none")
	}
	for _, s := range servers {
		if s.Region != RegionGlobal {
			t.Errorf("expected region %s, got %s for server %s", RegionGlobal, s.Region, s.Name)
		}
	}
}

func TestGetDownloadServers_CN(t *testing.T) {
	servers := GetDownloadServers(RegionCN)
	if len(servers) == 0 {
		t.Fatal("expected CN download servers, got none")
	}
	for _, s := range servers {
		if s.Region != RegionCN {
			t.Errorf("expected region %s, got %s for server %s", RegionCN, s.Region, s.Name)
		}
	}
}

func TestGetDownloadServers_Auto(t *testing.T) {
	servers := GetDownloadServers(RegionAuto)
	if len(servers) != len(DownloadServers) {
		t.Errorf("auto should return all %d servers, got %d", len(DownloadServers), len(servers))
	}
}

func TestGetUploadServers_Global(t *testing.T) {
	servers := GetUploadServers(RegionGlobal)
	if len(servers) == 0 {
		t.Fatal("expected global upload servers, got none")
	}
	for _, s := range servers {
		if s.Region != RegionGlobal {
			t.Errorf("expected region %s, got %s", RegionGlobal, s.Region)
		}
	}
}

func TestGetUploadServers_CN(t *testing.T) {
	servers := GetUploadServers(RegionCN)
	if len(servers) == 0 {
		t.Fatal("expected CN upload servers, got none")
	}
}

func TestGetUploadServers_Auto(t *testing.T) {
	servers := GetUploadServers(RegionAuto)
	if len(servers) != len(UploadServers) {
		t.Errorf("auto should return all %d servers, got %d", len(UploadServers), len(servers))
	}
}

func TestGetPingTargets_Global(t *testing.T) {
	targets := GetPingTargets(RegionGlobal)
	if len(targets) == 0 {
		t.Fatal("expected global ping targets, got none")
	}
}

func TestGetPingTargets_CN(t *testing.T) {
	targets := GetPingTargets(RegionCN)
	if len(targets) == 0 {
		t.Fatal("expected CN ping targets, got none")
	}
}

func TestGetPingTargets_Auto(t *testing.T) {
	targets := GetPingTargets(RegionAuto)
	cnTargets := GetPingTargets(RegionCN)
	globalTargets := GetPingTargets(RegionGlobal)
	expected := len(cnTargets) + len(globalTargets)
	if len(targets) != expected {
		t.Errorf("auto should return %d targets, got %d", expected, len(targets))
	}
}

func TestGetDownloadServers_UnknownRegion(t *testing.T) {
	servers := GetDownloadServers(Region("mars"))
	if len(servers) != 0 {
		t.Errorf("unknown region should return 0 servers, got %d", len(servers))
	}
}

func TestGetUploadServers_UnknownRegion(t *testing.T) {
	servers := GetUploadServers(Region("mars"))
	if len(servers) != 0 {
		t.Errorf("unknown region should return 0 servers, got %d", len(servers))
	}
}

func TestGetPingTargets_UnknownRegion(t *testing.T) {
	targets := GetPingTargets(Region("mars"))
	if targets != nil {
		t.Errorf("unknown region should return nil, got %v", targets)
	}
}

func TestServerFieldsNotEmpty(t *testing.T) {
	for _, s := range DownloadServers {
		if s.Name == "" || s.URL == "" || s.Region == "" || s.Provider == "" {
			t.Errorf("download server has empty field: %+v", s)
		}
	}
	for _, s := range UploadServers {
		if s.Name == "" || s.URL == "" || s.Region == "" || s.Provider == "" {
			t.Errorf("upload server has empty field: %+v", s)
		}
	}
}
