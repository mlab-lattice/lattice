# Setting up your local minikube with Google Open ID Connect (OIDC)

### Step 1 From the Google API console, create a project and a client id. 

Credit: https://medium.com/@jessgreb01/kubernetes-authn-authz-with-google-oidc-and-rbac-74509ca8267e


- From the dropdown, create a new project.
- Click ‘credentials’ from side nav bar,
- Select ‘OAuth consent screen’ and fill out form for your project and save
- Navigate back to ‘Credentials’ and click ‘Create credentials’, Select OAuth client ID
- Select ‘other’ application type and create the clientID and clientSecret. Store those somewhere safe for later steps.

### Step 2: start your minikube with oidc configs
```
$ minikube start \
      --extra-config=apiserver.oidc-issuer-url=https://accounts.google.com \
      --extra-config=apiserver.oidc-username-claim=email \
      --extra-config=apiserver.oidc-client-id="YOUR CLIENT ID"
```      

### Step 3: Create a role for your user in kubernetes

1- Create a yaml file using the following template provide your email 
```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-cluster-admins
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: oidc:<your email>
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: oidc:/cluster-admins
```

2- Seed the file to kubernetes

```
$ kubectl apply -f path-to-your-file
```

### Step 4: Install `k8s-oidc-helper` from abdulito's fork
This fork addresses this issue https://github.com/micahhausler/k8s-oidc-helper/issues/24

``` $ go get github.com/abdulito/k8s-oidc-helper```


### Step 5: Configure kubectl with using `k8s-oidc-helper` 

```$ k8s-oidc-helper --client-id=<client id> --client-secret=<client secret> --write=true```
This command will generate a token for you and save it back to kubectl's config. It will open a browser that will authenticate to google and give you a code.
Copy the code back and provide it to the tool and hit the return key. Your kubectl should have the user setup in config.


### Step 6: Test

Make sure that the auth is working

```
$ kubectl get nodes --user="your email"

NAME       STATUS    ROLES     AGE       VERSION
minikube   Ready     master    15h       v1.10.0

```




 
 
 