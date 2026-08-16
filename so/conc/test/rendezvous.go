package conc_test

import (
	"solod.dev/so/c"
	"solod.dev/so/conc"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// rdvTask carries a rendezvous and the number of points to hand off.
type rdvTask struct {
	rdv *conc.Rendezvous
	n   int
}

// sendPoints hands off n points and then closes the rendezvous.
func sendPoints(arg any) any {
	task := arg.(*rdvTask)
	for i := 0; i < task.n; i++ {
		p := point{x: i, y: i * 2}
		task.rdv.Send(&p)
	}
	task.rdv.Close()
	return nil
}

func TestRendezvous_Handoff(t *testing.T) {
	task := rdvTask{rdv: conc.NewRendezvous(t.Allocator(), c.Sizeof[point]()), n: 10}
	defer task.rdv.Free()

	thr := conc.Go(sendPoints, &task)
	count := 0
	var got point
	for task.rdv.Recv(&got) {
		if got.x != count || got.y != count*2 {
			t.Fatalf("Recv = {%d, %d}, want {%d, %d}",
				got.x, got.y, count, count*2)
			return
		}
		count++
	}
	thr.Wait()

	if count != 10 {
		t.Errorf("received %d points, want 10", count)
	}
}

func TestRendezvous_CloseEmpty(t *testing.T) {
	// Check that a receive on a closed rendezvous reports the channel closed.
	// A rendezvous holds nothing, so there is nothing to drain after Close.
	rdv := conc.NewRendezvous(t.Allocator(), c.Sizeof[point]())
	defer rdv.Free()
	rdv.Close()

	var got point
	if rdv.Recv(&got) {
		t.Error("Recv on a closed rendezvous = true, want false")
	}
	if rdv.RecvTimeout(&got, 0) != conc.Closed {
		t.Error("RecvTimeout on a closed rendezvous, want Closed")
	}
	p := point{x: 1, y: 2}
	if rdv.SendTimeout(&p, 0) != conc.Closed {
		t.Error("SendTimeout on a closed rendezvous, want Closed")
	}
}

func TestRendezvous_Timeout(t *testing.T) {
	rdv := conc.NewRendezvous(t.Allocator(), c.Sizeof[point]())
	defer rdv.Free()

	var got point
	if rdv.RecvTimeout(&got, 0) != conc.Timeout {
		t.Error("RecvTimeout with no sender, want Timeout")
	}
	p := point{x: 1, y: 2}
	if rdv.SendTimeout(&p, 10*time.Millisecond) != conc.Timeout {
		t.Error("SendTimeout with no receiver, want Timeout")
	}
}
