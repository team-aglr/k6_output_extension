# k6 custom output extensions

# create a build:

1. you can't have both >1 output extensions in the same folder while creating a build (so right now would have to move either tdigest_output_v2.go or tdigest_output_v3.go elsewhere)
2. run the command:
   xk6 build --with xk6-output-tdigest=.

# Run the test file after creating a build with the command:

./k6 run test.js --out logger --quiet --no-summary --iterations 5
