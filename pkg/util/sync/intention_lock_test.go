package sync

import (
	"sync"
	"testing"
)

func TestIntentionLock_Exclusive(t *testing.T) {
	var l IntentionLock
	unlocker, ok := l.TryLock(LockGranularityExclusive)
	if !ok {
		t.Fatal("expected to be able to lock fresh IntentionLock exclusively, but returned nil")
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, ok := l.TryLock(LockGranularityExclusive)
			if ok {
				t.Fatal("expected to not be able to lock exclusively locked lock exclusively, but was able to")
			}
		}()
	}

	wg.Wait()

	var wg2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()

			_, ok := l.TryLock(LockGranularityIntentionExclusive)
			if ok {
				t.Fatal("expected to not be able to lock exclusively locked lock exclusively, but was able to")
			}
		}()
	}

	wg2.Wait()

	unlocker.Unlock()

	_, ok = l.TryLock(LockGranularityExclusive)
	if !ok {
		t.Fatal("expected to be able to re-lock lock exclusively")
	}
}

func TestIntentionLock_IntentionExclusive(t *testing.T) {
	var l IntentionLock
	unlocker, ok := l.TryLock(LockGranularityIntentionExclusive)
	if !ok {
		t.Fatal("expected to be able to lock fresh IntentionLock with intention, but returned nil")
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, ok := l.TryLock(LockGranularityExclusive)
			if ok {
				t.Fatal("expected to not be able to lock exclusively locked lock exclusively, but was able to")
			}
		}()
	}

	wg.Wait()

	var acquiredWg sync.WaitGroup
	var doneWg sync.WaitGroup
	numGR := 10
	acquiredWg.Add(numGR)
	doneWg.Add(numGR)
	for i := 0; i < numGR; i++ {
		go func() {
			defer doneWg.Done()

			u, ok := l.TryLock(LockGranularityIntentionExclusive)
			if !ok {
				t.Fatal("expected to be able to lock exclusively locked lock intention exclusively, but was not able to")
			}

			acquiredWg.Done()
			acquiredWg.Wait()
			u.Unlock()
		}()
	}

	doneWg.Wait()

	unlocker.Unlock()

	_, ok = l.TryLock(LockGranularityExclusive)
	if !ok {
		t.Fatal("expected to be able to re-lock lock exclusively")
	}
}
