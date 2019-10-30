package main

import (
	"context"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/examples"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"

	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

func main() {
	examples.Run(func(ctx context.Context, client *vim25.Client) error {
		// ctx := context.Background()

		m := view.NewManager(client)
		err := m.Create()
		if err != nil {
			t.Fatal(err)
		}

		s := m.Service.NewServer()
		defer s.Close()

		c, err := govmomi.NewClient(ctx, s.URL, true)
		if err != nil {
			t.Fatal(err)
		}

		wait := make(chan bool)
		pc := property.DefaultCollector(c.Client)
		obj := Map.Any("VirtualMachine").(*VirtualMachine)
		ref := obj.Reference()
		vm := object.NewVirtualMachine(c.Client, ref)
		filter := new(property.WaitFilter).Add(ref, ref.Type, []string{"runtime.powerState"})
		// WaitOptions.maxWaitSeconds:
		// A value of 0 causes WaitForUpdatesEx to do one update calculation and return any results.
		filter.Options = &types.WaitOptions{
			MaxWaitSeconds: types.NewInt32(0),
		}
		// toggle power state to generate updates
		state := map[types.VirtualMachinePowerState]func(context.Context) (*object.Task, error){
			types.VirtualMachinePowerStatePoweredOff: vm.PowerOn,
			types.VirtualMachinePowerStatePoweredOn:  vm.PowerOff,
		}

		err = pc.CreateFilter(ctx, filter.CreateFilter)
		if err != nil {
			t.Fatal(err)
		}

		req := types.WaitForUpdatesEx{
			This:    pc.Reference(),
			Options: filter.Options,
		}

		go func() {
			for {
				res, err := methods.WaitForUpdatesEx(ctx, c.Client, &req)
				if err != nil {
					if ctx.Err() == context.Canceled {
						werr := pc.CancelWaitForUpdates(context.Background())
						t.Error(werr)
						return
					}
					t.Error(err)
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

		// wait for the enter update.
		<-wait

		// Now change the VM power state, to generate a modify update
		_, err = state[obj.Runtime.PowerState](ctx)
		if err != nil {
			t.Error(err)
		}

		// wait for the modify update.
		<-wait
		return nil
	})
}
