# k6 custom output extensions

# servers:

1. td_server: for receiving tdigest bytes (associated with trend metrics)
2. non_td_server: for receiving non-tdigest bytes for non-trend metrics

# xk6-output-tdigest directory:

1. Create a build by running the command:
   xk6 build --with xk6-output-tdigest=.
2. Run the k6 test file:
   ./k6 run test.js --out logger --quiet --no-summary --iterations 5
