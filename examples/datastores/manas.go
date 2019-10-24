/*
Copyright (c) 2017 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
This example program shows how the `view` and `property` packages can
be used to navigate a vSphere inventory structure using govmomi.
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var jsonParams map[string]interface{}

// ParseJSONInput : Parses a Json File to read SE Ova params
func ParseJSONInput() {
	jsonFile, err := os.Open("seovaparams.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("Successfully Opened seovaparams.json")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	err = json.Unmarshal([]byte(byteValue), &jsonParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	/*
		for k, v := range jsonParams {
			fmt.Println("Key : ", k, "Val : ", v)
		}
	*/
	return
}

// getJSONParam : Retrieves the value of the param
func getJSONParam(key string) string {
	keyInterface, ok := jsonParams[key]
	if !ok {
		fmt.Println(key, " not present in jsonParams")
		return ""
	}
	return (fmt.Sprintf("%v", keyInterface))
}

//VmwareCreateAviSEFromOVF : Creates a SEVM
func VmwareCreateAviSEFromOVF(ctx context.Context, c *vim25.Client) error {
	// Retrieve the host MOR string from JSONParams
	hostName := getJSONParam("vcenterHost")
	if hostName == "" {
		return errors.New("vCenterHost not present in JSONParams")
	}
	host, err := find.NewFinder(c).HostSystem(ctx, hostName)
	if err != nil {
		return err
	}
	pc := property.DefaultCollector(c)
	var hostSystem mo.HostSystem
	err = pc.RetrieveOne(ctx, host.Reference(), []string{"name", "parent", "runtime", "network", "configManager", "hardware", "vm"}, &hostSystem)
	if err != nil {
		return err
	}

	hostCPUCores := hostSystem.Hardware.CpuInfo.NumCpuThreads
	hostCPUMhz := hostSystem.Hardware.CpuInfo.Hz / 1000000

	fmt.Println("==================================================================")
	fmt.Println("Host CPU Cores : ", hostCPUCores)
	fmt.Println("Host CPU Mhz   : ", hostCPUMhz)

	hostMemSize := hostSystem.Hardware.MemorySize / (1000 * 1000)
	fmt.Println("Host memory : ", hostMemSize)
	fmt.Println("==================================================================\n")

	fmt.Println("Validating CPU reservation for Host : ", hostSystem.Name,
		" by scanning through all the VMs")
	var (
		hostTotalCPU              int64
		hostCurrentCPUReservation int64 = 0
		hostCurrentMemReservation int64 = 0
		seCPUReservationNeeded    int64 = 0
	)
	SeCoresRequired, err := strconv.ParseInt(getJSONParam("vcenterNumSeCores"), 10, 64)
	if err != nil {
		return errors.New("Could not retrieve SeCoresRequired")
	}
	SeMemoryRequired, err := strconv.ParseInt(getJSONParam("vcenterNumMem"), 10, 64)
	if err != nil {
		return errors.New("Could not retrieve numSeMemory")
	}

	hostTotalCPU = int64(hostCPUCores) * hostCPUMhz
	seCPUReservationNeeded = SeCoresRequired * hostCPUMhz
	// Calculate the Reservation used by all the powered on VMs in the Host
	for _, vmMor := range hostSystem.Vm {
		var vm mo.VirtualMachine
		err = pc.RetrieveOne(ctx, vmMor.Reference(), []string{"name", "config", "runtime"}, &vm)
		if err != nil {
			continue
		}
		if vm.Runtime.PowerState == types.VirtualMachinePowerStatePoweredOff {
			fmt.Println("VM : ", vm.Name, "\tPowered Off. Ignoring...")
			continue
		}
		hostCurrentCPUReservation += *vm.Config.CpuAllocation.Reservation
		hostCurrentMemReservation += *vm.Config.MemoryAllocation.Reservation
		fmt.Println("VM : ", vm.Name, "\t\tCPU Reservation : ", *vm.Config.CpuAllocation.Reservation,
			"\tMem Reservation : ", *vm.Config.MemoryAllocation.Reservation)
	}
	fmt.Println("\n==================================================================")
	fmt.Println("Host CPU/Memory summary......")
	fmt.Println("==================================================================")
	fmt.Println("Total CPU MHz reserved by existing VMs : ", hostCurrentCPUReservation)
	fmt.Println("CPU MHz needed by new SEVM             : ", seCPUReservationNeeded)
	fmt.Println("Host Total CPU MHz capable             : ", hostTotalCPU)
	fmt.Println("Mem(bytes) reserved by existing VMs    : ", hostCurrentMemReservation)
	fmt.Println("Mem(bytes) needed by new SEVM          : ", SeMemoryRequired)
	fmt.Println("Host Total Mem(bytes) capable          : ", hostMemSize)
	fmt.Println("==================================================================\n")
	if hostCurrentCPUReservation+seCPUReservationNeeded > hostTotalCPU {
		fmt.Println("Host : ", hostSystem.Name, " does not have adequate CPU resources")
		return errors.New("Host : ", hostSystem.Name, " does not have adequate CPU resources")
	}
	if hostCurrentMemReservation+seMemoryRequired > hostMemSize {
		fmt.Println("Host : ", hostSystem.Name, " does not have adequate Memory resources")
		return errors.New("Host : ", hostSystem.Name, " does not have adequate Memory resources")
	}
	return nil
}

