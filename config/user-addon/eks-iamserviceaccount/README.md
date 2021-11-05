# eks-iamserviceaccount

For use AWS CodeCommit and AWS CodeArtifact, 
Attach IAM Role on service account on User namespace.

```sh
make install EKS_CLUSTER_NAME=<YOUR_CLUSTER_NAME>
```

You can modify the policies in `job.yaml`