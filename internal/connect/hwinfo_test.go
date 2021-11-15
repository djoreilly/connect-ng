package connect

import (
	"net"
	"testing"
)

func TestLscpu2mapPhysical(t *testing.T) {
	m := lscpu2map(readTestFile("lscpu_phys.txt", t))

	if m["CPU(s)"] != "8" {
		t.Errorf("Found %s CPU(s), expected 8", m["CPU(s)"])
	}
	if m["Socket(s)"] != "1" {
		t.Errorf("Found %s Sockets(s), expected 1", m["Socket(s)"])
	}
	if _, ok := m["Hypervisor vendor"]; ok {
		t.Errorf("Hypervisor vendor should not be set")
	}
}

func TestLscpu2mapVirtual(t *testing.T) {
	m := lscpu2map(readTestFile("lscpu_virt.txt", t))

	if m["CPU(s)"] != "1" {
		t.Errorf("Found %s CPU(s), expected 1", m["CPU(s)"])
	}
	if m["Socket(s)"] != "1" {
		t.Errorf("Found %s Sockets(s), expected 1", m["Socket(s)"])
	}
	if hv, ok := m["Hypervisor vendor"]; !ok || hv != "KVM" {
		t.Errorf("Hypervisor vendor should be KVM")
	}
}

func TestIsUUID(t *testing.T) {
	var tests = []struct {
		s      string
		expect bool
	}{
		{"4C4C4544-0059-4810-8034-C2C04F335931", true},
		{"4C4C4544-0059-7777-8034-C2C04F335931", true},
		{"ec293a33-b805-7eef-b7c8-d1238386386f", true},
		{"failed:\n", false},
	}
	for _, test := range tests {
		got := isUUID(test.s)
		if got != test.expect {
			t.Errorf("Got isUUID(%s)==%v, expected %v", test.s, got, test.expect)
		}
	}
}

func TestPrivateIP(t *testing.T) {
	var tests = []struct {
		ip      string
		private bool
	}{
		{"10.0.1.1", true},
		{"192.168.100.10", true},
		{"172.18.10.10", true},
		{"8.8.8.8", false},
		{"172.15.0.1", false},
	}
	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		got := privateIP(ip)
		if got != test.private {
			t.Errorf("Got privateIP(%s)==%v, expected %v", test.ip, got, test.private)
		}
	}
}

func TestReadValues2map(t *testing.T) {
	m := readValues2map(readTestFile("read_values_s.txt", t))
	expect := map[string]string{
		"VM00 CPUs Total":      "1",
		"LPAR CPUs Total":      "6",
		"VM00 IFLs":            "1",
		"LPAR CPUs IFL":        "6",
		"VM00 Control Program": "z/VM    6.1.0",
	}
	for k, v := range expect {
		if m[k] != expect[k] {
			t.Errorf("m[%s]==%s, expected %s", k, m[k], v)
		}
	}
}
