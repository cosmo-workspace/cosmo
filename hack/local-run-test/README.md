# local test env for COSMO development

## Include in a single container


![overview](assets/test-env-1.dio.svg)


#### creat 
```
$ cd hack/local-run-test
$ make create-all
```

#### delete
```
$ make delete-all
```

#### See help for more information
```
$ make help
```

## Run the COSMO module outside of K8S

For repeating programming and testing.


![overview](assets/test-env-2.dio.svg)


#### creat 
```
$ cd hack/local-run-test
$ make create-all
$ make run-local
```

#### delete
```
$ make delete-all
```

#### See help for more information
```
$ make help
```