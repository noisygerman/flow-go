// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package badger

import (
	"errors"
	"math/rand"

	"github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog"
)

type Cleaner struct {
	log   zerolog.Logger
	db    *badger.DB
	ratio float64
	freq  int
	calls int
}

func NewCleaner(log zerolog.Logger, db *badger.DB) *Cleaner {
	// NOTE: we run garbage collection frequently at points in our business
	// logic where we are likely to have a small breather in activity; it thus
	// makes sense to run garbage collection often, with a smaller ratio, rather
	// than running it rarely and having big rewrites at once
	c := &Cleaner{
		log:   log.With().Str("component", "cleaner").Logger(),
		db:    db,
		ratio: 0.2,
		freq:  1000,
	}
	// we don't want the entire network to run GC at the same time, so
	// distribute evenly over time
	c.calls = rand.Intn(c.freq)
	return c
}

func (c *Cleaner) RunGC() {

	// only actually run approximately every frequency number of calls
	c.calls++
	if c.calls < c.freq {
		return
	}

	// we add 20% jitter into the interval, so that we don't risk nodes syncing
	// up on their GC calls over time
	c.calls = rand.Intn(c.freq / 5)

	// run the garbage collection in own goroutine and handle sentinel errors
	go func() {
		err := c.db.RunValueLogGC(c.ratio)
		if errors.Is(err, badger.ErrRejected) {
			// NOTE: this happens when a GC call is already running
			c.log.Warn().Msg("garbage collection on value log already running")
			return
		}
		if errors.Is(err, badger.ErrNoRewrite) {
			// NOTE: this happens when no files have any garbage to drop
			c.log.Debug().Msg("garbage collection on value log unnecessary")
			return
		}
		if err != nil {
			c.log.Error().Err(err).Msg("garbage collection on value log failed")
			return
		}
		c.log.Debug().Msg("garbage collection on value log executed")
	}()
}