// HostList : Retrieves the list of ESX Hosts from vCenter
func HostList(ctx context.Context, c *vim25.Client) error {

	// Create a view of HostSystem objects
	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all hosts
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.HostSystem.html
	var hss []mo.HostSystem
	//err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "parent", "runtime", "network", "configManager", "hardware", "vm"}, &hss)
	if err != nil {
		return err
	}

	// Print summary per host (see also: govc/host/info.go)
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	tw1 := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	//fmt.Fprintf(tw, "Name:\tUsed CPU:\tTotal CPU:\tFree CPU:\tUsed Memory:\tTotal Memory:\tFree Memory:\t\n")
	fmt.Fprintf(tw, "Name:\tNum NWs \tCPU:\tMHz:\tMemory(MB):\tMaintenance:\tConnected:\tPower State:\tCluster:\tNum VMs:\t\n")
	fmt.Fprintf(tw1, "Name:\tDevice:\tChassis:\tPort:\t\n")

	for _, hs := range hss {
		fmt.Fprintf(tw, "%s\t", hs.Name)
		fmt.Fprintf(tw, "%d\t", len(hs.Network))
		fmt.Fprintf(tw, "%d\t", hs.Hardware.CpuInfo.NumCpuThreads)
		fmt.Fprintf(tw, "%d\t", hs.Hardware.CpuInfo.Hz/(1000*1000))
		fmt.Fprintf(tw, "%d\t", hs.Hardware.MemorySize/(1000*1000))
		fmt.Fprintf(tw, "%t\t", hs.Runtime.InMaintenanceMode)
		fmt.Fprintf(tw, "%s\t", hs.Runtime.ConnectionState)
		fmt.Fprintf(tw, "%s\t", hs.Runtime.PowerState)
		parent := *hs.Parent
		if parent.Type == "ClusterComputeResource" {
			cluster := object.NewClusterComputeResource(c, parent)
			clusterName, err := cluster.ObjectName(ctx)
			if err == nil {
				fmt.Fprintf(tw, "%s\t", clusterName)
			}
		} else {
			fmt.Fprintf(tw, "NA\t")
		}
		fmt.Fprintf(tw, "%d\t", len(hs.Vm))
		hostSystem := object.NewHostSystem(c, hs.Reference())
		hostConfigManager := hostSystem.ConfigManager()
		if hostConfigManager.NetworkSystem != nil {
			hostNetworkSystem, err := hostConfigManager.NetworkSystem(ctx)
			if err == nil {
				if hs.Runtime.ConnectionState == "connected" {
					pnhInfo, err := hostNetworkSystem.QueryNetworkHint(ctx, nil)
					if err == nil {
						for _, pnh := range pnhInfo {
							lldpInfo := pnh.LldpInfo
							cdpInfo := pnh.ConnectedSwitchPort
							if lldpInfo != nil {
								fmt.Fprintf(tw1, "%s\t", hs.Name)
								fmt.Fprintf(tw1, "%s\t", pnh.Device)
								fmt.Fprintf(tw1, "%s\t", lldpInfo.ChassisId)
								fmt.Fprintf(tw1, "%s\t", lldpInfo.PortId)
							} else if cdpInfo != nil {
								fmt.Fprintf(tw1, "%s\t", hs.Name)
								fmt.Fprintf(tw1, "%s\t", pnh.Device)
								fmt.Fprintf(tw1, "%s\t", cdpInfo.DevId)
								fmt.Fprintf(tw1, "%s\t", cdpInfo.PortId)
							}
							fmt.Fprintf(tw1, "\n")
						}
					}
				}
			}
		}
		fmt.Fprintf(tw, "\n")
	}
	_ = tw.Flush()
	return nil
}

func main() {
	ParseJSONInput()
	examples.Run(VmwareCreateAviSEFromOVF)
	//examples.Run(HostList)
}
