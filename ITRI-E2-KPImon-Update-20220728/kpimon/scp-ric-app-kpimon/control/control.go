package control

import (
	"encoding/json"
	"errors"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/go-redis/redis"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"gerrit.o-ran-sc.org/r/ric-plt/sdlgo" // added by sww, ITRI - sdl
	"fmt" // added by sww, ITRI - log
//	"net" // added by sww, ITRI
	"net/http" // added by sww, ITRI
)

// added by sww, ITRI
type GlobalNbIdType struct {
	PlmnId   string `json:"plmnId"`
	NbId   string `json:"nbId"`
	CuUpId   string `json:"cuUpId"`
	DuId   string `json:"duId"`
}

type RanIdentity struct {
	InventoryName string     `json:"inventoryName"`
	GlobalNbId GlobalNbIdType `json:"globalNbId"`
	ConnectionStatus string     `json:"connectionStatus"`
}

type RanData struct {
//	cellId string
	cellIDs []string
	indFlag int
}

const (
    NODE_GNB = 1
    NODE_DU = 2
    NODE_CUCP = 3
    NODE_CUUP = 4
    NODE_CU = 5
)

const NODE_FLAG_CUCP = 0x01
const NODE_FLAG_DU = 0x02
const NODE_FLAG_CUUP = 0x04

type Control struct {
	ranList []string //nodeB list
	eventCreateExpired int32 //maximum time for the RIC Subscription Request event creation procedure in the E2 Node
	eventDeleteExpired int32 //maximum time for the RIC Subscription Request event deletion procedure in the E2 Node
	rcChan                chan *xapp.RMRParams //channel for receiving rmr message
	client                *redis.Client        //redis client
	eventCreateExpiredMap map[string]bool      //map for recording the RIC Subscription Request event creation procedure is expired or not
	eventDeleteExpiredMap map[string]bool      //map for recording the RIC Subscription Request event deletion procedure is expired or not
	eventCreateExpiredMu  *sync.Mutex          //mutex for eventCreateExpiredMap
	eventDeleteExpiredMu  *sync.Mutex          //mutex for eventDeleteExpiredMap
	subCreatedMap map[string]int32             //map for recording the RIC Subscription Request creation  // updated by sww, ITRI
	subCreatedMu  *sync.Mutex                  //mutex for subCreatedMap
	subIdMap map[string]int                    //map for recording the RIC Subscription Request ID  // updated by sww, ITRI
	indTimeMap map[string]time.Time            //map for recording the RIC Indication Timestamp  // updated by sww, ITRI
//	cidMap map[string]string                   //map for ran name of cid
	ranMap map[string]*RanData                  //map for ran name of cid
	sdlAccessMu  *sync.Mutex                   //mutex for sdl Access
	ranFuncID int
	indCount uint64
	running bool
}

const(
	SUB_PENDING = iota
	SUB_CREATED
	DISCONNECTED
)

var sdlUE *sdlgo.SdlInstance

var sdlCELL *sdlgo.SdlInstance

var myClient = &http.Client{Timeout: 15 * time.Second}

func writeRanName(cellIDHdr string, ranName string) {

	var cellMetrics CellMetricsEntry

	retMap, err := sdlCELL.Get([]string{cellIDHdr})

	if err != nil {
		panic(err)
	} else {
		fmt.Println("\nwriteRanName: cell: sdl get")
		fmt.Println(retMap)
		if retMap[cellIDHdr] != nil {
			cellJsonStr := retMap[cellIDHdr].(string)
			json.Unmarshal([]byte(cellJsonStr), &cellMetrics)
			fmt.Println(cellJsonStr)
			fmt.Println(cellMetrics)
		} else {
			fmt.Println("\n cell: not Exists")
			cellMetrics = CellMetricsEntry{}
		}
	}


	cellMetrics.RANName = ranName

	newCellJsonStr, err := json.Marshal(cellMetrics)
	if err != nil {
		fmt.Printf("\nFailed to marshal CellMetrics with CellID [%s]: %v", cellIDHdr, err)
	}
	err = sdlCELL.Set(cellIDHdr, newCellJsonStr)
	if err != nil {
		fmt.Printf("\nFailed to set CellMetrics into redis with CellID [%s]: %v", cellIDHdr, err)
	}
}

func parseRanNode(plmnId string, nbId string) (CellID string, err error) {
	i, err := strconv.ParseUint(nbId, 2, 64)
	if err != nil {
		return "", err
	}
//	CellID, err = strconv.FormatUint(i, 16, 64)
	nodeId := fmt.Sprintf("%09x", i)

	CellID = plmnId + nodeId

	fmt.Printf("SWW parseRanNode CellID=")
	fmt.Println(CellID)
	return CellID, nil
}


