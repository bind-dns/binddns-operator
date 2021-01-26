# binddns-operator

You can simply manages your DNS records with binddns-operator.

DnsDomain/DnsRule is the CRD of Kubernetes. Users can use them to change DNS records.

![show](docs/summary.gif)

## The Example

```
[root@localhost ~]# kubectl get dnsdomain
NAME             ENABLED   REMARK   UPDATE                
helloworld.com   true               2021-01-21 01:44:05
test.com         true               2021-01-20 21:38:09

[root@localhost ~]# kubectl get dnsrule
NAME                      ZONE             ENABLED   HOST   TYPE   DATA          TTL   MXPRIORITY
helloworld.com-b3378a6c   helloworld.com   true      www    A      1.1.1.1       10    0
test.com-8b223ed7         test.com         true      www    A      10.10.10.10   10    0

[root@localhost ~]# nslookup www.test.com 127.0.0.1
Server:		127.0.0.1
Address:	127.0.0.1#53

Name:	www.test.com
Address: 10.10.10.10
```

## How To Deploy

We need to deploy Controller and Webhook

### Deploy Controller

```
[root@localhost binddns-operator]# cd deploy/controller/

# Deploy DnsDomain CRD
[root@localhost controller]# kubectl apply -f crd_dnsdomains.yaml

# Deploy DnsRule CRD
[root@localhost controller]# kubectl apply -f crd_dnsrules.yaml

# Deploy RBAC
[root@localhost controller]# kubectl apply -f rbac.yaml

# Deploy ConfigMap
[root@localhost controller]# kubectl apply -f configmap.yaml

# Deploy Controller Deployment
[root@localhost controller]# kubectl apply -f deployment.yaml

```

### Deploy Webhook

```
[root@localhost binddns-operator]# cd deploy/webhook/

# Generate Secret
[root@localhost webhook]# ./webhook-create-signed-cert.sh --service binddns-webhook-svc --secret binddns-webhook-certs --namespace kube-system

# Generate Deployment
[root@localhost webhook]# cat mutatingwebhook.yaml | ./webhook-patch-ca-bundle.sh > mutatingwebhook-ca-bundle.yaml


# Deploy
[root@localhost webhook]# kubectl apply -f mutatingwebhook-ca-bundle.yaml
[root@localhost webhook]# kubectl apply -f service.yaml
[root@localhost webhook]# kubectl apply -f deployment.yaml
```

## Usage

- WebUI: [http://${IP}:5388/console/domains](http://${IP}:5388/console/domains)

- CRD: There is a demo at [deploy/demo/example.yaml](deploy/demo/example.yaml)


## Future

- The better WebUI 
- Dynamic rndc key
- Synchronous DnsDomain status
- ...

