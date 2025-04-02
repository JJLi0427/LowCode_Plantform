package main

import (
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
)

// Locl IP
func getLocalIPs() ([]string, []string, error) {
    var ipv4s, ipv6s []string

    interfaces, err := net.Interfaces()
    if err != nil {
        return nil, nil, err
    }

    for _, iface := range interfaces {
        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }

        for _, addr := range addrs {
            switch v := addr.(type) {
            case *net.IPNet:
                ip := v.IP
                if ip.IsLoopback() {
                    continue
                }
                if ip.To4() != nil {
                    ipv4s = append(ipv4s, ip.String())
                } else if ip.To16() != nil {
                    ipv6s = append(ipv6s, ip.String())
                }
            }
        }
    }

    return ipv4s, ipv6s, nil
}

// Global IP
func getPublicIP(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: ./ipget [IPv4|IPv6|both]")
        return
    }

    args := string(os.Args[1])

    ipv4s, ipv6s, err := getLocalIPs()
    if err != nil {
        fmt.Println("Get local IP error:", err)
        return
    }

    file, err := os.Create("../../../gapplications/ipget/ipget.txt")
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer file.Close()

    if args == "IPv4" || args == "both" {
        file.WriteString("Local IPv4:\n")
        for _, ip := range ipv4s {
            file.WriteString("    " + ip + "\n")
        }

        publicIPv4, err := getPublicIP("http://4.ipw.cn")
        if err != nil {
            file.WriteString("Get global IPv4 error: " + err.Error() + "\n")
        } else {
            file.WriteString("Global IPv4:\n")
            file.WriteString("    " + publicIPv4 + "\n")
        }
    }
    if args == "IPv6" || args == "both" {
        file.WriteString("Local IPv6:\n")
        for _, ip := range ipv6s {
            file.WriteString("    " + ip + "\n")
        }

        publicIPv6, err := getPublicIP("http://6.ipw.cn")
        if err != nil {
            file.WriteString("Get global IPv6 error: " + err.Error() + "\n")
        } else {
            file.WriteString("Global IPv6:\n")
            file.WriteString("    " + publicIPv6 + "\n")
        }
    }
}