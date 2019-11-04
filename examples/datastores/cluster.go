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
	"fmt"
	"log"

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

// METHOD
// func getDataCenter(ctx context.Context, c *vim25.Client, me mo.ManagedEntity) mo.Datacenter {
// 	// c, err := govmomi.NewClient(ctx)
// 	path, err := mo.Ancestors(ctx, c, c.ServiceContent.PropertyCollector, me.Reference())
// 	if err != nil {
// 		fmt.Printf("ERROR : %s | \n", err)
// 	}
// 	for i := range path {
// 		if path[i].Reference().Type == "Datacenter" {
// 			// log.Printf("Managed Entity Reference=%s DC=%s\n", mor.Reference(), path[i].Name)
// 			return path[i]
// 		}
// 	}
// 	return
// }

func main() {
	examples.Run(func(ctx context.Context, c *vim25.Client) error {

		// Create a view of Network types
		m := view.NewManager(c)

		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
		if err != nil {
			return err
		}

		defer v.Destroy(ctx)

		// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
		var clusters []mo.ClusterComputeResource
		err = v.Retrieve(ctx, []string{"ClusterComputeResource"}, nil, &clusters)
		if err != nil {
			return err
		}

		// fmt.Println(clusters)
		pc := property.DefaultCollector(c)

		for _, cl := range clusters {
			path, err := mo.Ancestors(ctx, c, c.ServiceContent.PropertyCollector, cl.Reference())
			if err != nil {
				fmt.Printf("ERROR : %s | \n", err)
			}
			for i := range path {
				if path[i].Reference().Type == "Datacenter" {
					log.Printf("Managed Entity Reference=%s DC=%s\n", cl.Reference(), path[i].Name)
				}
			}
			// fmt.Printf("DATACENTER %s | \n ", getDataCenter(ctx, c, cl))
			fmt.Printf("Name: %s |\n", cl.Name)
			// fmt.Printf("Datastore: %s \n| ", cl.Datastore)
			fmt.Printf("Host: %s | \n", cl.Host)
			fmt.Printf("Parent: %s | \n", cl.Parent)
			fmt.Printf("Managed Entity Reference: %s | \n", cl.GetManagedEntity().Reference().Value)
			fmt.Printf("OverallStatus: %s | \n", cl.Summary.GetComputeResourceSummary().OverallStatus)
			fmt.Printf("NumEffectiveHosts: %s | \n", cl.Summary.GetComputeResourceSummary().NumEffectiveHosts)

			for _, vmMor := range cl.Host {
				var hs mo.HostSystem
				err = pc.RetrieveOne(ctx, vmMor.Reference(), []string{"name", "config", "runtime"}, &hs)
				if err != nil {
					continue
				}
				fmt.Printf("HOSTS: %s \n", hs.Name)
			}
			fmt.Printf("\n--------------------------------\n")
		}

		return nil
	})
}
