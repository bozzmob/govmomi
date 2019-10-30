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

package main

import (
	"context"
	"time"
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/object"
	// "github.com/vmware/govmomi/simulator"

)

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
	examples.Run(func(ctx context.Context, client *vim25.Client) error {
		// Create a view of Network types
		m := view.NewManager(client)

		// clusterCv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
		// datacenterCv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datacenter"}, true)
		// networkCv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
		hostCv, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		if err != nil {
			return err
		}

		defer hostCv.Destroy(ctx)

		// cluster := Map.Any("ClusterComputeResource")
		// datacenter := Map.Any("Datacenter")
		// network := Map.Any("Network")
		// hostsystem := simulator.Map.Any("HostSystem")
		// hostsystem := 

		// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
		// reqtype := types.RetrieveProperties{
		// 	SpecSet: []types.PropertyFilterSpec{
		// 		{
		// 			PropSet: []types.PropertySpec{
		// 				{
		// 					DynamicData: types.DynamicData{},
		// 					Type:        "HostSystem",
		// 					PathSet:     []string{"name"},
		// 				},
		// 			},
		// 			ObjectSet: []types.ObjectSpec{
		// 				// {
		// 				// 	Obj: datacenter.Reference(),
		// 				// },
		// 				// {
		// 				// 	Obj: cluster.Reference(),
		// 				// },
		// 				// {
		// 				// 	Obj: network.Reference(),
		// 				// },
		// 				{
		// 					Obj: hostsystem.Reference(),
		// 				},
		// 			},
		// 		},
		// 	},
		// }

		// sipc := client.ServiceContent.PropertyCollector
		// si := new(NewServiceInstance(client.ServiceContent, folder))
		// var serviceInstance = types.ManagedObjectReference{
		// 	Type:  "ServiceInstance",
		// 	Value: "ServiceInstance",
		// }

		// sipc := serviceInstance.PropertyCollector

		// sipc := new(PropertyCollector)
		// res, err := pc.RetrieveProperties(ctx, req)
		// if err != nil {
		// 	println(err)
		// }

		pc := property.DefaultCollector(client)
		// obj := simulator.Map.Any("HostSystem")
		// ref := obj.Reference()

		filter := new(property.WaitFilter)
		// WaitOptions.maxWaitSeconds:
		// A value of 0 causes WaitForUpdatesEx to do one update calculation and return any results.
		filter.Options = &types.WaitOptions{
			MaxWaitSeconds: types.NewInt32(0),
		}

		err = pc.CreateFilter(ctx, filter.CreateFilter)
		if err != nil {
			println(err)
		}

		req := types.WaitForUpdatesEx{
			This:    pc.Reference(),
			Options: filter.Options,
		}

		wait := make(chan bool)

		go func() {
			for {
				res, err := methods.WaitForUpdatesEx(ctx, client, &req)
				if err != nil {
					if ctx.Err() == context.Canceled {
						// werr := pc.CancelWaitForUpdates(context.Background())
						werr := pc.CancelWaitForUpdates(context.Background())
						println(werr)
						return
					}
					println(err)
					return
				}

				set := res.Returnval
				if set == nil {
					// Retry if the result came back empty
					// That's a normal case when MaxWaitSeconds is set to 0.
					// It means we have no updates for now
					time.Sleep(500 * time.Millisecond)
					continue
				}

				req.Version = set.Version

				for _, fs := range set.FilterSet {
					// We expect the enter of VM first
					if fs.ObjectSet[0].Kind == types.ObjectUpdateKindEnter {
						wait <- true
						// Keep going
						continue
					}

					// We also expect a modify due to the power state change
					if fs.ObjectSet[0].Kind == types.ObjectUpdateKindModify {
						wait <- true
						// Now we can return to stop the routine
						return
					}
				}
			}
		}()
		<-wait

		// Now change the VM power state, to generate a modify update
		err = HostList(ctx,client)
		if err != nil {
			println(err)
		}

		// wait for the modify update.
		<-wait
		println("waited for HostList")

		return nil

		// content = res.Returnval

		return nil
	})
}
