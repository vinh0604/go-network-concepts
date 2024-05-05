package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vinh0604/go-network-concepts/internal/netfunc"
)

type NetworkInterface struct {
	Netmask       string `json:"netmask"`
	InterfaceName string `json:"interface"`
	Ad            int    `json:"ad"`
}

type RouterInfo struct {
	Connections map[string]NetworkInterface `json:"connections"`
	Netmask     string                      `json:"netmask"`
	IfCount     int                         `json:"if_count"`
	IfPrefix    string                      `json:"if_prefix"`
}

type NetworkInfo struct {
	Routers map[string]RouterInfo `json:"routers"`
	SrcDest [][]string            `json:"src-dest"`
}

func main() {
	networkInfoData, err := os.ReadFile("./data/dijkstra/example1.json")
	if err != nil {
		panic(err)
	}

	networkInfo := NetworkInfo{}
	json.Unmarshal(networkInfoData, &networkInfo)

	for i := range networkInfo.SrcDest {
		fmt.Println("Src:", networkInfo.SrcDest[i][0], "Dest:", networkInfo.SrcDest[i][1])

		shortestPath, err := dijkstrasShortestPath(networkInfo.Routers, networkInfo.SrcDest[i][0], networkInfo.SrcDest[i][1])
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Shortest path:", shortestPath)
		}
	}
}

const MAX_INT = int(^uint(0) >> 1)

func dijkstrasShortestPath(routers map[string]RouterInfo, srcIp string, destIp string) ([]string, error) {
	routerInfos := make(map[string]netfunc.RouterInfo)
	for router := range routers {
		netmask := routers[router].Netmask
		notation := strings.Split(netmask, "/")[1]
		notationInt, err := strconv.Atoi(notation)
		if err != nil {
			return nil, err
		}
		routerInfos[router] = netfunc.RouterInfo{
			NetmaskNotation: uint8(notationInt),
		}
	}

	srcRouter, err := netfunc.RouterForIp(routerInfos, srcIp)
	if err != nil {
		return nil, err
	}
	destRouter, err := netfunc.RouterForIp(routerInfos, destIp)
	if err != nil {
		return nil, err
	}

	// Dijkstra's algorithm
	toVisit := make(map[string]RouterInfo)
	distances := make(map[string]int)
	parents := make(map[string]*string)
	for router := range routers {
		distances[router] = MAX_INT
		toVisit[router] = routers[router]
		parents[router] = nil
	}

	distances[srcRouter] = 0
	for len(toVisit) > 0 {
		minDistance := MAX_INT
		current := ""
		for node := range toVisit {
			if distances[node] <= minDistance {
				minDistance = distances[node]
				current = node
			}
		}
		delete(toVisit, current)

		if distances[current] == MAX_INT {
			break
		}

		neighbors := routers[current].Connections
		for neighbor := range neighbors {
			neighborDistance := distances[current] + neighbors[neighbor].Ad
			if neighborDistance < distances[neighbor] {
				distances[neighbor] = neighborDistance
				parents[neighbor] = &current
			}
		}
	}

	if parents[destRouter] == nil {
		return []string{}, fmt.Errorf("no path found from %s to %s", srcIp, destIp)
	}

	path := []string{destRouter}
	current := destRouter
	for parents[current] != nil {
		current = *parents[current]
		path = append([]string{current}, path...)
	}
	return path, nil
}
