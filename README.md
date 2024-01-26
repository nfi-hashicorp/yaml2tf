# yaml2tf

WIP WIP WIP

`yaml2tf` converts arbitrary YAML into Terraform literals *without losing non-semantic information*, like *comments*, *string quoting style*, *map key order*, etc.

Motivation: I was working with `cloud-init` files in Terraform. They're YAML, and I needed to template them. The standard approach is to use `templatefile`, but then you're using a nice structured DSL to do string-templating on another structured DSL. You can end up with invalid YAML if you're not careful. `templatefile` isn't evaluated until run time (plan/apply), so you can't know you're missing a variable until then; your IDE won't save you.

With `yaml2tf`, you convert your existing YAML into Terraform, and then use standard Terraform constructs instead of `templatefile`'s pseudo-Terraform. You'll always have a real structure that you can just `yaml/jsonencode` when you go to consume it.

# vs [k2tf](https://github.com/sl1pm4t/k2tf)

`k2tf` only works for Kubernetes resources, AFAIK.

I'm not sure if it maintains non-semantic info, but probably not (it's hard!).

# vs `yamldecode`

You can do `echo 'yamldecode(file("my-manifest-file.yaml"))' | terraform console`, but it loses all non-semantic information.