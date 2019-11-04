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

	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func main() {
	examples.Run(func(ctx context.Context, client *vim25.Client) error {
		// Create a view of Network types
		m := view.NewManager(client)

		hostCv, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		if err != nil {
			return err
		}

		defer hostCv.Destroy(ctx)

		// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
		pc := property.DefaultCollector(client)
		filter := types.CreateFilter{
			This: pc.Reference(),
			Spec: types.PropertyFilterSpec{
				PropSet: []types.PropertySpec{
					{
						DynamicData: types.DynamicData{},
						Type:        "HostSystem",
						PathSet:     []string{"name", "vm", "hardware", "network", "config.network", "runtime"},
					},
				},
				ObjectSet: []types.ObjectSpec{
					{
						Obj:  client.ServiceContent.RootFolder,
						Skip: types.NewBool(false),
						SelectSet: []types.BaseSelectionSpec{
							&types.TraversalSpec{
								SelectionSpec: types.SelectionSpec{
									Name: "traverseFolders",
								},
								Type: "Folder",
								Path: "childEntity",
								Skip: types.NewBool(true),
								SelectSet: []types.BaseSelectionSpec{
									&types.TraversalSpec{
										Type: "HostSystem",
										Path: "vm",
										Skip: types.NewBool(false),
									},
									&types.SelectionSpec{
										Name: "traverseFolders",
									},
								},
							},
						},
					},
				},
				ReportMissingObjectsInResults: (*bool)(nil),
			},
			PartialUpdates: true,
		}

		if err = pc.CreateFilter(ctx, filter); err != nil {
			println(err)
		}

		updates, _ := pc.WaitForUpdates(ctx, "") // disregard first result
		updates, _ = pc.WaitForUpdates(ctx, updates.Version)
		println(updates.FilterSet[0].ObjectSet[0])

		return nil
	})
}
