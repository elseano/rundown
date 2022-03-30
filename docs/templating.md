# Templating

Rundown allows for some basic templating.

## Example <r section="ex1"/>

Given this:

~~~ markdown
# Deploy to Kubernetes

<r save-as="k8s.yml" replace="<<name>>:Example"/>

``` yaml
pod:
  name: <<name>> Pod
```

We expect the `<<name>>` to be replaced in the saved file.

<r stdout/>

``` bash
cat $K8S_FILE
```
~~~

Then rundown will perform the replacement in the save file, resulting in:

~~~ expected
# Deploy to Kubernetes

We expect the <<name>> to be replaced in the saved file.

↓ Running...
    pod:
      name: Example Pod
✔ Running...
~~~