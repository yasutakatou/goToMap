/*
 * goToMap (Realized Virtual)
 *
 * @author    yasutakatou
 * @copyright 2020 yasutakatou
 * @license   BSD-2-Clause License
 */
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"strconv"

	"github.com/gorilla/websocket"
	"gopkg.in/ini.v1"
)

var (
	Debug                bool
	games                gameData
	actions              []actionData
	plays                []playersData
	players              int
	actClient            []int
	ActiveClients        = make(map[ClientConn]int)
	ActiveClientsRWMutex sync.RWMutex
	BUFFERLIMIT          int = 1
)

type gameData struct {
	PLAYER []string
	GOAL   []string
	RESULT []string
}

type playersData struct {
	Name   string
	IP     string
	Avater string
	PosX   float64
	PosY   float64
	Angle  float64
}

type actionData struct {
	ADDRESS string
	DATA    string
}

type responseData struct {
	Command string `json:"Command"`
	Data    string `json:"Data"`
}

type ClientConn struct {
	websocket *websocket.Conn
	clientIP  net.Addr
}

func addClient(cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	ActiveClients[cc] = 0
	ActiveClientsRWMutex.Unlock()
}

func deleteClient(cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	delete(ActiveClients, cc)
	ActiveClientsRWMutex.Unlock()
}

func changeClient(cc ClientConn, ccc ClientConn) {
	ActiveClientsRWMutex.Lock()
	delete(ActiveClients, cc)
	ActiveClients[ccc] = 0
	ActiveClientsRWMutex.Unlock()
}

func main() {
	players = 0

	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_https := flag.Bool("https", false, "[-https=https mode (yes or no. yes is enable)]")
	_cert := flag.String("cert", "localhost.pem", "[-cert=ssl_certificate file path (if you don't use https, haven't to use this option)]")
	_key := flag.String("key", "localhost-key.pem", "[-key=ssl_certificate_key file path (if you don't use https, haven't to use this option)]")
	_port := flag.String("port", "8080", "[-port=port number]")
	_config := flag.String("config", "config", "[-config=config file name]")

	flag.Parse()

	Debug = bool(*_Debug)

	loadConfig(string(*_config))

	http.HandleFunc("/ws", wsHandler)

	if *_https == false {
		go func() {
			err := http.ListenAndServe(":"+string(*_port), nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	} else {
		go func() {
			err := http.ListenAndServeTLS(":"+string(*_port), string(*_cert), string(*_key), nil)
			if err != nil {
				log.Fatal("ListenAndServeTLS: ", err)
			}
		}()
	}

	for {
		_, ip, err := getIFandIP()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("source ip: ", ip, " port: ", string(*_port)+" Exit: Ctrl+c")
		}
		time.Sleep(time.Second * 3)
	}

	os.Exit(0)
}

// FYI: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getIFandIP() (string, string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return iface.Name, ip.String(), nil
		}
	}
	return "", "", errors.New("are you connected to the network?")
}

func loadConfig(filename string) {
	loadOptions := ini.LoadOptions{}
	loadOptions.UnparseableSections = []string{"PLAYER", "GOAL", "RESULT", "ACTION"}

	cfg, err := ini.LoadSources(loadOptions, filename)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	setSingleConfigStrs(&games.PLAYER, "PLAYER", cfg.Section("PLAYER").Body())
	setSingleConfigStrs(&games.GOAL, "GOAL", cfg.Section("GOAL").Body())
	setSingleConfigStrs(&games.RESULT, "RESULT", cfg.Section("RESULT").Body())
	setActionConfig("ACTION", cfg.Section("ACTION").Body())
}

func setActionConfig(configType, datas string) {
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			strs := strings.Split(v, "\t")
			actions = append(actions, actionData{ADDRESS: strs[0], DATA: strs[1]})
		}
		if Debug == true {
			fmt.Println(v)
		}
	}
}

func setSingleConfigStrs(config *[]string, configType, datas string) {
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			*config = append(*config, v)
		}
		if Debug == true {
			fmt.Println(v)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if Debug == true {
		fmt.Printf("websocket: %s | %s %s\n", r.RemoteAddr, r.Method, r.URL)
	}
	if websocket.IsWebSocketUpgrade(r) {
		serveWebSocket(w, r)
	}
}