func httpGetRanList(e2murl string) (*[]RanIdentity, error) {
	fmt.Println("\nInvoked httprestful.httpGetE2TList: " + e2murl)
	r, err := myClient.Get(e2murl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode == 200 {
//		fmt.Printf("http client raw response: %v\n", r)
		var ranNodes []RanIdentity
		err = json.NewDecoder(r.Body).Decode(&ranNodes)
		if err != nil {
			fmt.Println("Json decode failed: " + err.Error())
		}
		fmt.Printf("httprestful.httpGetXApps returns: %v\n", ranNodes)
		return &ranNodes, err
	}
	fmt.Printf("httprestful got an unexpected http status code: %v\n", r.StatusCode)
	return nil, nil
}

func contains(arr []string, str string) bool {
	for _, item := range arr {
		if str == item {
			return true
		}
	}
	return false
}

func init() {
	file := "/opt/kpimon.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetPrefix("[qSkipTool]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	logLevel := xapp.Config.GetInt("controls.logLevel")
	fmt.Printf("controls controls.logLevel=%d\n", logLevel)
	xapp.Logger.SetLevel(logLevel)

	// added by sww, ITRI - init sdl
	sdlUE = sdlgo.NewSdlInstance("TS-UE-metrics", sdlgo.NewDatabase()) // for UE
	sdlCELL = sdlgo.NewSdlInstance("TS-cell-metrics", sdlgo.NewDatabase()) // for CELL
}

func NewControl() Control {
	funcID := xapp.Config.GetInt("controls.ranFunctionID")
	fmt.Printf("controls controls.ranFunctionID=%d\n", funcID)
	if funcID == 0 {
		funcID = 11
	}

	return Control{nil,//ranNames,
		5, 5,
		make(chan *xapp.RMRParams),
		redis.NewClient(&redis.Options{
			Addr:     os.Getenv("redisAddr"), //"localhost:6379"
			Password: "",
			DB:       0,
		}),
		make(map[string]bool),
		make(map[string]bool),
		&sync.Mutex{},
		&sync.Mutex{},
		make(map[string]int32), // updated by sww, ITRI
		&sync.Mutex{},
		make(map[string]int), // updated by sww, ITRI
		make(map[string]time.Time), // updated by sww, ITRI
//		make(map[string]string), // updated by sww, ITRI
		make(map[string]*RanData), // updated by sww, ITRI
		&sync.Mutex{},
		funcID, // updated by sww, ITRI
		0, // updated by sww, ITRI
		true} // updated by sww, ITRI
}

func ReadyCB(i interface{}) {
	c := i.(*Control)

	c.startTimerSubReq()
	go c.controlLoop()
}

func (c *Control) Run() {
	_, err := c.client.Ping().Result()
	if err != nil {
		xapp.Logger.Error("Failed to connect to Redis DB with %v", err)
		fmt.Printf("\nFailed to connect to Redis DB with %v", err)
	}
//	if len(c.ranList) > 0 {
		xapp.SetReadyCB(ReadyCB, c)
		xapp.Run(c)
/*	} else {
		xapp.Logger.Error("gNodeB not set for subscription")
		fmt.Printf("\ngNodeB not set for subscription")
	}
*/
}

func (c *Control) Stop() {
	c.running = false;
	for _, ran := range c.ranList {
		fmt.Printf("Stop: ran=%s\n", ran)
                if c.subIdMap[ran] != -1 {
			c.sendRicSubDelRequestForRan(ran, 123, c.subIdMap[ran], c.ranFuncID)
		}
	}
}

func (c *Control) clearCellEntry(ranName string) {
	var cellMetrics CellMetricsEntry

	retMap, err := sdlCELL.Get(c.ranMap[ranName].cellIDs)

	if err != nil {
		panic(err)
		return
	}

	xapp.Logger.Debug("clearMetrics: CELL: sdl get %v", c.ranMap[ranName].cellIDs)
	fmt.Println(retMap)

	for _, cellID := range c.ranMap[ranName].cellIDs {
		if retMap[cellID] == nil {
			fmt.Println("CELL: not Exists")
			continue
		}

		cellJsonStr := retMap[cellID].(string)
		json.Unmarshal([]byte(cellJsonStr), &cellMetrics)
		fmt.Println(cellMetrics)

		if (c.ranMap[ranName].indFlag & NODE_FLAG_CUCP) > 0 {
			cellMetrics.RANName = ""
		}
		if (c.ranMap[ranName].indFlag & NODE_FLAG_CUUP) > 0 {
			cellMetrics.PDCPBytesDL = 0
			cellMetrics.PDCPBytesUL = 0
		}
		if (c.ranMap[ranName].indFlag & NODE_FLAG_DU) > 0 {
			cellMetrics.AvailPRBDL = 0
			cellMetrics.AvailPRBUL = 0
		}

		newCellJsonStr, err := json.Marshal(cellMetrics)
		if err != nil {
			xapp.Logger.Error("Failed to marshal CellMetrics with CELL ID [%s]: %v", cellID, err)
			return
		}

		xapp.Logger.Debug("clearMetrics: CELL: sdl set")
		fmt.Println(cellMetrics)
		err = sdlCELL.Set(cellID, newCellJsonStr)

		if err != nil {
			xapp.Logger.Error("Failed to set CellMetrics into redis with CELL ID [%s]: %v", cellID, err)
			return
		}
	}
}

func (c *Control) clearUeEntry(ranName string, ueID string) {
	var ueMetrics UeMetricsEntry

	ueMap, err := sdlUE.Get([]string{ueID})

	if err != nil {
		panic(err)
		return
	}

	xapp.Logger.Debug("clearMetrics: nodeType=%d UE: sdl get", c.ranMap[ranName].indFlag)
//	fmt.Println(ueMap)
	if ueMap[ueID] == nil {
		fmt.Println("UE: not Exists")
		return
	}

	ueJsonStr := ueMap[ueID].(string)
	json.Unmarshal([]byte(ueJsonStr), &ueMetrics)

	if !contains(c.ranMap[ranName].cellIDs, ueMetrics.ServingCellID) {
		// not match
		return;
	}

	fmt.Println(ueMetrics)

	if (c.ranMap[ranName].indFlag & NODE_FLAG_CUUP) > 0 {
		ueMetrics.PDCPBytesDL = 0
		ueMetrics.PDCPBytesUL = 0
		ueMetrics.MeasTimestampPDCPBytes.TVsec = 0
		ueMetrics.MeasTimestampPDCPBytes.TVnsec = 0
	}

	if (c.ranMap[ranName].indFlag & NODE_FLAG_DU) > 0 {
		ueMetrics.PRBUsageDL = 0
		ueMetrics.PRBUsageUL = 0
		ueMetrics.MeasTimestampPRB.TVsec = 0
		ueMetrics.MeasTimestampPRB.TVnsec = 0
	}

	if (c.ranMap[ranName].indFlag & NODE_FLAG_CUCP) > 0 {
		ueMetrics.ServingCellRF.RSRP = 0
		ueMetrics.ServingCellRF.RSRQ = 0
		ueMetrics.ServingCellRF.RSSINR = 0
		ueMetrics.NeighborCellsRF = nil
		ueMetrics.MeasTimeRF.TVsec = 0
		ueMetrics.MeasTimeRF.TVnsec = 0
		ueMetrics.BWPData.BWPID = 0
		ueMetrics.BWPData.LocationAndBandwidth = 0
	}

	if ueMetrics.MeasTimestampPDCPBytes.TVsec == 0 && ueMetrics.MeasTimestampPRB.TVsec == 0 && ueMetrics.MeasTimeRF.TVsec == 0 {
		ueMetrics.ServingCellID = ""
	}

	newUeJsonStr, err := json.Marshal(ueMetrics)
	if err != nil {
		xapp.Logger.Error("Failed to marshal UeMetrics with UE ID [%s]: %v", ueID, err)
		return
	}
	fmt.Printf("to set UeMetrics - UE ID [%s]: PDCPBytes=%d %d PRBUsage=%d %d\n", ueMetrics.UeID, ueMetrics.PDCPBytesDL, ueMetrics.PDCPBytesUL, ueMetrics.PRBUsageDL, ueMetrics.PRBUsageUL)

	xapp.Logger.Debug("clearUeEntry: UE: sdl set")
	fmt.Println(ueMetrics)
	err = sdlUE.Set(ueID, newUeJsonStr)

	if err != nil {
		xapp.Logger.Error("Failed to set UeMetrics into redis with UE ID [%s]: %v", ueID, err)
		return
	}
}

func (c *Control) clearMetrics(ranName string) {

	if _, ok := c.ranMap[ranName]; !ok {
		fmt.Println("clearMetrics: ranMap not found")
		return
	}

	if c.ranMap[ranName].indFlag == 0 || c.ranMap[ranName].cellIDs == nil {
		fmt.Printf("clearMetrics: no entry to delete: %s %v", ranName, c.ranMap[ranName])
		return
	}

	// clear cell data
	c.clearCellEntry(ranName)

	// clear ue data
	keys, err := sdlUE.GetAll()
	if err != nil {
		panic(err)
		return
	}

	for _, ueID := range keys {
		c.sdlAccessMu.Lock()
		c.clearUeEntry(ranName, ueID)
		c.sdlAccessMu.Unlock()
	}

	c.clearRanMap(ranName)
}
/*
func (c *Control) updateRanMap(ranName string, cellID string, flag int) {
	var ran RanData
	if _, ok := c.ranMap[ranName]; ok {
		fmt.Printf(" updateRanMap: exist: %s: %s, %d | %d\n", ranName, cellID, flag, c.ranMap[ranName].indFlag)
		ran.indFlag = c.ranMap[ranName].indFlag | flag
		if !contains(c.ranMap[ranName].cellIDs, cellID) {
			ran.cellIDs = append(c.ranMap[ranName].cellIDs, cellID)
		} else {
			ran.cellIDs = c.ranMap[ranName].cellIDs
		}
	} else {
		fmt.Printf(" updateRanMap: new: %s: %s, %d\n", ranName, cellID, flag)
		ran.indFlag = flag
		ran.cellIDs = append(c.ranMap[ranName].cellIDs, cellID)
	}
	c.ranMap[ranName] = ran
	fmt.Println(c.ranMap[ranName])
}

func (c *Control) clearRanMap(ranName string) {
	if _, ok := c.ranMap[ranName]; ok {
		c.ranMap[ranName] = RanData{ c.ranMap[ranName].cellIDs, 0, }
	}
}*/

func (c *Control) findRanNode(ranName string) (bool) {
	for _, ran := range c.ranList {
		if ran == ranName {
			return true
		}
	}

	return false
}

func (c *Control) updateRanList() (bool) {
//	fmt.Println("\nupdateRanList: httpGetRanList...")
//	ranNodes, err := httpGetRanList("http://service-ricplt-e2mgr-http.ricplt:3800/v1/nodeb/ids")
	ranNodes, err := httpGetRanList("http://service-ricplt-e2mgr-http.ricplt:3800/v1/nodeb/states")
	if err != nil {
		fmt.Println(err)
		return false
	}

	if ranNodes == nil || len(*ranNodes) <= 0 {
		fmt.Println("no ranNode")
		return false
	}

//	var cellId string

	for _, ran := range *ranNodes {
		if ran.ConnectionStatus == "CONNECTED" {

			fmt.Printf("updateRanList: ranNode %s cuupid: [%s] duid: [%s]\n", ran.InventoryName, ran.GlobalNbId.CuUpId, ran.GlobalNbId.DuId)
			if !c.findRanNode(ran.InventoryName) {
				fmt.Printf("updateRanList: new ranNode %s\n", ran.InventoryName)
/*				if ran.GlobalNbId.CuUpId == "" && ran.GlobalNbId.DuId == "" {
					cellId, _ = parseRanNode(ran.GlobalNbId.PlmnId, ran.GlobalNbId.NbId)
					writeRanName(cellId, ran.InventoryName)
					fmt.Printf("updateRanList: %s\n", ran.InventoryName)
				} else {
					fmt.Println(" updateRanList: not to write...")
                                }*/
				c.ranList = append(c.ranList, ran.InventoryName)
				c.subCreatedMap[ran.InventoryName] = SUB_PENDING
				c.subIdMap[ran.InventoryName] = -1
			} else if c.subCreatedMap[ran.InventoryName] == DISCONNECTED {
				fmt.Printf("updateRanList: %s -> SUB_PENDING\n", ran.InventoryName)
/*				if ran.GlobalNbId.CuUpId == "" && ran.GlobalNbId.DuId == "" {
					cellId, _ = parseRanNode(ran.GlobalNbId.PlmnId, ran.GlobalNbId.NbId)
					writeRanName(cellId, ran.InventoryName)
					fmt.Printf("updateRanList: %s\n", ran.InventoryName)
				} else {
					fmt.Println(" updateRanList: not to write...")
                                }*/
				c.subCreatedMap[ran.InventoryName] = SUB_PENDING
				c.subIdMap[ran.InventoryName] = -1
			} else if c.subCreatedMap[ran.InventoryName] == SUB_CREATED {
/*				fmt.Printf("updateRanList: %s check timestamp\n", ran.InventoryName)
				cur := time.Now()
				duration := cur.Sub(c.indTimeMap[ran.InventoryName])
		                if duration > 10 {
					fmt.Printf("updateRanList: %s ind timeout\n", ran.InventoryName)
			                if c.subIdMap[ran.InventoryName] != -1 {
						c.sendRicSubDelRequestForRan(ran.InventoryName, 123, c.subIdMap[ran.InventoryName], c.ranFuncID)
						c.subIdMap[ran.InventoryName] = -1
					}
					c.subCreatedMap[ran.InventoryName] = SUB_PENDING
				}*/
			}
		} else {
			fmt.Printf("updateRanList: %s %s\n", ran.InventoryName, ran.ConnectionStatus)
			if c.findRanNode(ran.InventoryName) && c.subCreatedMap[ran.InventoryName] != DISCONNECTED {
				fmt.Printf("updateRanList: %s -> DISCONNECTED subId=%d\n", ran.InventoryName, c.subIdMap[ran.InventoryName])
				c.clearMetrics(ran.InventoryName)
/*				if ran.GlobalNbId.CuUpId == "" && ran.GlobalNbId.DuId == "" {
					cellId, _ = parseRanNode(ran.GlobalNbId.PlmnId, ran.GlobalNbId.NbId)
					writeRanName(cellId, "")
					c.clearRanName(ran.InventoryName)
				} else {
					fmt.Println(" updateRanList: not to write...")
                                }*/
				c.subCreatedMap[ran.InventoryName] = DISCONNECTED
		                if c.subIdMap[ran.InventoryName] != -1 {
					c.sendRicSubDelRequestForRan(ran.InventoryName, 123, c.subIdMap[ran.InventoryName], c.ranFuncID)
					c.subIdMap[ran.InventoryName] = -1
				}

			}
		}
	}
	return true
}

func (c *Control) startTimerSubReq() {
	timerSR := time.NewTimer(5 * time.Second)
	count := 0

	if (c.ranList != nil && len(c.ranList) > 0) {
		for _, ran := range c.ranList {
			c.subCreatedMap[ran] = SUB_PENDING
			c.subIdMap[ran] = -1
		}
	}

	go func(t *time.Timer) {
		defer timerSR.Stop()
		for c.running {
			<-t.C
			count++

			updated := c.updateRanList()
//			updated := true
			if updated {
				xapp.Logger.Debug("try to send RIC_SUB_REQ to gNodeB with cnt=%d", count)
				err := c.sendRicSubRequest(1001, 1001, c.ranFuncID) // modified ranFunID by sww ITRI

				if err != nil {
					fmt.Println(err)
				}
			}



			t.Reset(5 * time.Second)
/*			if err != nil && count < MAX_SUBSCRIPTION_ATTEMPTS {
				t.Reset(5 * time.Second)
			} else {
				break
			}*/
		}
	}(timerSR)
}

func (c *Control) Consume(rp *xapp.RMRParams) (err error) {
	c.rcChan <- rp
	return
}

func (c *Control) rmrSend(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, false) {
		err = errors.New("rmr.Send() failed")
		xapp.Logger.Error("Failed to rmrSend to %v", err)
	}
	return
}

func (c *Control) rmrReplyToSender(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, true) {
		err = errors.New("rmr.Send() failed")
		xapp.Logger.Error("Failed to rmrReplyToSender to %v", err)
	}
	return
}

