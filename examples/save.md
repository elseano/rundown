# Saving content

Sometimes you want to show some content, and be able to refer to it later. This might be configuration you need to apply, or commands you need to pass to an executable. The `save:` label allows you to do just that!

    ``` json save:content
    { "key": "value" }
    ```

This will write the contents into a temporary file, which you can then reference later as an environment variable.

``` yaml save:content reveal
apiVersion: v1
kind: Pod
metadata:
  name: static-web
  labels:
    role: myrole
spec:
  containers:
    - name: web
      image: nginx
      ports:
        - name: web
          containerPort: 80
          protocol: TCP
```

Because we used the label's value as `content`, the file path is available under `$CONTENT`.

``` bash stdout reveal
echo "The filename is: $CONTENT"
echo
cat $CONTENT
```

## Changing the file extension

Some tools will detect the file type based on the extension. Rundown will append the save label value to the generated filename, but it's smart enough to ignore the extension for the variable.

    ``` json save:content.json
    { "key": "value" }
    ```

``` json save:content.json
{ "key": "value" }
```

Later, you can still refer to is as just `$CONTENT`.

``` bash reveal stdout
echo $CONTENT
```

## Injecting data

Sometimes you want to inject data into your content. The `env_aware` modifier will substitute all `$VAR` references with the actual environment value.

``` json save:content2.json env_aware reveal
{ "lastContentFile": "$CONTENT" }
```

``` bash reveal stdout
cat $CONTENT2
```