func goalCheck(strs string) int {
	for posInt, posStr := range games.GOAL {
		if strings.Index(strs, posStr) != -1 {
			return (posInt + 1)
		}
	}
	return 0
}

func actCheck(strs string) int {
	for i := 0; i < len(actions); i++ {
		if strings.Index(strs, actions[i].ADDRESS) != -1 {
			return (i + 1)
		}
	}
	return 0
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  (BUFFERLIMIT * 1024 * 1024),
	WriteBufferSize: (BUFFERLIMIT * 1024 * 1024),
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func winOrLose(winIp string) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()

	for client, _ := range ActiveClients {
		cIp := fmt.Sprintf("%s", client.clientIP)
		if Debug == true {
			fmt.Println("winnerIP: " + winIp + " clientIp: " + cIp)
		}
		if winIp == cIp {
			err := client.websocket.WriteJSON(responseData{Command: "message", Data: "Win"})
			if err != nil {
				log.Println(err)
			}
		} else {
			err := client.websocket.WriteJSON(responseData{Command: "message", Data: "Lose"})
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func sendAct(winIp string, act int) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()

	i := 0
	for client, _ := range ActiveClients {
		cIp := fmt.Sprintf("%s", client.clientIP)
		if winIp == cIp && actClient[i] != act {
			if err := client.websocket.WriteJSON(responseData{Command: "message", Data: actions[(act - 1)].DATA}); err != nil {
				fmt.Println(err)
			}
			actClient[i] = act
		}
		i = i + 1
	}
}

func sendCast(winIp, mess string) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()

	for client, _ := range ActiveClients {
		cIp := fmt.Sprintf("%s", client.clientIP)
		if winIp != cIp {
			if err := client.websocket.WriteJSON(responseData{Command: "message", Data: mess}); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func sendTo(winIp, name, mess string) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()

	strs := nameToIP(name)
	fmt.Println(strs)

	for client, _ := range ActiveClients {
		cIp := fmt.Sprintf("%s", client.clientIP)
		fmt.Println(cIp)
		if winIp != cIp && strs == cIp {
			if err := client.websocket.WriteJSON(responseData{Command: "message", Data: IPToName(winIp) + ">\n" + mess}); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func nameCheck(strs string) bool {
	for i := 0; i < len(plays); i++ {
		if plays[i].Name == strs {
			return false
		}
	}
	return true
}

func userlist(name string) string {
	strs := ""
	for i := 0; i < len(plays); i++ {
		if plays[i].Name != name {
			strs = strs + plays[i].Name + ","
		}
	}
	return strs
}

func nameToIP(strs string) string {
	for i := 0; i < len(plays); i++ {
		if strs == plays[i].Name {
			return plays[i].IP
		}
	}
	return ""
}

func IPToName(strs string) string {
	for i := 0; i < len(plays); i++ {
		if strs == plays[i].IP {
			return plays[i].Name
		}
	}
	return ""
}

func delPlayersArray(strs string) {
	for i := 0; i < len(plays); i++ {
		if strs == plays[i].Name {
			plays = unset(plays, i)
		}
	}
}

func unset(s []playersData, i int) []playersData {
	if i >= len(s) {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

func logoutClient(targ string) {
	strs := nameToIP(targ)
	for client, _ := range ActiveClients {
		if IPtoString(client) == strs {
			deleteClient(client)
			err := client.websocket.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func IPtoString(cc ClientConn) string {
	string := fmt.Sprintf("%s", cc.clientIP)
	return string
}

func serveWebSocket(wr http.ResponseWriter, req *http.Request) {
	var endFlag = false

	conn, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		fmt.Printf("websocket error: %s | %s\n", req.RemoteAddr, err)
		return
	}
	defer conn.Close()

	fmt.Printf("\n%s | join client!\n", req.RemoteAddr)

	client := conn.RemoteAddr()
	sockCli := ClientConn{conn, client}
	actClient = append(actClient, 0)
	addClient(sockCli)

	for {
		m := responseData{}

		err := conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error reading json.", err)
		}

		if Debug == true {
			fmt.Printf("Got message: %#v\n", m)
		}

		switch m.Command {
		case "logout":
			logoutClient(m.Data)
			if players > 0 {
				players = players - 1
			}
		case "users":
			strs := userlist(m.Data)
			if len(strs) > 4 {
				if err = conn.WriteJSON(responseData{Command: "list", Data: userlist(m.Data)}); err != nil {
					fmt.Println(err)
				}
			}
		case "to":
			strs := strings.Split(m.Data, ":")
			sendTo(req.RemoteAddr, strs[0], strs[1])
		case "cast":
			sendCast(req.RemoteAddr, m.Data)
		case "crumb":
			strs := strings.Split(m.Data, ":")
			actions = append(actions, actionData{ADDRESS: strs[0], DATA: strs[1]})
		case "start":
			strs := strings.Split(m.Data, ";")
			if nameCheck(strs[0]) == true {
				plays = append(plays, playersData{Name: strs[0], IP: req.RemoteAddr, Avater: strs[1], PosX: 0, PosY: 0, Angle: 0})
				if len(games.PLAYER) < players {
					if err = conn.WriteJSON(responseData{Command: "error", Data: "定員オーバーです"}); err != nil {
						fmt.Println(err)
					}
					deleteClient(sockCli)
				} else {
					if err = conn.WriteJSON(responseData{Command: "goto", Data: games.PLAYER[players]}); err != nil {
						fmt.Println(err)
					}
					players = players + 1
				}
			} else {
				if err = conn.WriteJSON(responseData{Command: "error", Data: "その名前もう居ますよ"}); err != nil {
					fmt.Println(err)
				}
				deleteClient(sockCli)
			}
		case "move":
			checks := goalCheck(m.Data)
			if checks > 0 && endFlag == false {
				endFlag = true
				if len(games.RESULT) >= checks {
					checks = checks - 1
					if err = conn.WriteJSON(responseData{Command: "message", Data: games.RESULT[checks]}); err != nil {
						fmt.Println(err)
					}
					time.Sleep(time.Second * 3)
					winOrLose(req.RemoteAddr)
				} else {
					if err = conn.WriteJSON(responseData{Command: "message", Data: games.RESULT[0]}); err != nil {
						fmt.Println(err)
					}
					time.Sleep(time.Second * 3)
					winOrLose(req.RemoteAddr)
				}
			} else {
				if len(m.Data) > 0 {
					act := actCheck(m.Data)
					if act > 0 && endFlag == false {
						sendAct(req.RemoteAddr, act)
					}
					updateStat(req.RemoteAddr,m.Data)	
				}
			}
		}
	}
}

func updateStat(cIp, strs string) {
	for i := 0; i < len(plays); i++ {
		if cIp == plays[i].IP {
			//https://www.google.co.jp/maps/@35.5773926,139.6606327,3a,75y,44.37h,89.35t/data=!3m6!1e1!3m4!1sdzcafrQ_B8ZOJTChCd3A6Q!2e0!7i16384!8i8192?hl=ja
			stra := strings.Split(strs, "/")
			strb := strings.Split(stra[4], ",")	
			strc := strings.Replace(strb[0], "@", "", -1)
			strd := strings.Replace(strb[4], "h", "", -1)

			fX, err := strconv.ParseFloat(strc, 64)
			if err == nil {
				plays[i].PosX = fX
			}

			fY, err := strconv.ParseFloat(strb[1], 64)
			if err == nil {
				plays[i].PosY = fY
			}

			fA, err := strconv.ParseFloat(strd, 64)
			if err == nil {
				plays[i].Angle = fA
			}
			fmt.Println(plays[i])
		}
	}
}

func disAvater(cIp, strs string) (string, string) {}
	for i := 0; i < len(plays); i++ {
		if cIp == plays[i].IP {
			//https://www.google.co.jp/maps/@35.5773926,139.6606327,3a,75y,44.37h,89.35t/data=!3m6!1e1!3m4!1sdzcafrQ_B8ZOJTChCd3A6Q!2e0!7i16384!8i8192?hl=ja
			stra := strings.Split(strs, "/")
			strb := strings.Split(stra[4], ",")	
			strc := strings.Replace(strb[0], "@", "", -1)
			strd := strings.Replace(strb[4], "h", "", -1)

			fX, err := strconv.ParseFloat(strc, 64)
			if err == nil {
				plays[i].PosX = fX
			}

			fY, err := strconv.ParseFloat(strb[1], 64)
			if err == nil {
				plays[i].PosY = fY
			}

			fA, err := strconv.ParseFloat(strd, 64)
			if err == nil {
				plays[i].Angle = fA
			}
			fmt.Println(plays[i])
		}
	}
}