func (c *Control) controlLoop() {
	for {
		msg := <-c.rcChan
		xapp.Logger.Debug("Received message type: %d", msg.Mtype)
		switch msg.Mtype {
		case 12050:
			c.handleIndication(msg)
		case 12011:
			c.handleSubscriptionResponse(msg)
		case 12012:
			c.handleSubscriptionFailure(msg)
		case 12021:
			c.handleSubscriptionDeleteResponse(msg)
		case 12022:
			c.handleSubscriptionDeleteFailure(msg)
		default:
			err := errors.New("Message Type " + strconv.Itoa(msg.Mtype) + " is discarded")
			xapp.Logger.Error("Unknown message type: %v", err)
		}
	}
}

func (c *Control) ParseNeighborCellRFList(ueData []string, count int) (neighborCellRFList []NeighborCellRFType, err error) {
//	fmt.Printf("SWW ParseNeighborCellRFList size=")
//	fmt.Println(size)

        neighborCellRFList = make([]NeighborCellRFType, count);

	cur := 10
	for index := 0; index < count; index++ {
		neighborCellRFList[index].CellID = ueData[cur]
		cur++
		rsrp, _ := strconv.ParseInt(ueData[cur], 10, 32)
		cur++
		rsrq, _ := strconv.ParseInt(ueData[cur], 10, 32)
		cur++
		rsSinr, _ := strconv.ParseInt(ueData[cur], 10, 32)
		cur++
		neighborCellRFList[index].CellRF.RSRP = int32(rsrp)
		neighborCellRFList[index].CellRF.RSRQ = int32(rsrq)
		neighborCellRFList[index].CellRF.RSSINR = int32(rsSinr)
	}

	return
}

func (c *Control) parseUeData(ueColumn string, nodeType int) (err error) {
	ueData := strings.Split(ueColumn, " ")
	if len(ueData) < 10 {
		xapp.Logger.Error("Failed to parse ts data - too few arguments: %d !!!", len(ueData))
		return
	}

	now := time.Now()
	ueID := ueData[0]
	servingCellID := ueData[1]
	neighborNum, _ := strconv.Atoi(ueData[9])

	var ueMetrics UeMetricsEntry

	c.sdlAccessMu.Lock()

	retMap, err := sdlUE.Get([]string{ueID})

	if err != nil {
		panic(err)
	} else {
		xapp.Logger.Debug("handleIndicationBWP: nodeType=%d UE: sdl get", nodeType)
		fmt.Println(retMap)
		if retMap[ueID] != nil {
			ueJsonStr := retMap[ueID].(string)
			json.Unmarshal([]byte(ueJsonStr), &ueMetrics)
			fmt.Println(ueMetrics)
		} else {
			ueMetrics = UeMetricsEntry{}
			ueMetrics.UeID = ueID // updated by sww, ITRI
			fmt.Println("oDUUE: not Exists")
		}
	}

	ueMetrics.ServingCellID = servingCellID

	if (nodeType & NODE_FLAG_CUUP) > 0 {
		dlPDCPBytes, _ := strconv.ParseInt(ueData[2], 10, 64)
		ulPDCPBytes, _ := strconv.ParseInt(ueData[3], 10, 64)
		ueMetrics.PDCPBytesDL = dlPDCPBytes
		ueMetrics.PDCPBytesUL = ulPDCPBytes
		ueMetrics.MeasTimestampPDCPBytes.TVsec = now.Unix()
		ueMetrics.MeasTimestampPDCPBytes.TVnsec = now.UnixNano()
	}

	if (nodeType & NODE_FLAG_DU) > 0 {
		dlPRBUsage, _ := strconv.ParseInt(ueData[4], 10, 64)
		ulPRBUsage, _ := strconv.ParseInt(ueData[5], 10, 64)
		ueMetrics.PRBUsageDL = dlPRBUsage
		ueMetrics.PRBUsageUL = ulPRBUsage
		ueMetrics.MeasTimestampPRB.TVsec = now.Unix()
		ueMetrics.MeasTimestampPRB.TVnsec = now.UnixNano()
	}

	if (nodeType & NODE_FLAG_CUCP) > 0 {
		rsrp, _ := strconv.ParseInt(ueData[6], 10, 32)
		rsrq, _ := strconv.ParseInt(ueData[7], 10, 32)
		rsSinr, _ := strconv.ParseInt(ueData[8], 10, 32)
		ueMetrics.ServingCellRF.RSRP = int32(rsrp)
		ueMetrics.ServingCellRF.RSRQ = int32(rsrq)
		ueMetrics.ServingCellRF.RSSINR = int32(rsSinr)
		ueMetrics.NeighborCellsRF, _ = c.ParseNeighborCellRFList(ueData, neighborNum)
		ueMetrics.MeasTimeRF.TVsec = now.Unix()
		ueMetrics.MeasTimeRF.TVnsec = now.UnixNano()
	}

	// check if more data for BWP
	cur := 10 + (neighborNum * 4)
	if (cur + 2) <= len(ueData) {
		bwpID, _ := strconv.ParseInt(ueData[cur], 10, 16)
		locationAndBandwidth, _ := strconv.ParseUint(ueData[cur+1], 10, 16)
		ueMetrics.BWPData.BWPID = uint16(bwpID)
		ueMetrics.BWPData.LocationAndBandwidth = uint16(locationAndBandwidth)
	}

	newUeJsonStr, err := json.Marshal(ueMetrics)
	if err != nil {
		xapp.Logger.Error("Failed to marshal UeMetrics with UE ID [%s]: %v", ueID, err)
	} else {
		fmt.Printf("to set UeMetrics - UE ID [%s]: PDCPBytes=%d %d PRBUsage=%d %d\n", ueID, ueMetrics.PDCPBytesDL, ueMetrics.PDCPBytesUL, ueMetrics.PRBUsageDL, ueMetrics.PRBUsageUL)

		xapp.Logger.Debug("handleIndicationBWP: UE: sdl set")
		fmt.Println(ueMetrics)
		err = sdlUE.Set(ueID, newUeJsonStr)
	}

	c.sdlAccessMu.Unlock()

	if err != nil {
		xapp.Logger.Error("Failed to set UeMetrics into redis with UE ID [%s]: %v", ueID, err)
		return
	}

	return nil
}

func (c *Control) parseCellData(cellColumn string, nodeType int, ranName string) (err error) {
	cellData := strings.Split(cellColumn, " ")
	if len(cellData) < 5 {
		xapp.Logger.Error("Failed to parse ts data - too few arguments: %d !!!", len(cellData))
		return
	}

	cellID := cellData[0]

	var cellMetrics CellMetricsEntry

	retMap, err := sdlCELL.Get([]string{cellID})

	if err != nil {
		panic(err)
	} else {
		xapp.Logger.Debug("handleIndicationBWP: nodeType=%d CELL: sdl get", nodeType)
		fmt.Println(retMap)
		if retMap[cellID] != nil {
			cellJsonStr := retMap[cellID].(string)
			json.Unmarshal([]byte(cellJsonStr), &cellMetrics)
			fmt.Println(cellMetrics)
		} else {
			cellMetrics = CellMetricsEntry{}
			fmt.Println("CELL: not Exists")
		}
	}

        c.updateRanMap(ranName, cellID, nodeType)

	if (nodeType & NODE_FLAG_CUCP) > 0 {
//                c.cidMap[ranName] = cellID
		cellMetrics.RANName = ranName
	}
	if (nodeType & NODE_FLAG_CUUP) > 0 {
		dlPDCPBytes, _  := strconv.ParseInt(cellData[1], 10, 64)
		ulPDCPBytes, _  := strconv.ParseInt(cellData[2], 10, 64)
		cellMetrics.PDCPBytesDL = dlPDCPBytes
		cellMetrics.PDCPBytesUL = ulPDCPBytes
	}

	if (nodeType & NODE_FLAG_DU) > 0 {
		dlAvailPRB , _  := strconv.ParseInt(cellData[3], 10, 64)
		ulAvailPRB, _  := strconv.ParseInt(cellData[4], 10, 64)
		cellMetrics.AvailPRBDL = dlAvailPRB
		cellMetrics.AvailPRBUL = ulAvailPRB
	}

	newCellJsonStr, err := json.Marshal(cellMetrics)
	if err != nil {
		xapp.Logger.Error("Failed to marshal CellMetrics with CELL ID [%s]: %v", cellID, err)
		return
	}

	xapp.Logger.Debug("handleIndicationBWP: CELL: sdl set")
	fmt.Println(cellMetrics)
	err = sdlCELL.Set(cellID, newCellJsonStr)

	if err != nil {
		xapp.Logger.Error("Failed to set CellMetrics into redis with CELL ID [%s]: %v", cellID, err)
		return
	}

	return nil
}

func (c *Control) handleIndicationBWP(indicationMsg *DecodedIndicationMessage, ranName string) (err error) {
	xapp.Logger.Debug("handleIndicationBWP: after GetIndicationMessage")

	indMsg := string(indicationMsg.IndMessage)
	tsData := strings.Split(indMsg, "\n")
	fmt.Println("tsData:")
	if len(indMsg) < 2 {
		xapp.Logger.Error("Failed to parse ts data - too few arguments!!!")
		return
	}
	fmt.Println(tsData[0])
	fmt.Println(tsData[1])
	nodeType, err := strconv.Atoi(tsData[0])
	if err != nil {
		xapp.Logger.Error("Failed to parse ts data - %v", err)
		return
	}
	// conver node type
	var nodeFlag int
	if nodeType == NODE_CUCP {
		nodeFlag = NODE_FLAG_CUCP
	} else if nodeType == NODE_CUUP {
		nodeFlag = NODE_FLAG_CUUP
	} else if nodeType == NODE_DU {
		nodeFlag = NODE_FLAG_DU
	} else if nodeType == NODE_CU {
		nodeFlag = NODE_CUCP | NODE_CUUP
	} else {
		nodeFlag = NODE_CUCP | NODE_CUUP | NODE_FLAG_DU
	}

	c.parseCellData(tsData[1], nodeFlag, ranName)

	if len(indMsg) > 2 {
		// ue data
		fmt.Println(tsData[2])
		ueNum, err := strconv.Atoi(tsData[2])
		if err != nil {
			xapp.Logger.Error("Failed to parse ts data - %v", err)
			return err
		}

		for i := 0; i < ueNum; i++ {
			c.parseUeData(tsData[3+i], nodeFlag)
		}
	}

	xapp.Logger.Debug("handleIndicationBWP: after parse")
	return nil
}

