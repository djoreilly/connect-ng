package connect

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	archX86  = "x86_64"
	archARM  = "aarch64"
	archS390 = "s390x"
)

type hwinfo struct {
	Hostname      string `json:"hostname"`
	Cpus          int    `json:"cpus"`
	Sockets       int    `json:"sockets"`
	Hypervisor    string `json:"hypervisor"`
	Arch          string `json:"arch"`
	UUID          string `json:"uuid"`
	CloudProvider string `json:"cloud_provider"`
}

func getHwinfo() (hwinfo, error) {
	hw := hwinfo{}
	var err error
	if hw.Arch, err = arch(); err != nil {
		return hwinfo{}, err
	}
	hw.Hostname = hostname()
	hw.CloudProvider = cloudProvider()

	var lscpuM map[string]string
	if hw.Arch == archX86 || hw.Arch == archARM {
		if lscpuM, err = lscpu(); err != nil {
			return hwinfo{}, err
		}
		hw.Cpus, _ = strconv.Atoi(lscpuM["CPU(s)"])
		hw.Sockets, _ = strconv.Atoi(lscpuM["Socket(s)"])
		if hw.UUID, err = uuid(); err != nil {
			return hwinfo{}, err
		}
	}

	if hw.Arch == archX86 {
		hw.Hypervisor = lscpuM["Hypervisor vendor"]
	}

	if hw.Arch == archARM {
		if hw.Hypervisor, err = hypervisor(); err != nil {
			return hwinfo{}, err
		}
	}

	if hw.Arch == archS390 {
		if err := cpuinfoS390(&hw); err != nil {
			return hwinfo{}, err
		}
		hw.UUID = uuidS390()
	}

	return hw, nil
}

func cpuinfoS390(hw *hwinfo) error {
	rvsOut, err := readValues("-s")
	if err != nil {
		return err
	}
	rvs := readValues2map(rvsOut)

	if cpus, ok := rvs["VM00 CPUs Total"]; ok {
		hw.Cpus, _ = strconv.Atoi(cpus)
	} else if cpus, ok := rvs["LPAR CPUs Total"]; ok {
		hw.Cpus, _ = strconv.Atoi(cpus)
	}

	if sockets, ok := rvs["VM00 IFLs"]; ok {
		hw.Sockets, _ = strconv.Atoi(sockets)
	} else if sockets, ok := rvs["LPAR CPUs IFL"]; ok {
		hw.Sockets, _ = strconv.Atoi(sockets)
	}

	if hypervisor, ok := rvs["VM00 Control Program"]; ok {
		// Strip and remove recurring whitespaces e.g. " z/VM    6.1.0" => "z/VM 6.1.0"
		subs := strings.Fields(hypervisor)
		hw.Hypervisor = strings.Join(subs, " ")
	} else {
		Debug.Print("Unable to find 'VM00 Control Program'. This system probably runs on an LPAR.")
	}

	return nil
}

func arch() (string, error) {
	output, err := execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func lscpu() (map[string]string, error) {
	output, err := execute([]string{"lscpu"}, nil)
	if err != nil {
		return nil, err
	}
	return lscpu2map(output), nil
}

func lscpu2map(b []byte) map[string]string {
	m := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		m[key] = val
	}
	return m
}

func cloudProvider() string {
	version, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_version")
	if err != nil {
		return ""
	}
	if bytes.Contains(version, []byte("amazon")) {
		return "amazon"
	}

	vendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/sys_vendor")
	if err != nil {
		return ""
	}
	if bytes.Contains(vendor, []byte("Amazon")) {
		return "Amazon"
	}
	if bytes.Contains(vendor, []byte("Microsoft")) {
		return "Microsoft"
	}
	if bytes.Contains(vendor, []byte("Google")) {
		return "Google"
	}
	return ""
}

func hypervisor() (string, error) {
	output, err := execute([]string{"systemd-detect-virt", "-v"}, []int{0, 1})
	if err != nil {
		return "", err
	}
	if bytes.Equal(output, []byte("none")) {
		return "", nil
	}
	return string(output), nil
}

// uuid returns the system uuid on x86 and arm
func uuid() (string, error) {
	if fileExists("/sys/hypervisor/uuid") {
		content, err := os.ReadFile("/sys/hypervisor/uuid")
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	content, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_uuid")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// uuidS390 returns the system uuid on S390 or "" if it cannot be found
func uuidS390() string {
	out, err := readValues("-u")
	if err != nil {
		return ""
	}
	uuid := string(out)
	if isUUID(uuid) {
		return uuid
	}
	Debug.Print("Not implemented. Unable to determine UUID for s390x. Set to \"\"")
	return ""
}

// isUUID returns true if s is a valid uuid
func isUUID(s string) bool {
	exp := `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	uuidRe := regexp.MustCompile(exp)
	return uuidRe.MatchString(s)
}

// getPrivateIPAddr returns the first private IP address on the host
func getPrivateIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())
		if privateIP(ip) {
			return ip.String(), nil
		}
	}
	return "", nil
}

// privateIP returns true if ip is in a RFC1918 range
func privateIP(ip net.IP) bool {
	for _, block := range []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"} {
		_, ipNet, _ := net.ParseCIDR(block)
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func hostname() string {
	name, err := os.Hostname()
	// TODO the Ruby version has this "(none)" check - why?
	if err == nil && name != "" && name != "(none)" {
		return name
	}
	Debug.Print(err)
	ip, err := getPrivateIPAddr()
	if err != nil {
		Debug.Print(err)
		return ""
	}
	return ip
}

// readValues calls read_values from SUSE/s390-tools
func readValues(arg string) ([]byte, error) {
	output, err := execute([]string{"read_values", arg}, nil)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// readValues2map parses the output of "read_values -s" on s390
func readValues2map(b []byte) map[string]string {
	br := bufio.NewScanner(bytes.NewReader(b))
	m := make(map[string]string)
	for br.Scan() {
		line := br.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		m[key] = val
	}
	return m
}
