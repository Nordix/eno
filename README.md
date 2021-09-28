# ENO
External Network Operator

## Deploy ENO using Cloud Infra Engine

Cloud Infra Engine framework ease the deployment of ENO together with all the required components in a fully configured K8s cluster.

More information regarding Cloud Infra Engine can be found [here](https://docs.nordix.org/submodules/infra/engine/docs/user-guide.html)

### Requirements

- Linux Distribution: Ubuntu 1804
- Minimum no of cores: 10
- Minimum RAM: 14G
- Minimum Disk Space: 150G
  
### Deploy ENO scenario

1. Clone Cloud Infra Engine repo

```bash
git clone https://gerrit.nordix.org/infra/engine.git
```
2. Move inside `engine` folder
```bash
cd engine/engine
```
3. Run ENO scenario
```bash
./deploy.sh \
-c k8-eno-ovs \
-v -p "https://gerrit.nordix.org/gitweb?p=infra/hwconfig.git;a=blob_plain;f=pods/eno-vpod1-pdf.yml" \
-i "https://gerrit.nordix.org/gitweb?p=infra/hwconfig.git;a=blob_plain;f=pods/eno-vpod1-idf.yml"
```

