# Example

This is an example of how you could use `mani`.

`mani.yaml`
```yaml
projects:
  - name: example
    path: .

  - name: idetheme
    path: frontend/idetheme
    url: https://github.com/samiralajmovic/idetheme
    tags:
      - frontend

  - name: dashgrid
    path: frontend/dashgrid
    url: https://github.com/samiralajmovic/dashgrid
    tags:
      - frontend

  - name: template-generator
    url: https://github.com/samiralajmovic/template-generator
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
