package main

import (
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	totalAccounts  = 50000
	maxAmountMoved = 10
	initialMOney   = 100
	threads        = 4
)

func perform_movement(ledger *[totalAccounts]int32, locks *[totalAccounts]sync.Locker, totalTrans *int64) {
	// we lock both the accounts which participate in a transaction
	// lock array represents lock for each account in our array
	for {
		accountA := rand.Intn(totalAccounts)
		accountB := rand.Intn(totalAccounts)
		for accountA == accountB { // if we randomly select 2 same accounts
			accountB = rand.Intn(totalAccounts)
		}
		amountToMove := rand.Int31n(maxAmountMoved) // way to choes random 32 bit integer
		toLock := []int{accountA, accountB}
		sort.Ints(toLock) // we sort the 2 accounts by their id, so that we lock the right account first and not the other account. Lock the lowest in order by id account first, avoid deadlock
		locks[toLock[0]].Lock()
		locks[toLock[1]].Lock()

		atomic.AddInt32(&ledger[accountA], -amountToMove)
		atomic.AddInt32(&ledger[accountB], amountToMove)
		atomic.AddInt64(totalTrans, 1)

		locks[toLock[1]].Unlock()
		locks[toLock[0]].Unlock()
	} // without locks, we will have race condition and total amount will vary across the system
}

func main() {
	println("total accounts", totalAccounts, "total threads", threads, "using spinlocks")
	var ledger [totalAccounts]int32
	var locks [totalAccounts]sync.Locker
	var totalTrans int64

	for i := 0; i < totalAccounts; i++ {
		ledger[i] = initialMOney
		locks[i] = NewSpinLock()
	}

	for i := 0; i < threads; i++ {
		go perform_movement(&ledger, &locks, &totalTrans)
	}

	// to check if our account is consistent or not
	for {
		time.Sleep(2000 * time.Millisecond)
		var sum int32
		// now we lock all the account, sum them up and unlock them \
		for i := 0; i < totalAccounts; i++ {
			locks[i].Lock() // lock every account
		}
		for i := 0; i < totalAccounts; i++ {
			sum += ledger[i]
		}
		for i := 0; i < totalAccounts; i++ {
			locks[i].Unlock()
		}
		println(totalTrans, sum)
	}

}
