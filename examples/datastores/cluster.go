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

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

// func getDC(en ManagedEntity) Datacenter {
// 	if (en != null) {
// 		// ManagedEntity parent = en.Parent
// 		if (parent != null)
// 		{
// 			if (parent.getMOR().getType().equals("Datacenter")) {
// 					return (Datacenter) parent;
// 			} else {
// 					return getDC(en.getParent());
// 			}
// 		}
// 	}
// 	return null;
// }

// datacenterPath returns the absolute path to the Datacenter containing the given ref
// func (f *Finder) getDatacenter(ctx context.Context, ref types.ManagedObjectReference) (*object.Datacenter, error) {
// 	mes, err := mo.Ancestors(ctx, f.client, f.client.ServiceContent.PropertyCollector, ref)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Chop leaves under the Datacenter
// 	for i := len(mes) - 1; i > 0; i-- {
// 		if mes[i].Self.Type == "Datacenter" {
// 			break
// 			// return mes[i]
// 		}
// 		mes = mes[:i]
// 	}

// 	// var p string

// 	// for _, me := range mes {
// 	// 	// Skip root entity in building inventory path.
// 	// 	if me.Parent == nil {
// 	// 		continue
// 	// 	}

// 	// 	p = p + "/" + me.Name
// 	// }

// 	return mes, nil
// }

// func getDataCenter(me mo.ManagedEntity) *object.Datacenter {
// 	// configs := []struct {
// 	// 	folder  mo.Folder
// 	// 	content types.ServiceContent
// 	// 	dc      *types.ManagedObjectReference
// 	// }
// 	parentManagedEntity := me.Parent

// 	fmt.Printf("ManagedEntity me ==> %s \n", me)
// 	fmt.Printf("parentManagedEntity me ==> %s \n", parentManagedEntity)
// 	if parentManagedEntity != nil {
// 		if parentManagedEntity.Type == "Datacenter" {
// 			fmt.Printf("parentManagedEntity.Type : %s | \n", parentManagedEntity.Type)
// 			return mo.Datacenter(parentManagedEntity)
// 		}
// 		fmt.Printf("parentManagedEntity.Type : %s | \n", parentManagedEntity.Type)
// 		return getDataCenter(parentManagedEntity)
// 	}
// 	return mo.Datacenter(parentManagedEntity)
// 	// es, err = me.Ancestors(ctx, c *vim25.Client, config.content.PropertyCollector, dc.Reference())
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
			// var group mo.Folder
			// fmt.Printf("DATACENTER %s | \n ", getDataCenter(cl.GetManagedEntity()))
			me := cl.GetManagedEntity().Parent

			fmt.Printf("DATACENTER %s | \n ", me)
			fmt.Printf("Name: %s |\n", cl.Name)
			// fmt.Printf("Datastore: %s \n| ", cl.Datastore)
			fmt.Printf("Host: %s | \n", cl.Host)
			fmt.Printf("Parent: %s | \n", cl.Parent)
			fmt.Printf("Managed Entity Reference: %s | \n", cl.GetManagedEntity().Reference().Value)
			fmt.Printf("OverallStatus: %s | \n", cl.Summary.GetComputeResourceSummary().OverallStatus)
			fmt.Printf("NumEffectiveHosts: %s | \n", cl.Summary.GetComputeResourceSummary().NumEffectiveHosts)
			// var dc mo.Datacenter
			// pc.RetrieveOne(ctx, cl.Reference(), []string{"name"}, &dc)
			// fmt.Printf("DataCenter %s \n ", dc.Name)

			// fmt.Printf("GetComputeResourceSummary: %s | \n", cl.Summary.GetComputeResourceSummary())
			// fmt.Printf("ManagedEntity: %s | \n", cl.ManagedEntity)
			// fmt.Printf("Host: %s | \n", cl.Host)
			// Datacenter ccr_dc []mo.Datacenter
			// ccr_dc =

			// for _, vmMor := range cl.Datacenter {
			// 	var dc mo.Datacenter
			// 	err = pc.RetrieveOne(ctx, vmMor.Reference(), []string{"name", "config", "runtime"}, &dc)
			// 	if err != nil {
			// 		continue
			// 	}
			// 	fmt.Printf("HOSTS: %s \n", dc.Name)
			// }

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
