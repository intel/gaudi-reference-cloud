# AuthZ service charts

## Additional configuration files

In the directory *`go/pkg/authz/default_data`*, we store additional configuration files for this service.

For any environment specific configuration files you should create a folder in the environment build folder. For example: In staging you want to create *`build/environments/staging/authz-files`*

Then you should overwrite the default values in the environment template. For example in staging you should modify *`deployment/helmfile/environments/staging.yaml.gotmpl`*:

```yaml
  authz:
    enabled: true
    groupsFile: ../../build/environments/staging/authz-files/groups.yaml
    modelConfFile: ../../build/environments/staging/authz-files/model.conf
    policiesPath: ../../build/environments/staging/authz-files/policies
    resourcesFile: ../../build/environments/staging/authz-files/resources.yaml
```

If the folder is *`default_data`* or *`authz-files`*, this is the folder and file structure you should follow:

```bash
default_data
├── groups.yaml             # YAML file with the group definitions for Casbin
├── model.conf              # Configuration for the Casbin model
├── policies                # Folder for the YAML files with the policy definitions for Casbin
│   ├── cloudaccount.yaml   # There should be a YAML file for each component
│   └── compute.yaml
└── resources.yaml          # YAML configuration file for the resources allowed actions
```

In the global helmfile template, there is a loop that reads each YAML file under the `policies` folder and add its content to a big **policy.csv** file, that loads all the policies as the pod starts.

The resulting **policy.csv** file and all other files in this folder will be mounted volumes in the pod as part of the deployment.

### Adding more config files

If you want to add another configuration file, you just need to modify the `deployment.yaml` and the `configmap.yaml` template.

For the `deployment.yaml`, you need to add a new mount volume under `spec.template.spec.containers.volumeMounts`:
```yaml
          - mountPath: /your_new_file.extension
            name: config
            subPath: your_new_file.extension
```
And, for the `configmap.yaml` add it under `data`:
```yaml
  your_new_file.extension: |
{{ .Values.yourNewFileContent | indent 4 }}
```
