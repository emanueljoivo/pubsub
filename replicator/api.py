#!flask/bin/python
import os

from flask import Flask
import kubernetes as kube
from kubernetes.client.apis import core_v1_api

import random, string

def randomword(length):
   letters = string.ascii_lowercase
   return ''.join(random.choice(letters) for i in range(length))

app = Flask(__name__)

@app.route('/create/<int:n>')
def index(n):
    for i in range(n):
        name = randomword(10)
        print(name)
        core_v1 = core_v1_api.CoreV1Api()

        #service
        service_manifest = {
            "apiVersion":"v1",
            "kind":"Service",
            "metadata":
                {"name":name+"-service"},
            "spec":{
                "selector":{"app":name},
                "ports":[{"protocol":"TCP","port":8003}],
                "type":"LoadBalancer"}}

        service = core_v1.create_namespaced_service(namespace="default", body=service_manifest)
        ingress = service.status.load_balancer.ingress
        while (not ingress):
            service = core_v1.read_namespaced_service(namespace="default", name=name+"-service")
            ingress = service.status.load_balancer.ingress
        print(ingress)
        ip = ingress[0].ip
        #pod
        pod_manifest = {
            "apiVersion":"v1",
            "kind":"Pod",
            "metadata":{"name":name + "-storage",
                "labels":{"app":name}
            },
            "spec":{
                "containers":[
                    {"name":name,
                     "image":"ignacioschmid/pubsub:storage",
                     "ports":[{"containerPort":8003}],
                     "env":[{"name":"SENTINEL_HOST","value":"http://127.0.0.1"},
                            {"name":"SENTINEL_PORT","value":"8080"},
                            {"name":"SERVER_ADRESS","value":ip},
                            {"name":"SERVER_PORT","value":"8003"},
                            {"name":"ID","value":name}]
            }]}}

            core_v1.create_namespaced_pod(body=pod_manifest,namespace="default")
    
    return "OK"


def setup():
    conf_path = os.environ['KUBECONFIG']
    kube.config.load_kube_config(conf_path)

if __name__ == '__main__':
    setup()
    app.run(host='0.0.0.0',debug=False)