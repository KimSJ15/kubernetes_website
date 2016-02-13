---
title: "Horizontal Pod Autoscaler"
---
Horizontal pod autoscaling is a [beta](/{{page.version}}/docs/api.html#api-versioning) feature in Kubernetes 1.1.
It allows the number of pods in a replication controller or deployment to scale automatically based on observed CPU usage.
In the future also other metrics will be supported.

In this document we explain how this feature works by walking you through an example of enabling horizontal pod autoscaling with the php-apache server.

## Prerequisites

This example requires a running Kubernetes cluster and kubectl in the version at least 1.1.
[Heapster](https://github.com/kubernetes/heapster) monitoring needs to be deployed in the cluster
as horizontal pod autoscaler uses it to collect metrics
(if you followed [getting started on GCE guide](/{{page.version}}/docs/getting-started-guides/gce),
heapster monitoring will be turned-on by default).


## Step One: Run & expose php-apache server

To demonstrate horizontal pod autoscaler we will use a custom docker image based on php-apache server.
The image can be found [here](https://releases.k8s.io/release-1.1/docs/user-guide/horizontal-pod-autoscaling/image).
It defines [index.php](image/index.php) page which performs some CPU intensive computations.

First, we will start a replication controller running the image and expose it as an external service:

<a name="kubectl-run"></a>

{% highlight console %}

$ kubectl run php-apache --image=gcr.io/google_containers/hpa-example --requests=cpu=200m
replicationcontroller "php-apache" created

$ kubectl expose rc php-apache --port=80 --type=LoadBalancer
service "php-apache" exposed

{% endhighlight %}

Now, we will wait some time and verify that both the replication controller and the service were correctly created and are running. We will also determine the IP address of the service:

{% highlight console %}

$ kubectl get pods
NAME               READY     STATUS    RESTARTS   AGE
php-apache-wa3t1   1/1       Running   0          12m

$ kubectl describe services php-apache | grep "LoadBalancer Ingress"
LoadBalancer Ingress:	146.148.24.244

{% endhighlight %}

We may now check that php-apache server works correctly by calling ``curl`` with the service's IP:

{% highlight console %}

$ curl http://146.148.24.244
OK!

{% endhighlight %}

Please notice that when exposing the service we assumed that our cluster runs on a provider which supports load balancers (e.g.: on GCE).
If load balancers are not supported (e.g.: on Vagrant), we can expose php-apache service as ``ClusterIP`` and connect to it using the proxy on the master:

{% highlight console %}

$ kubectl expose rc php-apache --port=80 --type=ClusterIP
service "php-apache" exposed

$ kubectl cluster-info | grep master
Kubernetes master is running at https://146.148.6.215

$ curl -k -u <admin>:<password> https://146.148.6.215/api/v1/proxy/namespaces/default/services/php-apache/
OK!

{% endhighlight %}


## Step Two: Create horizontal pod autoscaler

Now that the server is running, we will create a horizontal pod autoscaler for it.
To create it, we will use the [hpa-php-apache.yaml](hpa-php-apache.yaml) file, which looks like this:

{% highlight yaml %}

apiVersion: extensions/v1beta1
kind: HorizontalPodAutoscaler
metadata:
  name: php-apache
  namespace: default
spec:
  scaleRef:
    kind: ReplicationController
    name: php-apache
    namespace: default
  minReplicas: 1
  maxReplicas: 10
  cpuUtilization:
    targetPercentage: 50

{% endhighlight %}

This defines a horizontal pod autoscaler that maintains between 1 and 10 replicas of the Pods
controlled by the php-apache replication controller we created in the first step of these instructions.
Roughly speaking, the horizontal autoscaler will increase and decrease the number of replicas
(via the replication controller) so as to maintain an average CPU utilization across all Pods of 50%
(since each pod requests 200 milli-cores by [kubectl run](#kubectl-run), this means average CPU utilization of 100 milli-cores).
See [here](/{{page.version}}/docs/design/horizontal-pod-autoscaler.html#autoscaling-algorithm) for more details on the algorithm.

We will create the autoscaler by executing the following command:

{% highlight console %}

$ kubectl create -f docs/user-guide/horizontal-pod-autoscaling/hpa-php-apache.yaml
horizontalpodautoscaler "php-apache" created

{% endhighlight %}

Alternatively, we can create the autoscaler using [kubectl autoscale](../kubectl/kubectl_autoscale).
The following command will create the equivalent autoscaler as defined in the [hpa-php-apache.yaml](hpa-php-apache.yaml) file:

```

$ kubectl autoscale rc php-apache --cpu-percent=50 --min=1 --max=10
replicationcontroller "php-apache" autoscaled

```

We may check the current status of autoscaler by running:

{% highlight console %}

$ kubectl get hpa
NAME         REFERENCE                                   TARGET    CURRENT   MINPODS   MAXPODS   AGE
php-apache   ReplicationController/default/php-apache/   50%       0%        1         10        27s

{% endhighlight %}

Please note that the current CPU consumption is 0% as we are not sending any requests to the server
(the ``CURRENT`` column shows the average across all the pods controlled by the corresponding replication controller).

## Step Three: Increase load

Now, we will see how the autoscaler reacts on the increased load of the server.
We will start an infinite loop of queries to our server (please run it in a different terminal):

{% highlight console %}

$ while true; do curl http://146.148.6.244; done

{% endhighlight %}

We may examine, how CPU load was increased (the results should be visible after about 3-4 minutes) by executing:

{% highlight console %}

$ kubectl get hpa
NAME         REFERENCE                                   TARGET    CURRENT   MINPODS   MAXPODS   AGE
php-apache   ReplicationController/default/php-apache/   50%       305%      1         10        4m

{% endhighlight %}

In the case presented here, it bumped CPU consumption to 305% of the request.
As a result, the replication controller was resized to 7 replicas:

{% highlight console %}

$ kubectl get rc
CONTROLLER   CONTAINER(S)   IMAGE(S)                               SELECTOR         REPLICAS   AGE
php-apache   php-apache     gcr.io/google_containers/hpa-example   run=php-apache   7          18m

{% endhighlight %}

Now, we may increase the load even more by running yet another infinite loop of queries (in yet another terminal):

{% highlight console %}

$ while true; do curl http://146.148.6.244; done

{% endhighlight %}

In the case presented here, it increased the number of serving pods to 10:

{% highlight console %}

$ kubectl get hpa
NAME         REFERENCE                                   TARGET    CURRENT   MINPODS   MAXPODS   AGE
php-apache   ReplicationController/default/php-apache/   50%       65%       1         10        14m

$ kubectl get rc
CONTROLLER   CONTAINER(S)   IMAGE(S)                               SELECTOR         REPLICAS   AGE
php-apache   php-apache     gcr.io/google_containers/hpa-example   run=php-apache   10         24m

{% endhighlight %}

## Step Four: Stop load

We will finish our example by stopping the user load.
We will terminate both infinite ``while`` loops sending requests to the server and verify the result state:

{% highlight console %}

$ kubectl get hpa
NAME         REFERENCE                                   TARGET    CURRENT   MINPODS   MAXPODS   AGE
php-apache   ReplicationController/default/php-apache/   50%       0%        1         10        21m

$ kubectl get rc
CONTROLLER   CONTAINER(S)   IMAGE(S)                               SELECTOR         REPLICAS   AGE
php-apache   php-apache     gcr.io/google_containers/hpa-example   run=php-apache   1          31m

{% endhighlight %}

As we see, in the presented case CPU utilization dropped to 0, and the number of replicas dropped to 1.



