package redis

import "testing"

func TestCommunication(t *testing.T) {
	// assuming redis default settings
	n, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	ch, err := n.BeatmapRequestsChannel()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		err := n.SendBeatmapRequest(i)
		if err != nil {
			t.Fatal(err)
		}
		x := <-ch
		if x != i {
			t.Fatal("got", x, "want", i)
		}
	}
}
