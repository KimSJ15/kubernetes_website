---
title: "Secrets example"
---
Following this example, you will create a [secret](../secrets) and a [pod](../pods) that consumes that secret in a [volume](../volumes). See [Secrets design document](../../design/secrets) for more information.

## Step Zero: Prerequisites

This example assumes you have a Kubernetes cluster installed and running, and that you have
installed the `kubectl` command line tool somewhere in your path. Please see the [getting
started](/{{page.version}}/docs/getting-started-guides/) for installation instructions for your platform.

## Step One: Create the secret

A secret contains a set of named byte arrays.

Use the [`examples/secrets/secret.yaml`](secret.yaml) file to create a secret:

{% highlight console %}

$ kubectl create -f docs/user-guide/secrets/secret.yaml

{% endhighlight %}

You can use `kubectl` to see information about the secret:

{% highlight console %}

$ kubectl get secrets
NAME          TYPE      DATA
test-secret   Opaque    2

$ kubectl describe secret test-secret
Name:          test-secret
Labels:        <none>
Annotations:   <none>

Type:   Opaque

Data
====
data-1: 9 bytes
data-2: 11 bytes

{% endhighlight %}

## Step Two: Create a pod that consumes a secret

Pods consume secrets in volumes.  Now that you have created a secret, you can create a pod that
consumes it.

Use the [`examples/secrets/secret-pod.yaml`](secret-pod.yaml) file to create a Pod that consumes the secret.

{% highlight console %}

$ kubectl create -f docs/user-guide/secrets/secret-pod.yaml

{% endhighlight %}

This pod runs a binary that displays the content of one of the pieces of secret data in the secret
volume:

{% highlight console %}

$ kubectl logs secret-test-pod
2015-04-29T21:17:24.712206409Z content of file "/etc/secret-volume/data-1": value-1

{% endhighlight %}



