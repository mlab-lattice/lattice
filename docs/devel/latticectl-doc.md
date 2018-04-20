# latticectl
## Introduction 
command line utility for interacting with lattices and systems
## Commands 
### bootstrap  
### bootstrap kubernetes  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--api-var API-VAR` **(required)** | configuration for the api | 
| `--cloud-provider CLOUD-PROVIDER` **(required)** | cloud provider that the kubernetes cluster is running on | 
| `--cloud-provider-var CLOUD-PROVIDER-VAR` **(required)** | configuration for the cloud provider lattice bootstrapper | 
| `--component-build-docker-artifact-var COMPONENT-BUILD-DOCKER-ARTIFACT-VAR` **(required)** | configuration for the docker artifacts produced by the component builder | 
| `--component-builder-var COMPONENT-BUILDER-VAR` **(required)** | configuration for the component builder | 
| `--controller-manager-var CONTROLLER-MANAGER-VAR` **(required)** | configuration for the controller manager | 
| `--service-mesh SERVICE-MESH` **(required)** | service mesh to bootstrap the lattice with | 
| `--service-mesh-var SERVICE-MESH-VAR` **(required)** | configuration for the service mesh cluster bootstrapper | 
| `--dry-run DRY-RUN` | if set, will not actually bootstrap the cluster. useful with --print | 
| `--kubeconfig KUBECONFIG` | path to kubeconfig | 
| `--lattice-id LATTICE-ID` | ID of the Lattice to bootstrap | 
| `--print PRINT` | whether or not to print the resources created or that will be created | 


### context  
### context get  
### context set  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--system SYSTEM` |  | 


### local  
### local down  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--name NAME` |  | 
| `--work-directory WORK-DIRECTORY` |  | 


### local up  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--container-channel CONTAINER-CHANNEL` |  | 
| `--container-registry CONTAINER-REGISTRY` |  | 
| `--name NAME` |  | 
| `--work-directory WORK-DIRECTORY` |  | 


### services  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### services addresses  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--service SERVICE` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 


### services status  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--service SERVICE` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems build  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--version VERSION` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems builds  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems builds status  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--build BUILD` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems create  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--definition DEFINITION` **(required)** |  | 
| `--name NAME` **(required)** |  | 
| `--lattice LATTICE` |  | 


### systems delete  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--system SYSTEM` **(required)** |  | 
| `--lattice LATTICE` |  | 


### systems deploy  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--build BUILD` |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--version VERSION` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems deploys  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems deploys status  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--deploy DEPLOY` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems secrets  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--system SYSTEM` |  | 


### systems secrets get  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--name NAME` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--system SYSTEM` |  | 


### systems secrets set  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--name NAME` **(required)** |  | 
| `--value VALUE` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--system SYSTEM` |  | 


### systems status  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems teardown  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems teardowns  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems teardowns status  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--teardown TEARDOWN` **(required)** |  | 
| `--lattice LATTICE` |  | 
| `--output OUTPUT`, `-o OUTPUT` |  | 
| `--system SYSTEM` |  | 
| `--watch WATCH`, `-w WATCH` |  | 


### systems versions  
**Flags**:  

| Name | Description | 
| --- | --- | 
| `--lattice LATTICE` |  | 
| `--system SYSTEM` |  | 


