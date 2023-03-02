# k6 custom output extensions

# servers:

1. td_server: for receiving tdigest bytes (associated with trend metrics)
2. non_td_server: for receiving non-tdigest bytes for non-trend metrics

# xk6-output-tdigest directory:

1. Create a build by running the command:
   xk6 build --with xk6-output-tdigest=.
2. Run the k6 test file:
   ./k6 run test.js --out logger --quiet --no-summary --iterations 5

# Outstanding ToDos:

## tdigest_output_v3.go:

- [] flush-align/synchronize to rounded 10 second intervals (10 secs is arbitrary, can change to different interval)
  the call to flushMetrics
  - currently using PeriodicFlusher to conduct a repetitive task at intervals
  - I think we will want to use go's Ticker to do this instead
- [] refactor how connection is created 
   - wanted to add a Conn type to the Logger struct that keeps track of the connections to servers, but that type
     was undefined, so the simple fix was to inline eastablish a new connection every time flushMetrics is called,
     which is unideal
   - server is a SPOF - how are we building in resiliency
- [] track the start time of each rounded/flush-aligned interval (ie 1 min 20 secs, 1 min 30 secs, 1 min 40 secs)
- [] send start time along with count data to relevant server
  - not sure if there will be any issues in maintaing integrity of a time stamp when convert it to bytes along with
    metric name and metric value and pass all this info to the server
- [] need to figure out a way to pass the tdigest []byte, tdigest metric name, and timestamp as one message to the server
  - this bullet point assumes that putting all this info into one []byte will be problematic when trying to deserialize
    the components and maintain the integrity of each
- [] add logic for other Metric.Types
  - right now only have logic for handling counts and trends

## td_server:

- [] need to appropriately deserialize tdigests
  - need to figure out how to receive "buf" type appropriately
    - buf type is the necessary input to FromBytes function for recreating a tdigest that's sent as []bytes
- [] aggregate tdigests by metric name for each rounded time interval
- [] when rounded interval ends:
  - calculate the 50/90/95/99 interval for each unique metric's associated aggregated/merged tdigest
  - write these calculated values along with the timestamp of the start of the rounded interval, and
  - timestamp (start of rounded interval)
- [] reset running total tdigest when each rounded interval ends
  - do we need a buffer for certain messages that could potentially come in to the server while resetting the tdigest?
  - instead of resetting tdigests may be better to create unique tdigests associated with timestamp of the start of the
    relevant rounded time interval and metric name
    - store in map
    - once have reassurance that tdigest for metric/timestamp has been written to database can delete to ensure
      map doesn't get too large

## non_td_server:

- same high-level todos as td_server
  - processes that will look different b/c not working with tdigests:
   - deserialization
   - aggregation
   - map storage
