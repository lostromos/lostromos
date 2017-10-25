# Manual Testing

To run a functional test on your own to validate that the lostromos app is
working as expected you can go through the following steps.

1. Setup kubectl against a cluster (minikube works just fine)
2. `kubectl apply -f test-data/crd.yml`
3. `kubectl apply -f test-data/cr_things.yml`
4. `go run main.go start --config test-data/config.yaml`
    - See that it prints out that `thing1` and `thing2` were added
5. In another shell `kubectl apply -f test-data/cr_nemo.yml`
    - See that it prints out that `nemo` was added
6. `kubectl edit character nemo` and change a field
    - See that it prints out that `nemo` was changed
7. `kubectl delete -f test-data/cr_nemo.yml`
    - See that it prints out that `nemo` was deleted
8. `kubectl delete -f test-data/cr_things.yml`
    - See that it prints out that `nemo` was deleted
9. Thats it. You can stop the process and `kubectl delete -f test-data/crd.yml`
   to cleanup the rest of the test data.
