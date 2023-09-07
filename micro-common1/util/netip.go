package util

import (
	"fmt"
	"net"
)

//ip的相关函数

var(
	localStartA,localEndA uint32
	localStartB,localEndB uint32
	localStartC,localEndC uint32
)

func init()  {
	//局域网IP段
	/*10.0.0.0- 10.255.255.255
	172.16.0.0-   172.31.255.255
	192.168.0.0-192.168.255.255*/
	localStartA = uint32(10)<<24 | uint32(0)<<16 | uint32(0)<<8 | uint32(0)
	localEndA = uint32(10)<<24 | uint32(255)<<16 | uint32(255)<<8 | uint32(255)

	localStartB = uint32(172)<<24 | uint32(16)<<16 | uint32(0)<<8 | uint32(0)
	localEndB = uint32(172)<<24 | uint32(31)<<16 | uint32(255)<<8 | uint32(255)

	localStartB = uint32(192)<<24 | uint32(168)<<16 | uint32(0)<<8 | uint32(0)
	localEndB = uint32(192)<<24 | uint32(168)<<16 | uint32(255)<<8 | uint32(255)
}

func Ip2uint32(ip net.IP) uint32 {
	if len(ip) > 4 {
		ip = ip.To4()
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func IP32ToIpV4(ip32 uint32)net.IP  {
	return net.IPv4(byte(ip32 >> 24),byte(ip32 >> 16),byte(ip32 >> 8),byte(ip32))
}

func Ip32ToStr(ip32 uint32)string  {
	return fmt.Sprintf("%d.%d.%d.%d",byte(ip32 >> 24),byte(ip32 >> 16),byte(ip32 >> 8),byte(ip32))
}

func IsLocalNetByIp32(ip32 uint32) bool  {
	return byte(ip32 >> 24) == 127 || ip32 > localStartA && ip32 < localEndA ||
		ip32 > localStartB && ip32 < localEndB ||
		ip32 > localStartC && ip32 < localEndC
}

//IP是否是局域网IP
func IsLocalNet(ip net.IP) bool  {
	return IsLocalNetByIp32(Ip2uint32(ip))
}


//获取一个mac地址
func GetMacAddrs()string  {
	netInterfaces,err := net.Interfaces()
	if err != nil{
		return ""
	}
	for _,netinter := range netInterfaces{
		macAddr := netinter.HardwareAddr.String()
		if len(macAddr) > 0{
			return macAddr
		}
	}
	return ""
}

//获得一个公网IP
func GetWanNetIP()(string,uint32)  {
	iaddr,err := net.InterfaceAddrs()
	resultaddr := ""
	ip32 := uint32(0)
	if err == nil{
		for _,addr := range iaddr{
			ipNet,ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback(){
				//判定是否是广域网
				ipv4 := ipNet.IP.To4()
				if ipv4 == nil{
					continue
				}
				if !IsLocalNet(ipv4){
					return ipv4.String(),Ip2uint32(ipv4)
				}
				resultaddr = ipv4.String()
				ip32 = Ip2uint32(ipv4)
			}
		}
	}
	return resultaddr,ip32
}
