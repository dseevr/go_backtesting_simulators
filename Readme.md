# Go simulators

These 216 commits were written between 2014-10-04 and 2014-12-12.

These were my first attempts at building a financial market backtesting simulator from scratch.  I had no idea what I was doing, so I kind of architected myself into a corner and decided to rewrite it in a better way in a faster language (Rust).

There's a LOT of inefficient code that was never optimized, e.g., needless list traversals which scale really poorly as the number of trades in a session increase.

### exchange_simulator

This was my first stab a backtesting simulator.  I ran into a lot of design problems because I really had no idea of the scope of what I was trying to build.  Go's type system didn't really help.

### walk\_forward\_analaysis

This was a port of the exchange_simulator to implement Walk Forward Optimization/Analysis as detailed by Robert Pardo in "The Evaluation and Optimization of Trading Strategies".  I abandoned this in favor of redoing the simulator from scratch in Rust.  See the "rust\_simulator" repo.

### ga

These are a few random things I was testing when evaluating the idea of using Genetic Algorithm-based trading strategies.

### nn

These are a few utilities around training Neural Networks using **libfann**.  I wanted to build some NN-backed indicators which trading strategies could use, but I never got that far with the simulator.

## Usage

As of my last time touching this code (early Dec 2014), `go build` is sufficient for building it.  It doesn't use anything outside the stdlib except one of my own libraries on Github.

## License

BSD
