# Human Kubernetes Operator

This is yet another demo/_hello, world_ Kubernetes Operator.

## Goals

The operator is supposed to operate on the following example [`Custom Resource(CR)`](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
```yaml
apiVersion: mammals.earth.com/v1
kind: Human
metadata:
    name: Harry
    namespace: test
spec:
    hands: 2
    feet: 2
    mothertongue: English
    tail: 0
```

It is then supposed to do
- [ ] Creates a container which outputs to STDOUT - `{{ name }} has {{ feet }} feet, {{ hand }} hands and {{ tail }} tails. Also, {{ name }} speaks in {{ mothertongue }}`
- [ ] Creates a ConfigMap with data containing all the elements of the `.spec` of the Human CR