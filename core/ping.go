// Package core core/ping.go
package core

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"speedgo/commands"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type PingConfig struct {
	Targets     []string
	Count       int
	Timeout     time.Duration
	Concurrency int
	Region      Region
	Verbose     bool
}

type PingResult struct {
	Target string
	RTTs   []time.Duration
	MinRTT time.Duration
	MaxRTT time.Duration
	AvgRTT time.Duration
	Lost   int
	Errors []error
}

type pingSession struct {
	conn   *icmp.PacketConn
	id     int
	seq    int
	target string
}

// splitAndTrim 分割并清理字符串
func splitAndTrim(input, sep string) []string {
	parts := strings.Split(input, sep)
	var result []string
	for _, item := range parts {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitTargets 分割并验证目标地址
func splitTargets(targets string) []string {
	var result []string
	for _, target := range splitAndTrim(targets, ",") {
		if net.ParseIP(target) != nil || isValidHostname(target) {
			result = append(result, target)
		}
	}
	return result
}

// isValidHostname 验证主机名
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 255 {
		return false
	}
	return !strings.ContainsAny(hostname, " ")
}

func NewPingConfig(args []string) (*PingConfig, error) {
	cmd := commands.PingCmd
	if err := cmd.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing ping arguments: %w", err)
	}

	targetsStr := cmd.Lookup("targets").Value.String()
	count := cmd.Lookup("count").Value.(flag.Getter).Get().(int)
	timeout := cmd.Lookup("timeout").Value.(flag.Getter).Get().(time.Duration)
	concurrency := cmd.Lookup("concurrency").Value.(flag.Getter).Get().(int)
	region := Region(cmd.Lookup("region").Value.String())
	verbose := cmd.Lookup("verbose").Value.(flag.Getter).Get().(bool)

	var targets []string
	if targetsStr != "" {
		targets = splitTargets(targetsStr)
	} else {
		targets = GetPingTargets(region)
	}
	if len(targets) == 0 {
		return nil, errors.New("no valid targets provided")
	}

	return &PingConfig{
		Targets:     targets,
		Count:       count,
		Timeout:     timeout,
		Concurrency: concurrency,
		Region:      region,
		Verbose:     verbose,
	}, nil
}

func RunPing(ctx context.Context, args []string) error {
	config, err := NewPingConfig(args)
	if err != nil {
		return err
	}

	fmt.Printf("Starting ping test to %d targets...\n", len(config.Targets))
	results := pingTargets(ctx, config)
	printResults(results)
	return nil
}

func pingTargets(ctx context.Context, config *PingConfig) []PingResult {
	results := make([]PingResult, len(config.Targets))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Concurrency)

	for i, target := range config.Targets {
		wg.Add(1)
		go func(idx int, target string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = pingTarget(ctx, target, config)
			if config.Verbose {
				fmt.Printf("Completed ping to %s\n", target)
			}
		}(i, target)
	}

	wg.Wait()
	return results
}

func pingTarget(ctx context.Context, target string, config *PingConfig) PingResult {
	result := PingResult{
		Target: target,
		RTTs:   make([]time.Duration, 0, config.Count),
	}

	ipAddr, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("resolving address: %w", err))
		result.Lost = config.Count
		return result
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// ICMP not available (no root), fall back to TCP ping
		if config.Verbose {
			fmt.Printf("ICMP not available, using TCP ping for %s\n", target)
		}
		return tcpPingTarget(ctx, target, config)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("closing connection: %w", err))
		}
	}()

	session := &pingSession{
		conn:   conn,
		id:     os.Getpid() & 0xffff,
		seq:    1,
		target: ipAddr.String(),
	}

	for i := 0; i < config.Count; i++ {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, ctx.Err())
			return result
		default:
			rtt, err := session.ping(config.Timeout)
			if err != nil {
				if config.Verbose {
					fmt.Printf("Ping %s failed: %v\n", target, err)
				}
				result.Lost++
				result.Errors = append(result.Errors, err)
			} else {
				result.RTTs = append(result.RTTs, rtt)
				if config.Verbose {
					fmt.Printf("Ping %s: RTT = %v\n", target, rtt)
				}
			}
			session.seq++
			time.Sleep(time.Second)
		}
	}

	result.calculateStats()
	return result
}