func (c *Control) handleIndication(params *xapp.RMRParams) (err error) {
	var e2ap *E2ap

	indicationMsg, err := e2ap.GetIndicationMessage(params.Payload)
	c.indTimeMap[params.Meid.RanName] = time.Now()
	fmt.Printf("\nhandleIndication {%s} time=%v\n", params.Meid.RanName, c.indTimeMap[params.Meid.RanName])
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication message: %v", err)
		return
	}

	xapp.Logger.Debug("handleIndication indCount=%d", c.indCount)
	c.indCount++

	fmt.Printf("\nRIC Indication message from {%s} received", params.Meid.RanName)
	fmt.Printf("\nRequestID: %d", indicationMsg.RequestID)
	fmt.Printf("\nRequestSequenceNumber: %d", indicationMsg.RequestSequenceNumber)
	fmt.Printf("\nFunctionID: %d", indicationMsg.FuncID)
	fmt.Printf("\nActionID: %d", indicationMsg.ActionID)
	fmt.Printf("\nIndicationSN: %d", indicationMsg.IndSN)
	fmt.Printf("\nIndicationType: %d", indicationMsg.IndType)
	fmt.Printf("\nIndicationHeader: %x", indicationMsg.IndHeader)
	fmt.Printf("\nIndicationMessage: %x", indicationMsg.IndMessage)
	fmt.Printf("\nCallProcessID: %x", indicationMsg.CallProcessID)

	if indicationMsg.FuncID == 11 {
		c.handleIndicationKPM(indicationMsg, params.Meid.RanName)
	} else if indicationMsg.FuncID == 12 {
		c.handleIndicationBWP(indicationMsg, params.Meid.RanName)
	}

	return nil
}

