# Example

This is an example of how you could use `mani`.

`mani.yaml`
```yaml
projects:
  - name: example
    path: .

  - name: pinto
    path: frontend/pinto
    url: git@github.com:alajmo/pinto
    tags:
      - frontend

  - name: dashgrid
    path: frontend/dashgrid
    url: git@github.com:alajmo/dashgrid
    tags:
      - frontend

  - name: template-generator
    url: git@github.com:alajmo/template-generator
    tags:
      - bash

commands:
  - name: fetch
    command: git fetch

  - name: status
    command: git status

  - name: checkout
    args:
      branch: master
    command: git checkout $branch

  - name: create-branch
    command: git checkout -b $branch

  - name: multi
    command: | #Multi line command
      echo "2nd line "
      echo "3rd line"
```

Given the above `mani.yaml` we can run commands like:

```sh
# Get latest changes
$ mani run fetch --all-projects
```
