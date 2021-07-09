#!/bin/python
"""This script starts one demo microservice system along with its support 
services. Then it will start an incident simulation process.

Here is the time schedule of the simulation:
0   -   1m      Build and start the microservice system according to the config.
1m  -   1m      Start the load test.
1m  -   6m      Generate 5m normal metric data.
6m  -   6m      Inject an incident (an error).
6m  -   11m     Generate 5m abnormal metric data.
11m -   11m     Stop the simulation. Exits.
"""
import argparse
import datetime
import sys
import os
import subprocess
import time
from subprocess import TimeoutExpired, CalledProcessError
import pickle
import json
import sys

from py_lib.get_prom_data import query_prom_data_range, get_query1

def get_svc_names(config_file):
    svc_names = []
    with open(config_file, "rt") as f:
        obj = json.loads(f.read())
        for svc in obj:
            svc_names.append(svc['Name'].lower())
    return svc_names

def run_cmds(args, target, cwd=None, timeout=None, env=None):
    """Run shell cmds `args` in subprocess. `target` describes the task.
    `cwd` is the working directory to run the cmds.
    """
    try:
        # proc = subprocess.run(
        #     args, encoding='utf-8', check=True, shell=True, cwd=cwd, timeout=timeout, env=env)
        p = subprocess.Popen(args, encoding='utf-8', shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            cwd=cwd, env=env)
        returncode = p.poll()
        while returncode is None:
            line = p.stdout.readline()
            returncode = p.poll()
            line = line.strip()
            sys.stdout.write(line)
        print(f"{target} returned with {returncode}")
    except TimeoutExpired as e:
        p.terminate()
        print(f"{target} timeout.")
        exit(1)
    except CalledProcessError as e:
        p.terminate()
        print(f"{target} failed with error:", e)
        exit(1)
    except KeyboardInterrupt as e:
        p.terminate()
        print(f"{target} is interrupted.")
        exit(1)