func (c *Control) handleIndicationKPM(indicationMsg *DecodedIndicationMessage, ranName string) (err error) {
	var e2sm *E2sm

	indicationHdr, err := e2sm.GetIndicationHeader(indicationMsg.IndHeader)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Header: %v", err)
		return
	}

	var cellIDHdr string
	var plmnIDHdr string
	var sliceIDHdr int32
	var fiveQIHdr int64

	fmt.Printf("\n-----------RIC Indication Header-----------")
	if indicationHdr.IndHdrType == 1 {
		fmt.Printf("\nRIC Indication Header Format: %d", indicationHdr.IndHdrType)
		indHdrFormat1 := indicationHdr.IndHdr.(*IndicationHeaderFormat1) // updated by sww, ITRI

		fmt.Printf("\nGlobalKPMnodeIDType: %d", indHdrFormat1.GlobalKPMnodeIDType)

		if indHdrFormat1.GlobalKPMnodeIDType == 1 {
			globalKPMnodegNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodegNBIDType) // updated by sww, ITRI

			globalgNBID := globalKPMnodegNBID.GlobalgNBID

			fmt.Printf("\nPlmnID: %x", globalgNBID.PlmnID.Buf)
			fmt.Printf("\ngNB ID Type: %d", globalgNBID.GnbIDType)
			if globalgNBID.GnbIDType == 1 {
				gNBID := globalgNBID.GnbID.(*GNBID) // updated by sww, ITRI
				fmt.Printf("\ngNB ID ID: %x, Unused: %d", gNBID.Buf, gNBID.BitsUnused)
			}

			if globalKPMnodegNBID.GnbCUUPID != nil {
				fmt.Printf("\ngNB-CU-UP ID: %x", globalKPMnodegNBID.GnbCUUPID.Buf)
			}

			if globalKPMnodegNBID.GnbDUID != nil {
				fmt.Printf("\ngNB-DU ID: %x", globalKPMnodegNBID.GnbDUID.Buf)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 2 {
			globalKPMnodeengNBID := indHdrFormat1.GlobalKPMnodeID.(GlobalKPMnodeengNBIDType)

			fmt.Printf("\nPlmnID: %x", globalKPMnodeengNBID.PlmnID.Buf)
			fmt.Printf("\nen-gNB ID Type: %d", globalKPMnodeengNBID.GnbIDType)
			if globalKPMnodeengNBID.GnbIDType == 1 {
				engNBID := globalKPMnodeengNBID.GnbID.(*ENGNBID) // updated by sww, ITRI
				fmt.Printf("\nen-gNB ID ID: %x, Unused: %d", engNBID.Buf, engNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 3 {
			globalKPMnodengeNBID := indHdrFormat1.GlobalKPMnodeID.(GlobalKPMnodengeNBIDType)

			fmt.Printf("\nPlmnID: %x", globalKPMnodengeNBID.PlmnID.Buf)
			fmt.Printf("\nng-eNB ID Type: %d", globalKPMnodengeNBID.EnbIDType)
			if globalKPMnodengeNBID.EnbIDType == 1 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_Macro) // updated by sww, ITRI
				fmt.Printf("\nng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 2 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_ShortMacro) // updated by sww, ITRI
				fmt.Printf("\nng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 3 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_LongMacro) // updated by sww, ITRI
				fmt.Printf("\nng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 4 {
			globalKPMnodeeNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodeeNBIDType) // updated by sww, ITRI

			fmt.Printf("\nPlmnID: %x", globalKPMnodeeNBID.PlmnID.Buf)
			fmt.Printf("\neNB ID Type: %d", globalKPMnodeeNBID.EnbIDType)
			if globalKPMnodeeNBID.EnbIDType == 1 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Macro) // updated by sww, ITRI
				fmt.Printf("\neNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 2 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Home) // updated by sww, ITRI
				fmt.Printf("\neNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 3 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_ShortMacro) // updated by sww, ITRI
				fmt.Printf("\neNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 4 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_LongMacro) // updated by sww, ITRI
				fmt.Printf("\neNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			}

		}

		if indHdrFormat1.NRCGI != nil {
			fmt.Printf("\nnRCGI.PlmnID: %x nRCGI.NRCellID ID: %x, Unused: %d",  indHdrFormat1.NRCGI.PlmnID.Buf, indHdrFormat1.NRCGI.NRCellID.Buf, indHdrFormat1.NRCGI.NRCellID.BitsUnused)

			cellIDHdr, err = e2sm.ParseNRCGI(*indHdrFormat1.NRCGI)
			if err != nil {
				xapp.Logger.Error("Failed to parse NRCGI in RIC Indication Header: %v", err)
				return
			}
		} else {
			cellIDHdr = ""
		}

		if indHdrFormat1.PlmnID != nil {
			fmt.Printf("\nPlmnID: %x", indHdrFormat1.PlmnID.Buf)

			plmnIDHdr, err = e2sm.ParsePLMNIdentity(indHdrFormat1.PlmnID.Buf, indHdrFormat1.PlmnID.Size)
			if err != nil {
				xapp.Logger.Error("Failed to parse PlmnID in RIC Indication Header: %v", err)
				return
			}
		} else {
			plmnIDHdr = ""
		}

		if indHdrFormat1.SliceID != nil {
			fmt.Printf("\nSST: %x", indHdrFormat1.SliceID.SST.Buf)

			if indHdrFormat1.SliceID.SD != nil {
				fmt.Printf("\nSD: %x", indHdrFormat1.SliceID.SD.Buf)
			}

			sliceIDHdr, err = e2sm.ParseSliceID(*indHdrFormat1.SliceID)
			if err != nil {
				xapp.Logger.Error("Failed to parse SliceID in RIC Indication Header: %v", err)
				return
			}
		} else {
			sliceIDHdr = -1
		}

		if indHdrFormat1.FiveQI != -1 {
			fmt.Printf("\n5QI: %d\n", indHdrFormat1.FiveQI)
		}
		fiveQIHdr = indHdrFormat1.FiveQI

		if indHdrFormat1.Qci != -1 {
			fmt.Printf("\nQCI: %d", indHdrFormat1.Qci)
		}

		if indHdrFormat1.UeMessageType != -1 {
			fmt.Printf("\nUe Report type: %d", indHdrFormat1.UeMessageType)
		}

		if indHdrFormat1.GnbDUID != nil {
			fmt.Printf("\ngNB-DU-ID: %x", indHdrFormat1.GnbDUID.Buf)
		}

		if indHdrFormat1.GnbNameType == 1 {
			fmt.Printf("\ngNB-DU-Name: %x", (indHdrFormat1.GnbName.(*GNB_DU_Name)).Buf) // updated by sww, ITRI
		} else if indHdrFormat1.GnbNameType == 2 {
			fmt.Printf("\ngNB-CU-CP-Name: %x", (indHdrFormat1.GnbName.(*GNB_CU_CP_Name)).Buf) // updated by sww, ITRI
		} else if indHdrFormat1.GnbNameType == 3 {
			fmt.Printf("\ngNB-CU-UP-Name: %x", (indHdrFormat1.GnbName.(*GNB_CU_UP_Name)).Buf) // updated by sww, ITRI
		}

		if indHdrFormat1.GlobalgNBID != nil {
			fmt.Printf("\nPlmnID: %x", indHdrFormat1.GlobalgNBID.PlmnID.Buf)
			fmt.Printf("\ngNB ID Type: %d", indHdrFormat1.GlobalgNBID.GnbIDType)
			if indHdrFormat1.GlobalgNBID.GnbIDType == 1 {
				gNBID := indHdrFormat1.GlobalgNBID.GnbID.(*GNBID) // updated by sww, ITRI
				fmt.Printf("\ngNB ID ID: %x, Unused: %d", gNBID.Buf, gNBID.BitsUnused)
			}
		}

	} else {
		xapp.Logger.Error("Unknown RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		return
	}

	indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
		return
	}

	var nodeFlag int
	var flag bool
	var containerType int32
	var timestampPDCPBytes *Timestamp
	var dlPDCPBytes int64
	var ulPDCPBytes int64
	var timestampPRB *Timestamp
	var availPRBDL int64
	var availPRBUL int64

	xapp.Logger.Debug("handleIndication: after GetIndicationMessage")
	fmt.Printf("\n-----------RIC Indication Message-----------\n")
	fmt.Printf("StyleType: %d IndMsgType: %d\n", indMsg.StyleType, indMsg.IndMsgType)
	if indMsg.IndMsgType == 1 {
		indMsgFormat1 := indMsg.IndMsg.(*IndicationMessageFormat1) // updated by sww, ITRI
		fmt.Printf("PMContainerCount: %d\n", indMsgFormat1.PMContainerCount)
		nodeFlag = 0

		for i := 0; i < indMsgFormat1.PMContainerCount; i++ {
			flag = false
			timestampPDCPBytes = nil
			dlPDCPBytes = -1
			ulPDCPBytes = -1
			timestampPRB = nil
			availPRBDL = -1
			availPRBUL = -1

			pmContainer := indMsgFormat1.PMContainers[i]

			if pmContainer.PFContainer != nil {
				containerType = pmContainer.PFContainer.ContainerType
				fmt.Printf("PMContainers[%d]: PFContainerType: %d\n", i, containerType)

				if containerType == 1 {
					fmt.Println("oDU PF Container: ")
					nodeFlag = nodeFlag | NODE_FLAG_DU

					oDU := pmContainer.PFContainer.Container.(*ODUPFContainerType) // updated by sww, ITRI

					cellResourceReportCount := oDU.CellResourceReportCount
					fmt.Printf("CellResourceReportCount: %d", cellResourceReportCount)

					for j := 0; j < cellResourceReportCount; j++ {
						fmt.Printf("\nCellResourceReport[%d]: ", j)

						cellResourceReport := oDU.CellResourceReports[j]

						fmt.Printf("\nnRCGI.PlmnID: %x nRCGI.nRCellID: %x", cellResourceReport.NRCGI.PlmnID.Buf, cellResourceReport.NRCGI.NRCellID.Buf)

						cellID, err := e2sm.ParseNRCGI(cellResourceReport.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse CellID in DU PF Container: %v", err)
							fmt.Printf("\nFailed to parse CellID in DU PF Container: %v", err)
							continue
						}
						fmt.Printf("\nCompare cellID: %s cellIDHdr: %s\n", cellID, cellIDHdr)
						if cellID == cellIDHdr {
							flag = true
						}

						fmt.Printf("TotalofAvailablePRBsDL/UL: %d %d\n", cellResourceReport.TotalofAvailablePRBs.DL, cellResourceReport.TotalofAvailablePRBs.UL)

						if flag {
							availPRBDL = cellResourceReport.TotalofAvailablePRBs.DL
							availPRBUL = cellResourceReport.TotalofAvailablePRBs.UL
						}

						servedPlmnPerCellCount := cellResourceReport.ServedPlmnPerCellCount
						fmt.Printf("\nServedPlmnPerCellCount: %d", servedPlmnPerCellCount)

						for k := 0; k < servedPlmnPerCellCount; k++ {
							fmt.Printf("\nServedPlmnPerCell[%d]: ", k)

							servedPlmnPerCell := cellResourceReport.ServedPlmnPerCells[k]

							fmt.Printf("\nPlmnID: %x", servedPlmnPerCell.PlmnID.Buf)

							if servedPlmnPerCell.DUPM5GC != nil {
								slicePerPlmnPerCellCount := servedPlmnPerCell.DUPM5GC.SlicePerPlmnPerCellCount
								fmt.Printf("\nSlicePerPlmnPerCellCount: %d", slicePerPlmnPerCellCount)

								for l := 0; l < slicePerPlmnPerCellCount; l++ {
									fmt.Printf("\nSlicePerPlmnPerCell[%d]: ", l)

									slicePerPlmnPerCell := servedPlmnPerCell.DUPM5GC.SlicePerPlmnPerCells[l]

									fmt.Printf("\nSliceID.sST: %x", slicePerPlmnPerCell.SliceID.SST.Buf)
									if slicePerPlmnPerCell.SliceID.SD != nil {
										fmt.Printf("\nSliceID.sD: %x", slicePerPlmnPerCell.SliceID.SD.Buf)
									}

									fQIPERSlicesPerPlmnPerCellCount := slicePerPlmnPerCell.FQIPERSlicesPerPlmnPerCellCount
									fmt.Printf("\n5QIPerSlicesPerPlmnPerCellCount: %d", fQIPERSlicesPerPlmnPerCellCount)

									for m := 0; m < fQIPERSlicesPerPlmnPerCellCount; m++ {
										fmt.Printf("\n5QIPerSlicesPerPlmnPerCell[%d]: ", m)

										fQIPERSlicesPerPlmnPerCell := slicePerPlmnPerCell.FQIPERSlicesPerPlmnPerCells[m]

										fmt.Printf("\n5QI: %d", fQIPERSlicesPerPlmnPerCell.FiveQI)
										fmt.Printf("\nPrbUsageDL: %d", fQIPERSlicesPerPlmnPerCell.PrbUsage.DL)
										fmt.Printf("\nPrbUsageUL: %d", fQIPERSlicesPerPlmnPerCell.PrbUsage.UL)
									}
								}
							}

							if servedPlmnPerCell.DUPMEPC != nil {
								perQCIReportCount := servedPlmnPerCell.DUPMEPC.PerQCIReportCount
								fmt.Printf("\nPerQCIReportCount: %d", perQCIReportCount)

								for l := 0; l < perQCIReportCount; l++ {
									fmt.Printf("\nPerQCIReports[%d]: ", l)

									perQCIReport := servedPlmnPerCell.DUPMEPC.PerQCIReports[l]

									fmt.Printf("\nQCI: %d", perQCIReport.QCI)
									fmt.Printf("\nPrbUsageDL: %d", perQCIReport.PrbUsage.DL)
									fmt.Printf("\nPrbUsageUL: %d", perQCIReport.PrbUsage.UL)
								}
							}
						}
					}
				} else if containerType == 2 {
					nodeFlag = nodeFlag | NODE_FLAG_CUCP
					fmt.Println("oCU-CP PF Container:")

					oCUCP := pmContainer.PFContainer.Container.(*OCUCPPFContainerType) // updated by sww, ITRI

					if oCUCP.GNBCUCPName != nil {
						fmt.Printf("gNB-CU-CP Name: %x", oCUCP.GNBCUCPName.Buf)
					}

					fmt.Printf("NumberOfActiveUEs: %d\n", oCUCP.CUCPResourceStatus.NumberOfActiveUEs)

					flag = true
				} else if containerType == 3 {
					nodeFlag = nodeFlag | NODE_FLAG_CUUP
					fmt.Println("oCU-UP PF Container:")

					oCUUP := pmContainer.PFContainer.Container.(*OCUUPPFContainerType) // updated by sww, ITRI

					if oCUUP.GNBCUUPName != nil {
						fmt.Printf("gNB-CU-UP Name: %x", oCUUP.GNBCUUPName.Buf)
					}

					cuUPPFContainerItemCount := oCUUP.CUUPPFContainerItemCount
					fmt.Printf("CU-UP PF Container Item Count: %d", cuUPPFContainerItemCount)

					for j := 0; j < cuUPPFContainerItemCount; j++ {
						fmt.Printf("\nCU-UP PF Container Item [%d]: ", j)

						cuUPPFContainerItem := oCUUP.CUUPPFContainerItems[j]

						fmt.Printf("\nInterfaceType: %d", cuUPPFContainerItem.InterfaceType)

						cuUPPlmnCount := cuUPPFContainerItem.OCUUPPMContainer.CUUPPlmnCount
						fmt.Printf("\nCU-UP Plmn Count: %d", cuUPPlmnCount)

						for k := 0; k < cuUPPlmnCount; k++ {
							cuUPPlmn := cuUPPFContainerItem.OCUUPPMContainer.CUUPPlmns[k]

							fmt.Printf("\nCU-CP PlmnID: %x", cuUPPlmn.PlmnID.Buf)

							plmnID, err := e2sm.ParsePLMNIdentity(cuUPPlmn.PlmnID.Buf, cuUPPlmn.PlmnID.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse PlmnID in CU-UP PF Container: %v", err)
								continue
							}

							if cuUPPlmn.CUUPPM5GC != nil {
								sliceToReportCount := cuUPPlmn.CUUPPM5GC.SliceToReportCount
								fmt.Printf("\nSliceToReportCount: %d", sliceToReportCount)

								for l := 0; l < sliceToReportCount; l++ {
									fmt.Printf("\nSliceToReport[%d]: ", l)

									sliceToReport := cuUPPlmn.CUUPPM5GC.SliceToReports[l]

									fmt.Printf("\nSliceID.sST: %x", sliceToReport.SliceID.SST.Buf)
									if sliceToReport.SliceID.SD != nil {
										fmt.Printf("\nSliceID.sD: %x", sliceToReport.SliceID.SD.Buf)
									}

									sliceID, err := e2sm.ParseSliceID(sliceToReport.SliceID)
									if err != nil {
										xapp.Logger.Error("Failed to parse sliceID in CU-UP PF Container with PlmnID [%s]: %v", plmnID, err)
										continue
									}

									fQIPERSlicesPerPlmnCount := sliceToReport.FQIPERSlicesPerPlmnCount
									fmt.Printf("\n5QIPerSlicesPerPlmnCount: %d", fQIPERSlicesPerPlmnCount)

									for m := 0; m < fQIPERSlicesPerPlmnCount; m++ {
										fmt.Printf("\n5QIPerSlicesPerPlmn[%d]: ", m)

										fQIPERSlicesPerPlmn := sliceToReport.FQIPERSlicesPerPlmns[m]

										fiveQI := fQIPERSlicesPerPlmn.FiveQI

										fmt.Printf("\nCompare Hdr: plmnID: %s %s sliceID: %d %d fiveQI: %d %d\n", plmnID, plmnIDHdr, sliceID, sliceIDHdr, fiveQI, fiveQIHdr)
										if plmnID == plmnIDHdr && sliceID == sliceIDHdr && fiveQI == fiveQIHdr {
											flag = true
										}

										if fQIPERSlicesPerPlmn.PDCPBytesDL != nil {
											if flag {
												dlPDCPBytes, err = e2sm.ParseInteger(fQIPERSlicesPerPlmn.PDCPBytesDL.Buf, fQIPERSlicesPerPlmn.PDCPBytesDL.Size)
												if err != nil {
													xapp.Logger.Error("Failed to parse PDCPBytesDL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													continue
												}
												fmt.Printf("PDCPBytesDL:%d\n", dlPDCPBytes)
											}
										}

										if fQIPERSlicesPerPlmn.PDCPBytesUL != nil {
											if flag {
												ulPDCPBytes, err = e2sm.ParseInteger(fQIPERSlicesPerPlmn.PDCPBytesUL.Buf, fQIPERSlicesPerPlmn.PDCPBytesUL.Size)
												if err != nil {
													xapp.Logger.Error("Failed to parse PDCPBytesUL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													continue
												}
												fmt.Printf("PDCPBytesUL:%d\n", ulPDCPBytes)
											}
										}
									}
								}
							}

							if cuUPPlmn.CUUPPMEPC != nil {
								cuUPPMEPCPerQCIReportCount := cuUPPlmn.CUUPPMEPC.CUUPPMEPCPerQCIReportCount
								fmt.Printf("\nPerQCIReportCount: %d", cuUPPMEPCPerQCIReportCount)

								for l := 0; l < cuUPPMEPCPerQCIReportCount; l++ {
									fmt.Printf("\nPerQCIReport[%d]: ")

									cuUPPMEPCPerQCIReport := cuUPPlmn.CUUPPMEPC.CUUPPMEPCPerQCIReports[l]

									fmt.Printf("\nQCI: %d", cuUPPMEPCPerQCIReport.QCI)

									if cuUPPMEPCPerQCIReport.PDCPBytesDL != nil {
										fmt.Printf("\nPDCPBytesDL: %x", cuUPPMEPCPerQCIReport.PDCPBytesDL.Buf)
									}
									if cuUPPMEPCPerQCIReport.PDCPBytesUL != nil {
										fmt.Printf("\nPDCPBytesUL: %x", cuUPPMEPCPerQCIReport.PDCPBytesUL.Buf)
									}
								}
							}
						}
					}
				} else {
					xapp.Logger.Error("Unknown PF Container type: %d", containerType)
					continue
				}
			}

			if pmContainer.RANContainer != nil {
				timestamp, _ := e2sm.ParseTimestamp(pmContainer.RANContainer.Timestamp.Buf, pmContainer.RANContainer.Timestamp.Size)
				fmt.Printf("\nRANContainer: Timestamp=[sec: %d, nsec: %d]", timestamp.TVsec, timestamp.TVnsec)

				containerType = pmContainer.RANContainer.ContainerType
				if containerType == 1 {
					fmt.Printf("\nRANContainer DU Usage Report: ")

					oDUUE := pmContainer.RANContainer.Container.(*DUUsageReportType) // updated by sww, ITRI

					for j := 0; j < oDUUE.CellResourceReportItemCount; j++ {
						cellResourceReportItem := oDUUE.CellResourceReportItems[j]

						fmt.Printf("\nnRCGI.PlmnID: %x, Size=%d", cellResourceReportItem.NRCGI.PlmnID.Buf, cellResourceReportItem.NRCGI.PlmnID.Size)
						fmt.Printf("\nnRCGI.NRCellID: %x, Size=%d, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.Size, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in DU Usage Report: %v", err)
							continue
						}

						for k := 0; k < cellResourceReportItem.UeResourceReportItemCount; k++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[k]


							fmt.Printf("\nC-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in DU Usage Report with Serving Cell ID [%s]: %v", servingCellID, err)
								continue
							}

							fmt.Printf("\n UE ID [%d]: ueResourceReportItem.PRBUsageDL/UL=%d %d\n", ueID, ueResourceReportItem.PRBUsageDL, ueResourceReportItem.PRBUsageUL)

							var ueMetrics UeMetricsEntry // updated by sww, ITRI - exception fixed

							// updated by sww, ITRI - sdl
							c.sdlAccessMu.Lock()
							retMap, err := sdlUE.Get([]string{strconv.FormatInt(ueID, 10)})

							if err != nil {
								panic(err)
							} else {
								xapp.Logger.Debug("handleIndication: oDUUE: sdl get")
								fmt.Println(retMap)
								if retMap[strconv.FormatInt(ueID, 10)] != nil {
									ueJsonStr := retMap[strconv.FormatInt(ueID, 10)].(string)
									json.Unmarshal([]byte(ueJsonStr), &ueMetrics)
									fmt.Println(ueMetrics)
								} else {
									ueMetrics = UeMetricsEntry{}
									ueMetrics.UeID = strconv.FormatInt(ueID, 10) // updated by sww, ITRI
									fmt.Println("\n oDUUE: not Exists")
								}
							}

/*
							if isUeExist, _ := c.client.Exists(strconv.FormatInt(ueID, 10)).Result(); isUeExist == 1 {
								ueJsonStr, _ := c.client.Get(strconv.FormatInt(ueID, 10)).Result()
								json.Unmarshal([]byte(ueJsonStr), &ueMetrics)
							} else {
								ueMetrics = UeMetricsEntry{}
							}
*/


							ueMetrics.ServingCellID = servingCellID

							if flag {
								timestampPRB = timestamp
							}

							ueMetrics.MeasTimestampPRB.TVsec = timestamp.TVsec
							ueMetrics.MeasTimestampPRB.TVnsec = timestamp.TVnsec

							if ueResourceReportItem.PRBUsageDL != -1 {
								ueMetrics.PRBUsageDL = ueResourceReportItem.PRBUsageDL
							}

							if ueResourceReportItem.PRBUsageUL != -1 {
								ueMetrics.PRBUsageUL = ueResourceReportItem.PRBUsageUL
							}

							newUeJsonStr, err := json.Marshal(ueMetrics)
							if err != nil {
								xapp.Logger.Error("Failed to marshal UeMetrics with UE ID [%s]: %v", ueID, err)
							} else {
								fmt.Printf("\n to set UeMetrics - UE ID [%d]: ueMetrics.PRBUsageDL=%d PRBUsageUL=%d\n", ueID,  ueMetrics.PRBUsageDL,  ueMetrics.PRBUsageUL)

								// updated by sww, ITRI - sdl
								xapp.Logger.Debug("handleIndication: oDUUE: sdl set")
								fmt.Println(ueMetrics)
								err = sdlUE.Set(strconv.FormatInt(ueID, 10), newUeJsonStr)
//								err = c.client.Set(strconv.FormatInt(ueID, 10), newUeJsonStr, 0).Err()
							}
							c.sdlAccessMu.Unlock()

							if err != nil {
								xapp.Logger.Error("Failed to set UeMetrics into redis with UE ID [%s]: %v", ueID, err)
								continue
							}
						}
					}
				} else if containerType == 2 {
					fmt.Printf("\nRANContainer CU-CP Usage Report: ")

					oCUCPUE := pmContainer.RANContainer.Container.(*CUCPUsageReportType) // updated by sww, ITRI

					for j := 0; j < oCUCPUE.CellResourceReportItemCount; j++ {
						cellResourceReportItem := oCUCPUE.CellResourceReportItems[j]

						fmt.Printf("\nnRCGI.PlmnID: %x", cellResourceReportItem.NRCGI.PlmnID.Buf)
						fmt.Printf("\nnRCGI.NRCellID: %x, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in CU-CP Usage Report: %v", err)
							continue
						}

						for k := 0; k < cellResourceReportItem.UeResourceReportItemCount; k++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[k]

							fmt.Printf("\nC-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in CU-CP Usage Report with Serving Cell ID [%s]: %v", err)
								continue
							}

							var ueMetrics UeMetricsEntry // updated by sww, ITRI
/*
							if isUeExist, _ := c.client.Exists(strconv.FormatInt(ueID, 10)).Result(); isUeExist == 1 {
								ueJsonStr, _ := c.client.Get(strconv.FormatInt(ueID, 10)).Result()
								json.Unmarshal([]byte(ueJsonStr), &ueMetrics) // updated by sww, ITRI
							} else {
								ueMetrics = UeMetricsEntry{}
							}
*/

							// updated by sww, ITRI - sdl
							c.sdlAccessMu.Lock()
							retMap, err := sdlUE.Get([]string{strconv.FormatInt(ueID, 10)})

							if err != nil {
								panic(err)
							} else {
								fmt.Println("\n oCUCPUE: sdl get")
								fmt.Println(retMap)
								if retMap[strconv.FormatInt(ueID, 10)] != nil {
									ueJsonStr := retMap[strconv.FormatInt(ueID, 10)].(string)
									json.Unmarshal([]byte(ueJsonStr), &ueMetrics)
									fmt.Println(ueMetrics)
								} else {
									fmt.Println("\n oCUCPUE: not Exists")
									ueMetrics = UeMetricsEntry{}
									ueMetrics.UeID = strconv.FormatInt(ueID, 10) // updated by sww, ITRI
								}
							}


							ueMetrics.ServingCellID = servingCellID

							ueMetrics.MeasTimeRF.TVsec = timestamp.TVsec
							ueMetrics.MeasTimeRF.TVnsec = timestamp.TVnsec

							fmt.Println("\n -------- check ServingCellRF/NeighborCellRF: ----------")
							fmt.Println(ueResourceReportItem.ServingCellRF)
							fmt.Println(ueResourceReportItem.NeighborCellRF)
							if ueResourceReportItem.ServingCellRF != nil {
								// updated by sww, ITRI
								servingCellRF, _ := e2sm.ParseCellRF(ueResourceReportItem.ServingCellRF.Buf, ueResourceReportItem.ServingCellRF.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse ServingCellRF in CU-CP Usage Report with UE ID [%s]: %v", ueID, err)
									c.sdlAccessMu.Unlock()
									continue
								}

								ueMetrics.ServingCellRF.RSRP = servingCellRF.RSRP
								ueMetrics.ServingCellRF.RSRQ = servingCellRF.RSRQ
								ueMetrics.ServingCellRF.RSSINR = servingCellRF.RSSINR

								bwpData := e2sm.ParseBWPData(ueResourceReportItem.ServingCellRF.Buf, ueResourceReportItem.ServingCellRF.Size)
								if bwpData != nil {
									ueMetrics.BWPData.BWPID = bwpData.BWPID
									ueMetrics.BWPData.LocationAndBandwidth = bwpData.LocationAndBandwidth
								}
/*
								err = json.Unmarshal(ueResourceReportItem.ServingCellRF.Buf, &ueMetrics.ServingCellRF)
								if err != nil {
									xapp.Logger.Error("Failed to Unmarshal ServingCellRF in CU-CP Usage Report with UE ID [%s]: %v", ueID, err)
									continue
								}
*/
							}

							if ueResourceReportItem.NeighborCellRF != nil {
								// updated by sww, ITRI
								ueMetrics.NeighborCellsRF, _ = e2sm.ParseNeighborCellRFList(ueResourceReportItem.NeighborCellRF.Buf, ueResourceReportItem.NeighborCellRF.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse ServingCellRF in CU-CP Usage Report with UE ID [%s]: %v", ueID, err)
									c.sdlAccessMu.Unlock()
									continue
								}

/*								err = json.Unmarshal(ueResourceReportItem.NeighborCellRF.Buf, &ueMetrics.NeighborCellsRF)
								if err != nil {
									xapp.Logger.Error("Failed to Unmarshal NeighborCellRF in CU-CP Usage Report with UE ID [%s]: %v", ueID, err)
									continue
								}
*/
							}

							newUeJsonStr, err := json.Marshal(ueMetrics)
							if err != nil {
								xapp.Logger.Error("Failed to marshal UeMetrics with UE ID [%s]: %v", ueID, err)
							} else {
								// updated by sww, ITRI - sdl
								xapp.Logger.Debug("handleIndication: oCUCPUE: sdl set")
								fmt.Println(ueMetrics)
								err = sdlUE.Set(strconv.FormatInt(ueID, 10), newUeJsonStr)
//								err = c.client.Set(strconv.FormatInt(ueID, 10), newUeJsonStr, 0).Err()
							}
							c.sdlAccessMu.Unlock()

							if err != nil {
								xapp.Logger.Error("Failed to set UeMetrics into redis with UE ID [%s]: %v", ueID, err)
								continue
							}
						}
					}
				} else if containerType == 3 { // updated by sww, ITRI - wrong type value fixed
					fmt.Printf("\nRANContainer CU-UP Usage Report: ")

					oCUUPUE := pmContainer.RANContainer.Container.(*CUUPUsageReportType) // updated by sww, ITRI

					for j := 0; j < oCUUPUE.CellResourceReportItemCount; j++ {
						cellResourceReportItem := oCUUPUE.CellResourceReportItems[j]

						fmt.Printf("\nnRCGI.PlmnID: %x", cellResourceReportItem.NRCGI.PlmnID.Buf)
						fmt.Printf("\nnRCGI.NRCellID: %x, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in CU-UP Usage Report: %v", err)
							continue
						}

						for k := 0; k < cellResourceReportItem.UeResourceReportItemCount; k++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[k]

							fmt.Printf("\nC-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in CU-UP Usage Report Serving Cell ID [%s]: %v", servingCellID, err)
								continue
							}

							var ueMetrics UeMetricsEntry // updated by sww, ITRI - exception fixed

							// updated by sww, ITRI - sdl
							c.sdlAccessMu.Lock()
							retMap, err := sdlUE.Get([]string{strconv.FormatInt(ueID, 10)})

							if err != nil {
								panic(err)
							} else {
								fmt.Println("\n oCUUPUE: sdl get")
								fmt.Println(retMap)
								if retMap[strconv.FormatInt(ueID, 10)] != nil {
									ueJsonStr := retMap[strconv.FormatInt(ueID, 10)].(string)
									json.Unmarshal([]byte(ueJsonStr), &ueMetrics)
									fmt.Println(ueMetrics)
								} else {
									fmt.Println("\n oCUUPUE: not Exists")
									ueMetrics = UeMetricsEntry{}
									ueMetrics.UeID = strconv.FormatInt(ueID, 10) // updated by sww, ITRI
								}
							}

/*
							if isUeExist, _ := c.client.Exists(strconv.FormatInt(ueID, 10)).Result(); isUeExist == 1 {
								ueJsonStr, _ := c.client.Get(strconv.FormatInt(ueID, 10)).Result()
								json.Unmarshal([]byte(ueJsonStr), &ueMetrics) // updated by sww, ITRI - exception fixed
							} else {
								ueMetrics = UeMetricsEntry{} // updated by sww, ITRI - exception fixed
							}
*/

							ueMetrics.ServingCellID = servingCellID

							if flag {
								timestampPDCPBytes = timestamp
							}

							ueMetrics.MeasTimestampPDCPBytes.TVsec = timestamp.TVsec
							ueMetrics.MeasTimestampPDCPBytes.TVnsec = timestamp.TVnsec

							if ueResourceReportItem.PDCPBytesDL != nil {
								ueMetrics.PDCPBytesDL, err = e2sm.ParseInteger(ueResourceReportItem.PDCPBytesDL.Buf, ueResourceReportItem.PDCPBytesDL.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse PDCPBytesDL in CU-UP Usage Report with UE ID [%s]: %v", ueID, err)
									c.sdlAccessMu.Unlock()
									continue
								}
							}

							if ueResourceReportItem.PDCPBytesUL != nil {
								ueMetrics.PDCPBytesUL, err = e2sm.ParseInteger(ueResourceReportItem.PDCPBytesUL.Buf, ueResourceReportItem.PDCPBytesUL.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse PDCPBytesUL in CU-UP Usage Report with UE ID [%s]: %v", ueID, err)
									c.sdlAccessMu.Unlock()
									continue
								}
							}

							newUeJsonStr, err := json.Marshal(ueMetrics)
							if err != nil {
								xapp.Logger.Error("Failed to marshal UeMetrics with UE ID [%s]: %v", ueID, err)
							} else {
								// updated by sww, ITRI - sdl
								fmt.Println(" oCUUPUE: sdl set")
								fmt.Println(ueMetrics)
								err = sdlUE.Set(strconv.FormatInt(ueID, 10), newUeJsonStr)
//								err = c.client.Set(strconv.FormatInt(ueID, 10), newUeJsonStr, 0).Err()
							}
							c.sdlAccessMu.Unlock()

							if err != nil {
								xapp.Logger.Error("Failed to set UeMetrics into redis with UE ID [%s]: %v", ueID, err)
								continue
							}
						}
					}
				} else {
					xapp.Logger.Error("Unknown PF Container Type: %d", containerType)
					continue
				}
			}

			if flag {
				var cellMetrics CellMetricsEntry // updated by sww, ITRI - exception fixed

				// updated by sww, ITRI - sdl
				retMap, err := sdlCELL.Get([]string{cellIDHdr})

				if err != nil {
					panic(err)
				} else {
					fmt.Println("\n cell: sdl get")
					fmt.Println(retMap)
					if retMap[cellIDHdr] != nil {
						cellJsonStr := retMap[cellIDHdr].(string)
						json.Unmarshal([]byte(cellJsonStr), &cellMetrics)
						fmt.Println(cellJsonStr)
						fmt.Println(cellMetrics)
					} else {
						fmt.Println("\n cell: not Exists")
						cellMetrics = CellMetricsEntry{}
					}
				}

/*
				if isCellExist, _ := c.client.Exists(cellIDHdr).Result(); isCellExist == 1 {
					cellJsonStr, _ := c.client.Get(cellIDHdr).Result()
					json.Unmarshal([]byte(cellJsonStr), &cellMetrics) // updated by sww, ITRI - exception fixed
				} else {
					cellMetrics = CellMetricsEntry{} // updated by sww, ITRI - exception fixed
				}
*/

//				fmt.Printf("check cucp=%v cellMetrics.RANName=%s ranName=%s\n", cucp, cellMetrics.RANName, ranName)
			        c.updateRanMap(ranName, cellIDHdr, nodeFlag)
				if (nodeFlag & NODE_FLAG_CUCP) > 0 {
//			                c.cidMap[ranName] = cellIDHdr
					cellMetrics.RANName = ranName
				}
				if timestampPDCPBytes != nil {
					cellMetrics.MeasTimestampPDCPBytes.TVsec = timestampPDCPBytes.TVsec
					cellMetrics.MeasTimestampPDCPBytes.TVnsec = timestampPDCPBytes.TVnsec
				}
				if dlPDCPBytes != -1 {
					cellMetrics.PDCPBytesDL = dlPDCPBytes
				}
				if ulPDCPBytes != -1 {
					cellMetrics.PDCPBytesUL = ulPDCPBytes
				}
				if timestampPRB != nil {
					cellMetrics.MeasTimestampPRB.TVsec = timestampPRB.TVsec
					cellMetrics.MeasTimestampPRB.TVnsec = timestampPRB.TVnsec
				}
				if availPRBDL != -1 {
					cellMetrics.AvailPRBDL = availPRBDL
				}
				if availPRBUL != -1 {
					cellMetrics.AvailPRBUL = availPRBUL
				}

				newCellJsonStr, err := json.Marshal(cellMetrics)
				if err != nil {
					xapp.Logger.Error("Failed to marshal CellMetrics with CellID [%s]: %v", cellIDHdr, err)
					continue
				}
				// updated by sww, ITRI - sdl
				xapp.Logger.Debug("handleIndication: cell: sdl set")
				fmt.Println(cellMetrics)
				err = sdlCELL.Set(cellIDHdr, newCellJsonStr)
//				err = c.client.Set(cellIDHdr, newCellJsonStr, 0).Err()
				if err != nil {
					xapp.Logger.Error("Failed to set CellMetrics into redis with CellID [%s]: %v", cellIDHdr, err)
					continue
				}
			}
		}
	} else {
		xapp.Logger.Error("Unknown RIC Indication Message Format: %d", indMsg.IndMsgType)
		return
	}

	return nil
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_RESP is %d - 12011@%s", params.SubId, params.Meid.RanName)

	ranName := params.Meid.RanName
	c.subCreatedMap[ranName] = SUB_CREATED // updated by sww, ITRI
	c.subIdMap[ranName] = params.SubId
	c.indTimeMap[ranName] = time.Now()
	//fmt.Printf("handleSubscriptionResponse: subCreatedMap[%s]=SUB_CREATED params.SubId=%d\n", ranName, params.SubId)

	c.eventCreateExpiredMu.Lock()
	_, ok := c.eventCreateExpiredMap[ranName]
	if !ok {
		c.eventCreateExpiredMu.Unlock()
		//xapp.Logger.Debug("RIC_SUB_REQ has been deleted!")
		return nil
	} else {
		c.eventCreateExpiredMap[ranName] = true
		c.eventCreateExpiredMu.Unlock()
	}

	var cep *E2ap
	subscriptionResp, err := cep.GetSubscriptionResponseMessage(params.Payload)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Subscription Response message: %v", err)
		return
	}

	fmt.Printf("\nRIC Subscription Response message from {%s} received", params.Meid.RanName)
	fmt.Printf("\nSubscriptionID: %d", params.SubId)
	fmt.Printf("\nRequestID: %d", subscriptionResp.RequestID)
	fmt.Printf("\nRequestSequenceNumber: %d", subscriptionResp.RequestSequenceNumber)
	fmt.Printf("\nFunctionID: %d", subscriptionResp.FuncID)

	fmt.Printf("\nActionAdmittedList:")
	for index := 0; index < subscriptionResp.ActionAdmittedList.Count; index++ {
		fmt.Printf("\n[%d]ActionID: %d", index, subscriptionResp.ActionAdmittedList.ActionID[index])
	}

	fmt.Printf("\nActionNotAdmittedList:")
	for index := 0; index < subscriptionResp.ActionNotAdmittedList.Count; index++ {
		fmt.Printf("\n[%d]ActionID: %d", index, subscriptionResp.ActionNotAdmittedList.ActionID[index])
		fmt.Printf("\n[%d]CauseType: %d    CauseID: %d", index, subscriptionResp.ActionNotAdmittedList.Cause[index].CauseType, subscriptionResp.ActionNotAdmittedList.Cause[index].CauseID)
	}

	return nil
}

