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

def get_node_cluster(k8s_conf_path):
    """ Gets the IP address of one slave node contained
    in a Kubernetes cluster. The k8s API aways returns information
    about the master node followed by the information of the slaves.
    Therefore, in order to avoid get the IP of the master node,
    this function always get the last node listed by the API.
    Raises:
        Exception -- It was not possible to connect with the
        Kubernetes cluster.
    Returns:
        string -- The node IP
    """
    try:
        kube.config.load_kube_config(k8s_conf_path)
        CoreV1Api = kube.client.CoreV1Api()
        for node in CoreV1Api.list_node().items:
            is_ready = \
                [s for s in node.status.conditions
                 if s.type == 'Ready'][0].status == 'True'
            if is_ready:
                node_info = node
        node_ip = node_info.status.addresses[0].address
        return node_ip
    except Exception:
        API_LOG.log("Connection with the cluster %s \
                    was not successful" % k8s_conf_path)
                    
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
                "type":"NodePort"}}

        service = core_v1.create_namespaced_service(namespace="default", body=service_manifest)
        port = service.spec.ports[0].node_port
        ip = get_node_cluster(os.environ['KUBECONFIG'])
        print(ip + ':' + str(port))
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
                     "image":"ignacioschmid/pubsub:storage_test",
                     "ports":[{"containerPort":8003}],
                     "env":[{"name":"SENTINEL_HOST","value":"http://192.168.25.68"},
                            {"name":"SENTINEL_PORT","value":"8080"},
                            {"name":"SERVER_ADDRESS","value":ip},
                            {"name":"SERVER_PORT","value":str(port)},
                            {"name":"ID","value":name}]
            }]}}

        pod = core_v1.create_namespaced_pod(body=pod_manifest,namespace="default")
    
    return ip + ':' + str(port)


def setup():
    conf_path = os.environ['KUBECONFIG']
    kube.config.load_kube_config(conf_path)

if __name__ == '__main__':
    setup()
    app.run(host='0.0.0.0',debug=False)