if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Demo1 microservice system incident simulation")
    parser.add_argument("--registry", type=str, default="vmhost3.local")
    parser.add_argument("--network", type=str, default="my_network")
    parser.add_argument("--sshhost", type=str, default="pyc@vmhost1.local")
    parser.add_argument("--sshpw", type=str, default="pyc5279101")
    parser.add_argument("--rootdir", type=str, default=os.getcwd())
    parser.add_argument("--pipeline", type=str, nargs="*", default="all")
    parser.add_argument("--expname", type=str, default="demo1")
    parser.add_argument("--verbose", type=bool, default=True, dest='verbose')
    args = parser.parse_args()

    # Helper function to check the pipeline position.
    def check_pipeline(pipeline):
        if args.pipeline == 'all':
            return True
        if isinstance(args.pipeline, list) and pipeline in args.pipeline:
            return True
        return False

    if args.verbose:
        print("{:-^80}".format("Demo1 microservice system incident simulation"))

    env = os.environ
    env['GOOS'] = "linux"
    env['GOARCH'] = "amd64"
    env['CGO_ENABLED'] = "0"
    def build_go_binaries(target):
        run_cmds("go get; go build;cp {} {}".format(target, os.path.join(args.rootdir, f"build/package/{target}")),
                f"building {target}",
                cwd=os.path.join(args.rootdir, "cmd", target), env=env)


    # region Setup prometheus inside the swarm. Use docker-cli commands.
    if check_pipeline("prometheus"):
        target = "prometheus"
        print("{:.^80}".format(f"Build and deploy {target}"))
        # Build and push.
        run_cmds("docker build -t {tag} {dir};docker push {tag}".format(
                tag=args.registry+"/"+target,
                dir=os.path.join(args.rootdir, "deployments", "prometheus")), "Building "+target)
        # Remove any existing prometheus instances and deploy a new instance.
        run_cmds(
            """docker service rm {target}
            docker service create -p 9090:9090 --constraint node.role==manager \
                --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local \
                --mount type=volume,source=prometheus-data,target=/prometheus \
                --name={target} --replicas=1 --network={} {}/{target}
                """.format(args.network, args.registry, target=target), "deploying service "+target)
    # endregion


    # region Setup swarm-prometheus-discovery service.
    if check_pipeline("discovery"):
        target = "swarm-prometheus-discovery"
        print("{:.^80}".format(f"Build and deploy {target}"))
        if args.verbose:
            print(f"compiling {target}")
        build_go_binaries("swarm-prometheus-discovery")
        # build docker image
        if args.verbose:
            print(f"building docker image {target}")
        run_cmds("docker build -t {tag} {target};docker push {tag}".format(tag=args.registry+"/"+target,
                                                                        target=target),
                "Building and pushing "+target, cwd=os.path.join(args.rootdir, "build/package"))
        run_cmds(
            """docker service rm {target}
            docker service create \
                --constraint node.role==manager \
                --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local \
                --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
                --name={target} --replicas=1 --network={network} \
                --env NETWORK_NAME={network} \
                --env IGNORED_SERVICE_STR=prometheus,grafana,zipkin,{target},rabbitmq \
                --env INTERVAL=15 \
                {registry}/{target}
            """.format(target=target, network=args.network, registry=args.registry), "deploying service "+target)
    # endregion

    # region customservice
    if check_pipeline("customservice"):
        target = 'customservice'
        print("{:.^80}".format(f"Compile and package {target}"))
        # compile customservice
        if args.verbose:
            print("compiling customservice")
        build_go_binaries('customservice')  # compile customservice
        if args.verbose:
            print("compiling healthchecker")
        build_go_binaries('healthchecker')  # compile healthchecker
        # build docker image
        if args.verbose:
            print(f"building docker image {target}")
        run_cmds("docker build -t {tag} {target};docker push {tag}".format(
            target=target, tag=f"{args.registry}/{target}"),
            f"building docker image {target}",
            cwd=os.path.join(args.rootdir, "build/package"))
    # endregion

    # region Topology deploy
    if check_pipeline("topologydeploy"):
        target="topologydeploy"
        print("{:.^80}".format(f"Build and run {target}"))
        # Compile topologydeploy
        run_cmds("go get;go build;", f"compiling {target}", cwd=os.path.join(args.rootdir, "cmd", target))
        # Generate deploy scripts
        run_cmds(
            """./topologydeploy --useswarm=true --sshvolhost={} \
            --sshvolpath={}/ --sshvolpasswd={} --configfile={}\
            """.format(args.sshhost, os.path.join(args.rootdir, "deployments", "customtopology", "services"), args.sshpw,
                    os.path.join(args.rootdir, "deployments/customtopology/config.json")),
            f"Running {target}", cwd=os.path.join(args.rootdir, "cmd", target))
        

        # deploy services
        run_cmds("./scripts/deploy_customservices.sh", "deploying services")
    # endregion

    # region simple load test
    if check_pipeline("loadtest"):
        target = "loadtest"
        print("{:.^80}".format(f"Build and deploy {target}"))
        build_go_binaries(target)
        # Write Dockerfile
        with open(os.path.join(args.rootdir, "build", "package", target, "Dockerfile"), "wt") as f:
            # TODO: modify topology config.json to add a default gateway service name, i.e. entryservice.
            f.write("FROM iron/base\n" + 
                "ADD loadtest /\n" + 
                "ENTRYPOINT [\"./loadtest\"]\n" + 
                "CMD [\"-baseAddr\", \"servicea\", \"-port\", \"6767\", \"-zuul=false\", \"-users\", \"1\", \"-delay\", \"5000\"]")
        run_cmds("docker build -t {tag} {target};docker push {tag}".format(
            target=target, tag=f"{args.registry}/{target}"),
            f"building docker image {target}",
            cwd=os.path.join(args.rootdir, "build/package"))
        # Remove any existing loadtest instances and deploy a new instance.
        try:
            run_cmds(
                """docker service rm {target}
                docker service create --constraint node.role==manager \
                    --name={target} --replicas=1 --network={} {}/{target}
                    """.format(args.network, args.registry, target=target), "deploying service "+target)
        except KeyboardInterrupt as e:
            run_cmds(f"docker service rm {target}", "removing service "+target)
    # endregion

    sim_start = datetime.datetime.now()
    print("The Simulation Starts at:", sim_start.strftime("%Y%m%d-%H%M%S"))
    
    # region simple incident injection
    incident_duration = datetime.timedelta(seconds=350)
    if check_pipeline("injection"):
        # incident_duration = datetime.timedelta(minutes=10)
        print("{:.^80}".format("Simple incident injection (manually)"))
        while True:
            now = datetime.datetime.now()
            if now - sim_start < incident_duration:
                print("Sleeping...(Cur: {} Rem: {})".format(now.strftime("%Y%m%d-%H%M%S"), 
                                                            str(incident_duration - (now - sim_start))), 
                    end='\r')
                time.sleep(1)
            else:
                print("")
                print("Incident duration ends.")
                break
    # endregion

    # remove services
    # run_cmds("./scripts/remove_customservices.sh", "deploying services")

    # region Record simulation results
    if check_pipeline("record"):
        print("{:.^80}".format("Recording simulation results"))
        # TODO: move the following config name to the argparse
        svc_names = get_svc_names("deployments/customtopology/config.json")
        all_metric_data = query_prom_data_range(svc_names, get_query1, sim_start, sim_start+incident_duration * 2, 
            sampling_rate=1, is_summary=False, url="http://vmhost1.local:9090")
        metric_save_name = "{}_metric_data.pkl".format(args.expname)
        with open(metric_save_name, "wb") as f:
            pickle.dump(all_metric_data, f)
        with open(args.expname, "wt") as f:
            f.write("Simulation Start   : " + sim_start.strftime("%Y%m%d-%H%M%S") + "\n")
            f.write("Simulation Duration: " + str(incident_duration) + "\n")
            f.write("Metric Data        : " + metric_save_name + "\n")
    # endregion