func (c *Control) handleSubscriptionFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_FAILURE is %d", params.SubId)

	ranName := params.Meid.RanName
	c.subCreatedMap[ranName] = SUB_CREATED // SWW
	c.subIdMap[ranName] = params.SubId
	c.eventCreateExpiredMu.Lock()
	_, ok := c.eventCreateExpiredMap[ranName]
	if !ok {
		c.eventCreateExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_REQ has been deleted!")
		fmt.Printf("\nRIC_SUB_REQ has been deleted!")
		return nil
	} else {
		c.eventCreateExpiredMap[ranName] = true
		c.eventCreateExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) handleSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_DEL_RESP is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventDeleteExpiredMu.Lock()
	_, ok := c.eventDeleteExpiredMap[ranName]
	if !ok {
		c.eventDeleteExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ has been deleted!")
		return nil
	} else {
		c.eventDeleteExpiredMap[ranName] = true
		c.eventDeleteExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) handleSubscriptionDeleteFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_DEL_FAILURE is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventDeleteExpiredMu.Lock()
	_, ok := c.eventDeleteExpiredMap[ranName]
	if !ok {
		c.eventDeleteExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ has been deleted!")
		return nil
	} else {
		c.eventDeleteExpiredMap[ranName] = true
		c.eventDeleteExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) setEventCreateExpiredTimer(ranName string) {
	c.eventCreateExpiredMu.Lock()
	c.eventCreateExpiredMap[ranName] = false
	c.eventCreateExpiredMu.Unlock()

	timer := time.NewTimer(time.Duration(c.eventCreateExpired) * time.Second)
	go func(t *time.Timer) {
		defer t.Stop()
		xapp.Logger.Debug("RIC_SUB_REQ[%s]: Waiting for RIC_SUB_RESP...", ranName)
		for {
			select {
			case <-t.C:
				c.eventCreateExpiredMu.Lock()
				isResponsed := c.eventCreateExpiredMap[ranName]
				delete(c.eventCreateExpiredMap, ranName)
				c.eventCreateExpiredMu.Unlock()
				if !isResponsed {
					xapp.Logger.Debug("RIC_SUB_REQ[%s]: RIC Event Create Timer experied!", ranName)
					xapp.Logger.Debug("Let us retry...", ranName)
					c.startTimerSubReq()
					// c.sendRicSubDelRequest(subID, requestSN, funcID)
					return
				}
			default:
				c.eventCreateExpiredMu.Lock()
				flag := c.eventCreateExpiredMap[ranName]
				if flag {
					delete(c.eventCreateExpiredMap, ranName)
					c.eventCreateExpiredMu.Unlock()
					xapp.Logger.Debug("RIC_SUB_REQ[%s]: RIC Event Create Timer canceled!", ranName)
					fmt.Printf("\nRIC_SUB_REQ[%s]: RIC Event Create Timer canceled!", ranName)
					return
				} else {
					c.eventCreateExpiredMu.Unlock()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(timer)
}

func (c *Control) setEventDeleteExpiredTimer(ranName string) {
	c.eventDeleteExpiredMu.Lock()
	c.eventDeleteExpiredMap[ranName] = false
	c.eventDeleteExpiredMu.Unlock()

	timer := time.NewTimer(time.Duration(c.eventDeleteExpired) * time.Second)
	go func(t *time.Timer) {
		defer t.Stop()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: Waiting for RIC_SUB_DEL_RESP...", ranName)
		for {
			select {
			case <-t.C:
				c.eventDeleteExpiredMu.Lock()
				isResponsed := c.eventDeleteExpiredMap[ranName]
				delete(c.eventDeleteExpiredMap, ranName)
				c.eventDeleteExpiredMu.Unlock()
				if !isResponsed {
					xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer experied!", ranName)
					return
				}
			default:
				c.eventDeleteExpiredMu.Lock()
				flag := c.eventDeleteExpiredMap[ranName]
				if flag {
					delete(c.eventDeleteExpiredMap, ranName)
					c.eventDeleteExpiredMu.Unlock()
					xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer canceled!", ranName)
					return
				} else {
					c.eventDeleteExpiredMu.Unlock()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(timer)
}

func (c *Control) getRtPeriodEnum(reportPeriod int) (rtPeriodEnum int64) {
	if reportPeriod == 10 {
		return 0
	}
	if reportPeriod == 20 {
		return 1
	}
	if reportPeriod == 32 {
		return 2
	}
	if reportPeriod == 40 {
		return 3
	}
	if reportPeriod == 60 {
		return 4
	}
	if reportPeriod == 64 {
		return 5
	}
	if reportPeriod == 70 {
		return 6
	}
	if reportPeriod == 80 {
		return 7
	}
	if reportPeriod == 128 {
		return 8
	}
	if reportPeriod == 160 {
		return 9
	}
	if reportPeriod == 256 {
		return 10
	}
	if reportPeriod == 320 {
		return 11
	}
	if reportPeriod == 512 {
		return 12
	}
	if reportPeriod == 640 {
		return 13
	}
	if reportPeriod == 1024 {
		return 14
	}
	if reportPeriod == 1280 {
		return 15
	}
	if reportPeriod == 2048 {
		return 16
	}
	if reportPeriod == 2560 {
		return 17
	}
	if reportPeriod == 5120 {
		return 18
	}
	if reportPeriod == 10240 {
		return 19
	}
	return 13
}

func (c *Control) sendRicSubRequest(subID int, requestSN int, funcID int) (err error) {
	var e2ap *E2ap
	var e2sm *E2sm

	var reportPeriod int = xapp.Config.GetInt("controls.reportPeriod")
	var rtPeriodEnum int64 = c.getRtPeriodEnum(reportPeriod)
	fmt.Printf("sendRicSubRequest reportPeriod=%d rtPeriodEnum=%d\n", reportPeriod, rtPeriodEnum)
	var eventTriggerCount int = 1
	var periods []int64 = []int64{rtPeriodEnum}
	var eventTriggerDefinition []byte = make([]byte, 8)
	_, err = e2sm.SetEventTriggerDefinition(eventTriggerDefinition, eventTriggerCount, periods)
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
		return err
	}

	fmt.Printf("Set EventTriggerDefinition: %x\n", eventTriggerDefinition)

	var actionCount int = 1
	var ricStyleType []int64 = []int64{0}
	var actionIds []int64 = []int64{0}
	var actionTypes []int64 = []int64{0}
	var actionDefinitions []ActionDefinition = make([]ActionDefinition, actionCount)
	var subsequentActions []SubsequentAction = []SubsequentAction{SubsequentAction{0, 0, 0}}

	for index := 0; index < actionCount; index++ {
		if ricStyleType[index] == 0 {
			actionDefinitions[index].Buf = nil
			actionDefinitions[index].Size = 0
		} else {
			actionDefinitions[index].Buf = make([]byte, 8)
			_, err = e2sm.SetActionDefinition(actionDefinitions[index].Buf, ricStyleType[index])
			if err != nil {
				xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
				return err
			}
			actionDefinitions[index].Size = len(actionDefinitions[index].Buf)

			fmt.Printf("Set ActionDefinition[%d]: %x\n", index, actionDefinitions[index].Buf)
		}
	}

	sending := false
	for _, ran := range c.ranList {
		if c.subCreatedMap[ran] == SUB_PENDING && !sending { // updated by sww, ITRI
        	        if sending {
				time.Sleep(100 * time.Millisecond)
			}
			params := &xapp.RMRParams{}
			params.Mtype = 12010
			params.SubId = subID

			xapp.Logger.Debug("sendRicSubRequest: Send RIC_SUB_REQ to {%s}", ran)

			params.Payload = make([]byte, 1024)
			params.Payload, err = e2ap.SetSubscriptionRequestPayload(params.Payload, 1001, uint16(requestSN), uint16(funcID), eventTriggerDefinition, len(eventTriggerDefinition), actionCount, actionIds, actionTypes, actionDefinitions, subsequentActions)
			if err != nil {
				xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
				return err
			}

			fmt.Printf("Set Payload: %x\n", params.Payload)

			params.Meid = &xapp.RMRMeid{RanName: ran}
			xapp.Logger.Debug("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)

			err = c.rmrSend(params)
			if err != nil {
				xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
				return err
			}
			sending = true

//			c.setEventCreateExpiredTimer(params.Meid.RanName)
			//c.ranList = append(c.ranList[:index], c.ranList[index+1:]...)
			//index--
		}
	}

	return nil
}

func (c *Control) sendRicSubDelRequest(subID int, requestSN int, funcID int) (err error) {
	params := &xapp.RMRParams{}
	params.Mtype = 12020
	params.SubId = subID
	var e2ap *E2ap

	params.Payload = make([]byte, 1024)
	params.Payload, err = e2ap.SetSubscriptionDeleteRequestPayload(params.Payload, 100, uint16(requestSN), uint16(funcID))
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

	fmt.Printf("\nSet Payload: %x", params.Payload)

	if funcID == 0 {
		params.Meid = &xapp.RMRMeid{PlmnID: "::", EnbID: "::", RanName: "0"}
	} else {
		params.Meid = &xapp.RMRMeid{PlmnID: "::", EnbID: "::", RanName: "3"}
	}

	xapp.Logger.Debug("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)

	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

	c.setEventDeleteExpiredTimer(params.Meid.RanName)

	return nil
}

func (c *Control) sendRicSubDelRequestForRan(ran string, subID int, requestSN int, funcID int) (err error) {
	params := &xapp.RMRParams{}
	params.Mtype = 12020
	params.SubId = subID
	var e2ap *E2ap

	params.Payload = make([]byte, 1024)
	params.Payload, err = e2ap.SetSubscriptionDeleteRequestPayload(params.Payload, uint16(subID), uint16(requestSN), uint16(funcID))
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

	fmt.Printf("\nsendRicSubDelRequestForRan %s Set Payload: %x", ran, params.Payload)

	params.Meid = &xapp.RMRMeid{RanName: ran}
	params.SubId = subID
	xapp.Logger.Debug("The RMR message to be sent is %d with SubId=%d %d", params.Mtype, params.SubId, requestSN)
	//fmt.Printf("\nThe RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)

	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		fmt.Printf("\nFailed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

//	c.setEventDeleteExpiredTimer(params.Meid.RanName)

	return nil
}

func (c *Control) clearRanMap(ranName string) {
	if _, ok := c.ranMap[ranName]; ok {
//		c.ranMap[ranName] = RanData{ c.ranMap[ranName].cellIDs, 0, }
		c.ranMap[ranName].clear()
//		fmt.Printf("clearRanMap: %v\n", c.ranMap[ranName])
	}
}

func (c *Control) updateRanMap(ranName string, cellID string, flag int) {
	if _, ok := c.ranMap[ranName]; ok {
		fmt.Printf(" updateRanMap: exist: %s: %s, %d | %d\n", ranName, cellID, flag, c.ranMap[ranName].indFlag)
		c.ranMap[ranName].update(cellID, flag)
	} else {
		fmt.Printf(" updateRanMap: new: %s: %s, %d\n", ranName, cellID, flag)
		c.ranMap[ranName] = &RanData{
			[]string{cellID},
			flag}
	}
//	fmt.Println(c.ranMap[ranName])
}

func (c *RanData) clear() {
	c.indFlag = 0
}

func (c *RanData) update(cellID string, flag int) {
	c.indFlag = c.indFlag | flag
	if !contains(c.cellIDs, cellID) {
		c.cellIDs = append(c.cellIDs, cellID)
	}
}

