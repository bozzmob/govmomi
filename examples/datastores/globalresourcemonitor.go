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
	"context"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/simulator/esx"
	"github.com/vmware/govmomi/simulator/vpx"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func main() {
	examples.Run(func(ctx context.Context, c *vim25.Client) error {
		// Create a view of Network types
		m := view.NewManager(c)

		cluster_cv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
		datacenter_cv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datacenter"}, true)
		network_cv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
		host_cv, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		if err != nil {
			return err
		}

		defer v.Destroy(ctx)

		cluster := Map.Any("ClusterComputeResource")
		datacenter := Map.Any("Datacenter")
		network := Map.Any("Network")
		hostsystem :=  Map.Any("HostSystem")

		// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
		req := types.RetrieveProperties{
		SpecSet: []types.PropertyFilterSpec{
			{
				PropSet: []types.PropertySpec{
					{
						DynamicData: types.DynamicData{},
						Type:        "ClusterComputeResource",
						PathSet:     []string{"name"},
					},
				},
				ObjectSet: []types.ObjectSpec{
					{
						Obj: datacenter.Reference(),
					},
					{
						Obj: cluster.Reference(),
					},
					{
						Obj: network.Reference(),
					},
					{
						Obj: hostsystem.Reference(),
					},
				},
			},
		},
	}

	pc := client.PropertyCollector()

	res, err := pc.RetrieveProperties(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	content = res.Returnval

		return nil
	})
}