// tcpPingTarget measures latency by TCP connection to port 80/443.
// Used as fallback when ICMP is not available (no root privileges).
func tcpPingTarget(ctx context.Context, target string, config *PingConfig) PingResult {
	result := PingResult{
		Target: target,
		RTTs:   make([]time.Duration, 0, config.Count),
	}

	port := "80"
	// Try HTTPS port for common domains
	addr := net.JoinHostPort(target, port)

	for i := 0; i < config.Count; i++ {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, ctx.Err())
			return result
		default:
			rtt, err := tcpPing(ctx, addr, config.Timeout)
			if err != nil {
				if config.Verbose {
					fmt.Printf("TCP Ping %s failed: %v\n", target, err)
				}
				result.Lost++
				result.Errors = append(result.Errors, err)
			} else {
				result.RTTs = append(result.RTTs, rtt)
				if config.Verbose {
					fmt.Printf("TCP Ping %s: RTT = %v\n", target, rtt)
				}
			}
			if i < config.Count-1 {
				time.Sleep(time.Second)
			}
		}
	}

	result.calculateStats()
	return result
}

// tcpPing measures a single TCP connection round-trip time.
func tcpPing(ctx context.Context, addr string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var d net.Dialer
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", addr)
	rtt := time.Since(start)
	if err != nil {
		return 0, fmt.Errorf("TCP connect: %w", err)
	}
	conn.Close()
	return rtt, nil
}

func (s *pingSession) ping(timeout time.Duration) (time.Duration, error) {
	// 生成随机数据作为 payload
	payload := make([]byte, 56) // 标准 ping 使用 56 字节
	rand.Read(payload)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   s.id,
			Seq:  s.seq,
			Data: payload,
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, fmt.Errorf("marshaling ICMP message: %w", err)
	}

	// 在发送前刷新任何待处理的响应
	if err := s.conn.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
		return 0, fmt.Errorf("setting flush deadline: %w", err)
	}
	for {
		_, _, err := s.conn.ReadFrom(make([]byte, 1500))
		if err != nil {
			break
		}
	}

	start := time.Now()
	_, err = s.conn.WriteTo(msgBytes, &net.IPAddr{IP: net.ParseIP(s.target)})
	if err != nil {
		return 0, fmt.Errorf("sending ICMP message: %w", err)
	}

	if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, fmt.Errorf("setting read deadline: %w", err)
	}

	reply := make([]byte, 1500)
	n, _, err := s.conn.ReadFrom(reply)
	if err != nil {
		return 0, fmt.Errorf("reading ICMP reply: %w", err)
	}

	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		return 0, fmt.Errorf("parsing ICMP reply: %w", err)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		echo, ok := rm.Body.(*icmp.Echo)
		if !ok {
			return 0, errors.New("invalid ICMP echo reply")
		}
		if echo.ID != s.id || echo.Seq != s.seq {
			return 0, errors.New("received wrong ICMP reply")
		}
		return time.Since(start), nil
	default:
		return 0, fmt.Errorf("received ICMP message of unexpected type: %v", rm.Type)
	}
}

func (r *PingResult) calculateStats() {
	if len(r.RTTs) == 0 {
		return
	}

	r.MinRTT = r.RTTs[0]
	r.MaxRTT = r.RTTs[0]
	var total time.Duration

	for _, rtt := range r.RTTs {
		total += rtt
		if rtt < r.MinRTT {
			r.MinRTT = rtt
		}
		if rtt > r.MaxRTT {
			r.MaxRTT = rtt
		}
	}
	r.AvgRTT = total / time.Duration(len(r.RTTs))
}

func printResults(results []PingResult) {
	fmt.Println("\nPING STATISTICS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%-20s %10s %10s %10s %12s\n", "TARGET", "MIN", "AVG", "MAX", "LOSS")
	fmt.Println(strings.Repeat("-", 60))

	for _, result := range results {
		if len(result.RTTs) == 0 {
			fmt.Printf("%-20s %10s %10s %10s %11d%%\n",
				result.Target,
				"N/A",
				"N/A",
				"N/A",
				100)

			if len(result.Errors) > 0 {
				fmt.Printf("  Errors:\n")
				for _, err := range result.Errors {
					fmt.Printf("  - %v\n", err)
				}
			}
		} else {
			lossPercent := float64(result.Lost) * 100 / float64(len(result.RTTs)+result.Lost)

			// 格式化延迟值，统一使用毫秒为单位
			_min := float64(result.MinRTT.Microseconds()) / 1000
			_avg := float64(result.AvgRTT.Microseconds()) / 1000
			_max := float64(result.MaxRTT.Microseconds()) / 1000

			fmt.Printf("%-20s %9.1fms %9.1fms %9.1fms %10.1f%%\n",
				result.Target,
				_min,
				_avg,
				_max,
				lossPercent)
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}

const protocolICMP = 1
